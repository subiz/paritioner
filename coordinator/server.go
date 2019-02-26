package main

import (
	"context"
	"fmt"
	"github.com/subiz/errors"
	pb "github.com/subiz/partitioner/header"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"net"
	"time"
)

// Server acts as a gateway (proxy) so multiple coordinator can run within
// in a single GRPC server
// Server uses the `cluster` parameter in each request determinds which
// coordinator should handle the request
type Server struct {
	// map of server by cluster_id, this map is write one when creating server
	// so we don't need a lock for this map
	serverMap map[string]*server
}

// server is a GRPC server
// each coordinator cluster should creates
type server struct {
	coor      *Coor
	cluster   string
	streamMgr *StreamMgr
	chans     *MultiChan
}

// daemon loads all clusters and start GRPC server
func daemon(ctx *cli.Context) error {
	db := NewDB(cf.CassandraSeeds)
	bigServer := &Server{}
	bigServer.serverMap = make(map[string]*server)
	var err error
	for _, service := range cf.Services {
		s := &server{
			cluster:   service,
			streamMgr: NewStreamMgr(),
			chans:     NewMultiChan(),
			coor:      NewCoordinator(service, db, bigServer),
		}
		bigServer.serverMap[service] = s
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCoordinatorServer(grpcServer, bigServer)

	lis, err := net.Listen("tcp", ":"+cf.Port)
	if err != nil {
		return err
	}

	return grpcServer.Serve(lis)
}

type vote struct {
	term   int32
	accept bool
}

// makeChanId returns composited channel id, which unique for each worker
// and its term
func makeChanId(workerid string, term int32) string {
	return fmt.Sprintf("%s|%d", workerid, term)
}

// Leave is called by a worker to tell coordinator that its no longer a
// member of the cluster.
func (me *Server) Leave(ctx context.Context, p *pb.WorkerRequest) (*pb.Empty, error) {
	server := me.serverMap[p.GetCluster()]
	if server == nil {
		return nil, errors.New(400, errors.E_unknown, "cluster not found", p.GetCluster())
	}

	if err := server.coor.Leave(p); err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

// Join is called by a worker to tell coordinator that it want to be a
// member of the cluster.
func (me *Server) Join(ctx context.Context, p *pb.WorkerRequest) (*pb.Empty, error) {
	server := me.serverMap[p.GetCluster()]
	if server == nil {
		return nil, errors.New(400, errors.E_unknown, "cluster not found", p.GetCluster())
	}

	if err := server.coor.Join(p); err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func (me *Server) Rebalance(w *pb.WorkerRequest, stream pb.Coordinator_RebalanceServer) error {
	server := me.serverMap[w.GetCluster()]
	if server == nil {
		return errors.New(400, errors.E_unknown, "cluster not found", w.GetCluster())
	}

	if err := server.coor.Join(w); err != nil {
		return err
	}

	server.streamMgr.Pull(w.GetId(), stream)
	return nil
}

// Accept used by a worker to tell coordinator that it has received an update
// signal and ready for applying the change
func (me *Server) Accept(ctx context.Context, w *pb.WorkerRequest) (*pb.Empty, error) {
	server := me.serverMap[w.GetCluster()]
	if server == nil {
		return nil, errors.New(400, errors.E_unknown, "cluster not found", w.GetCluster())
	}

	chanid := makeChanId(w.GetId(), w.GetTerm())
	server.chans.Send(chanid, vote{term: w.Term, accept: true}, 3*time.Second)
	return &pb.Empty{}, nil
}

// Deny used by a worker to deny prepare request from a coordinator
// This function is just a concept, not fully implemented yet
func (me *Server) Deny(ctx context.Context, w *pb.WorkerRequest) (*pb.Empty, error) {
	server := me.serverMap[w.GetCluster()]
	if server == nil {
		return nil, errors.New(400, errors.E_unknown, "cluster not found", w.GetCluster())
	}

	chanid := makeChanId(w.GetId(), w.GetTerm())
	server.chans.Send(chanid, vote{term: w.Term, accept: false}, 3*time.Second)
	return &pb.Empty{}, nil
}

// GetConfig returns the current configuration of a coordinator
// This function block while the coordinator in middle of a transition
func (me *Server) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.Configuration, error) {
	server := me.serverMap[req.Cluster]
	if server == nil {
		return nil, errors.New(400, errors.E_unknown, "cluster not found", req.Cluster)
	}
	return server.coor.GetConfig(), nil
}

// Prepare is used by coordinator to send updates to its workers
// This function blocks until the worker reply or timeout
func (me *Server) Prepare(cluster, workerid string, conf *pb.Configuration) error {
	server := me.serverMap[cluster]
	if server == nil {
		return errors.New(400, errors.E_unknown, "cluster not found", cluster)
	}

	if err := server.streamMgr.Send(workerid, conf); err != nil {
		return err
	}
	chanid := makeChanId(workerid, conf.GetTerm())

	for {
		msg, err := server.chans.Recv(chanid, 30*time.Second)
		if err != nil {
			return err
		}
		vote := msg.(vote)
		if vote.term == conf.Term { // ignore outdated answer
			continue
		}

		if vote.accept {
			return nil
		}
		return errors.New(500, errors.E_partition_rebalance_timeout,
			"worker donot accept")
	}
}
