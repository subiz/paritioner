package main

import (
	"context"
	pb "github.com/subiz/partitioner/header"
	"google.golang.org/grpc"
	"io"
	"net"
	"testing"
	"time"
)

func TestRebalance(t *testing.T) {
	// start server
	db := NewDBMock()
	bigServer := &Server{}
	bigServer.serverMap = make(map[string]*server)
	var err error
	for _, service := range []string{"cluster1"} {
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

	lis, err := net.Listen("tcp", "127.0.0.1:12455")
	if err != nil {
		t.Fatal(err)
	}

	go grpcServer.Serve(lis)
	time.Sleep(1 * time.Second)
	confchan := make(chan *pb.Configuration, 0)
	go goodWorker("127.0.0.1:12455", "cluster1", "worker1", confchan)
	time.Sleep(1 * time.Second)
	go badWorker("127.0.0.1:12455", "cluster1", "worker2")

	for conf := range confchan {
		println(conf.Cluster, conf.Term, len(conf.Workers))
	}
}

func badWorker(host, cluster, workerid string) {
	// first join then leave
	client, err := dialCoordinator(host)
	if err != nil {
		panic(err)
	}
	stream, err := client.Rebalance(context.Background(), &pb.WorkerRequest{
		Version: VERSION,
		Cluster: cluster,
		Id:      workerid,
		Host:    workerid + ":1234",
		Term:    0,
	})

	if err != nil {
		panic(err)
	}

	conf, err := stream.Recv()
	if err != nil {
		panic(err)
	}
	_, err = client.Accept(context.Background(), &pb.WorkerRequest{
		Version: VERSION,
		Cluster: cluster,
		Id:      workerid,
		Term:    conf.Term,
	})
	if err != nil {
		panic(err)
	}

	client.Leave(context.Background(), &pb.WorkerRequest{
		Version: VERSION,
		Cluster: cluster,
		Id:      workerid,
		Term:    conf.Term,
	})
}

func goodWorker(host, cluster, workerid string, confchan chan *pb.Configuration) {
	client, err := dialCoordinator(host)
	if err != nil {
		panic(err)
	}
	stream, err := client.Rebalance(context.Background(), &pb.WorkerRequest{
		Version: VERSION,
		Cluster: cluster,
		Id:      workerid,
		Host:    workerid + ":1234",
		Term:    0,
	})

	if err != nil {
		panic(err)
	}

	for {
		conf, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			panic(err)
		}
		confchan <- conf
		_, err = client.Accept(context.Background(), &pb.WorkerRequest{
			Version: VERSION,
			Cluster: "cluster1",
			Id:      "worker1",
			Term:    conf.Term,
		})
		if err != nil {
			panic(err)
		}
	}
}

// dialGrpc makes a GRPC connection to service
func dialCoordinator(service string) (pb.CoordinatorClient, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(5*time.Second))

	cc, err := grpc.Dial(service, opts...)
	if err != nil {
		return nil, err
	}

	return pb.NewCoordinatorClient(cc), nil
}
