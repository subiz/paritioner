package main

import (
	"context"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/subiz/errors"
	pb "github.com/subiz/header/partitioner"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"net"
)

type BigServer struct {
	serverMap map[string]*Server
}

type Config struct {
	CassandraSeeds []string `required:"true"`
	Port           string   `required:"true"`
	Services       []string `required:"true"`
}

var c Config

func main() {
	envconfig.MustProcess("coor", &c)
	app := cli.NewApp()
	app.Name = "coordinator"

	app.Action = func(c *cli.Context) error {
		fmt.Println("hello from partition coordinator")
		return nil
	}

	app.Commands = []cli.Command{
		{Name: "daemon", Usage: "run server", Action: daemon},
	}
	app.RunAndExitOnError()
}

func daemon(ctx *cli.Context) error {
	db := NewDB(c.CassandraSeeds)
	bigServer := &BigServer{}
	bigServer.serverMap = make(map[string]*Server)
	for _, service := range c.Services {
		bigServer.serverMap[service] = NewServer(service, db)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCoordinatorServer(grpcServer, bigServer)

	lis, err := net.Listen("tcp", ":"+c.Port)
	if err != nil {
		return err
	}

	return grpcServer.Serve(lis)
}

func (me *BigServer) Rebalance(wid *pb.WorkerID, stream pb.Coordinator_RebalanceServer) error {
	server := me.serverMap[wid.GetCluster()]
	if server == nil {
		return errors.New(400, errors.E_unknown, "cluster not found", wid.GetCluster())
	}
	server.Pull(wid.GetId(), stream)
	return nil
}

func (me *BigServer) Accept(ctx context.Context, wid *pb.WorkerID) (*pb.Empty, error) {
	server := me.serverMap[wid.GetCluster()]
	if server == nil {
		return nil, errors.New(400, errors.E_unknown, "cluster not found", wid.GetCluster())
	}
	server.Vote(wid.GetId(), wid.GetTerm(), true)
	return &pb.Empty{}, nil
}

func (me *BigServer) Deny(ctx context.Context, wid *pb.WorkerID) (*pb.Empty, error) {
	server := me.serverMap[wid.GetCluster()]
	if server == nil {
		return nil, errors.New(400, errors.E_unknown, "cluster not found", wid.GetCluster())
	}
	server.Vote(wid.GetId(), wid.GetTerm(), false)
	return &pb.Empty{}, nil
}

func (me *BigServer) GetConfig(ctx context.Context, cluster *pb.Cluster) (*pb.Configuration, error) {
	server := me.serverMap[cluster.GetId()]
	if server == nil {
		return nil, errors.New(400, errors.E_unknown, "cluster not found", cluster.GetId())
	}
	return server.GetConfig(), nil
}
