package client

import (
	"context"
	pb "github.com/subiz/partitioner/header"
	"google.golang.org/grpc"
	"net"
)

type sampleWorker struct {
	id     string
	config *pb.Configuration
}

func (s *sampleWorker) Notify(_ context.Context, _ *pb.Configuration) (*pb.Empty, error) {
	return nil, nil
}

func (s *sampleWorker) GetConfig(_ context.Context, _ *pb.GetConfigRequest) (*pb.Configuration, error) {
	println("CALLED getconfig")
	return s.config, nil
}

func (s *sampleWorker) Hello(_ context.Context, _ *pb.GetConfigRequest) (*pb.WorkerInfo, error) {
	println("CALLED hello")
	return &pb.WorkerInfo{Id: s.id}, nil
}

func RunWorker(port, id string, config *pb.Configuration) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}

	w := &sampleWorker{config: config}
	grpcserver := grpc.NewServer()
	pb.RegisterWorkerServer(grpcserver, w)
	pb.RegisterHelloServer(grpcserver, w)
	if err := grpcserver.Serve(lis); err != nil {
		panic(err)
	}
}
