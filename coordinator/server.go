package coordinator

import (
	"context"
	"fmt"
	"github.com/subiz/errors"
	pb "github.com/subiz/partitioner/header"
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
	coor    *Coor
	cluster string
}

func RunCoordinator(seeds, clusters []string, port, us, pw string) {
	db := NewDB(seeds, us, pw)
	bigServer := &Server{}
	bigServer.serverMap = make(map[string]*server)
	var err error
	for _, cluster := range clusters {
		coordinator := NewCoordinator(cluster, db, bigServer)
		s := &server{cluster: cluster, coor: coordinator}
		bigServer.serverMap[cluster] = s
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCoordinatorServer(grpcServer, bigServer)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}

	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}

// makeChanId returns composited channel id, which unique for each worker
// and its term
func makeChanId(workerid string, term int32) string {
	return fmt.Sprintf("%s|%d", workerid, term)
}

// Leave is called by a worker to tell coordinator that its no longer a
// member of the cluster.
func (me *Server) Leave(ctx context.Context, p *pb.LeaveRequest) (*pb.Empty, error) {
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
func (me *Server) Join(ctx context.Context, p *pb.JoinRequest) (*pb.Configuration, error) {
	server := me.serverMap[p.GetCluster()]
	if server == nil {
		return nil, errors.New(400, errors.E_unknown, "cluster not found", p.GetCluster())
	}

	if err := server.coor.Join(p); err != nil {
		return nil, err
	}

	return server.coor.GetConfig(), nil
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
// TODO: should cache dialed client
func (me *Server) Prepare(host string, conf *pb.Configuration) error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(5*time.Second))

	var err error
	var cc *grpc.ClientConn
	for i := 0; i < 10; i++ {
		if cc, err = grpc.Dial(host, opts...); err == nil {
			break
		}
	}
	if err != nil {
		return err
	}
	client := pb.NewWorkerClient(cc)
	_, err = client.Notify(context.Background(), conf)
	return err
}
