package main

import (
	"context"
	"fmt"
	"github.com/subiz/errors"
	"github.com/subiz/goutils/log"
	pb "github.com/subiz/header/partitioner"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"net"
	"time"
)

type Server struct {
	serverMap map[string]*server
}

type server struct {
	coor      *Coor
	cluster   string
	streamMgr *StreamMgr
	chans     *MultiChan
}

func daemon(ctx *cli.Context) error {
	db := NewDB(c.CassandraSeeds)
	bigServer := &Server{}
	bigServer.serverMap = make(map[string]*server)
	for _, service := range c.Services {
		s := &server{
			cluster:   service,
			streamMgr: NewStreamMgr(),
			chans:     NewMultiChan(),
			coor:      NewCoordinator(service, db, bigServer),
		}
		bigServer.serverMap[service] = s
		go lookupDNS(s)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCoordinatorServer(grpcServer, bigServer)

	lis, err := net.Listen("tcp", ":"+c.Port)
	if err != nil {
		return err
	}

	return grpcServer.Serve(lis)
}

type vote struct {
	term   int32
	accept bool
}

func makeChanId(workerid string, term int32) string {
	return fmt.Sprintf("%s|%d", workerid, term)
}

func (me *Server) Rebalance(wid *pb.WorkerID, stream pb.Coordinator_RebalanceServer) error {
	server := me.serverMap[wid.GetCluster()]
	if server == nil {
		return errors.New(400, errors.E_unknown, "cluster not found", wid.GetCluster())
	}
	server.streamMgr.Pull(wid.GetId(), stream, &pb.Configuration{TotalPartitions: -1})
	return nil
}

func (me *Server) Accept(ctx context.Context, wid *pb.WorkerID) (*pb.Empty, error) {
	server := me.serverMap[wid.GetCluster()]
	if server == nil {
		return nil, errors.New(400, errors.E_unknown, "cluster not found", wid.GetCluster())
	}

	chanid := makeChanId(wid.GetId(), wid.GetTerm())
	server.chans.Send(chanid, vote{term: wid.Term, accept: true}, 3*time.Second)
	return &pb.Empty{}, nil
}

func (me *Server) Deny(ctx context.Context, wid *pb.WorkerID) (*pb.Empty, error) {
	server := me.serverMap[wid.GetCluster()]
	if server == nil {
		return nil, errors.New(400, errors.E_unknown, "cluster not found", wid.GetCluster())
	}

	chanid := makeChanId(wid.GetId(), wid.GetTerm())
	server.chans.Send(chanid, vote{term: wid.Term, accept: false}, 3*time.Second)
	return &pb.Empty{}, nil
}

func (me *Server) GetConfig(ctx context.Context, cluster *pb.Cluster) (*pb.Configuration, error) {
	server := me.serverMap[cluster.GetId()]
	if server == nil {
		return nil, errors.New(400, errors.E_unknown, "cluster not found", cluster.GetId())
	}
	return server.coor.GetConfig(), nil
}

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
		msg, err := server.chans.Recv(chanid, 0)
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

func safe(f func()) {
	defer func() { recover() }()
	f()
}

func lookupDNS(s *server) {
	for {
		safe(func() {
			_, addrs, err := net.LookupSRV("", "", s.cluster)
			for _, record := range addrs {
				fmt.Printf("Target: %s:%d\n", record.Target, record.Port)
			}
			if err != nil {
				fmt.Printf("Could not get IPs: %v\n", err)
			}

			ips, err := net.LookupIP(s.cluster)
			if err != nil {
				fmt.Printf("Could not get IPs: %v\n", err)
				return
			}

			fmt.Println("looking up dns, got", ips)
			conf := s.coor.GetConfig()
			if len(ips) == len(conf.GetPartitions()) { // no change
				return
			}
			workers := make([]string, 0)
			for i := 0; i < len(ips); i++ {
				workers = append(workers, fmt.Sprintf("%s-%d", s.cluster, i))
			}
			if err := s.coor.ChangeWorkers(workers); err != nil {
				log.Error(err)
				return
			}
		})
		time.Sleep(2 * time.Second)
	}
}
