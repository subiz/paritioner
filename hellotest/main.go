package main

import (
	"context"
	"fmt"
	pb "github.com/subiz/header/partitioner"
	"github.com/subiz/partitioner/client"
	"github.com/subiz/partitioner/worker"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"time"
)

const (
	host = ":50051"
)

// server is used to implement partitioner.HelloServer.
type server struct{}

// Hello implements partitioner.HelloServer
func (s *server) Hello(ctx context.Context, in *pb.String) (*pb.String, error) {
	log.Printf("Received: %v", in.Str)
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return &pb.String{Str: "Hello " + in.Str + "from" + hostname}, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "hello test"

	app.Action = func(c *cli.Context) error {
		fmt.Println("hello from hello test")
		return nil
	}

	app.Commands = []cli.Command{
		{Name: "server", Usage: "run server", Action: runServer},
		{Name: "client", Usage: "run client", Action: runClient},
	}
	app.RunAndExitOnError()
}

func runServer(ctx *cli.Context) {
	lis, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	println("hostname", hostname)

	w := worker.NewWorker(host, "hellotest", hostname, "coordinator:8021")
	interceptor := grpc.UnaryInterceptor(w.CreateIntercept(pb.NewHelloClient(nil)))
	grpcserver := grpc.NewServer(interceptor)
	pb.RegisterWorkerServer(grpcserver, w)
	pb.RegisterHelloServer(grpcserver, &server{})
	reflection.Register(grpcserver)
	if err := grpcserver.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func runClient(ctx *cli.Context) {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(10*time.Second))
	opts = append(opts, client.NewClient("localhost:50051").WithInterceptor())
	conn, err := grpc.Dial("localhost:50051", opts...)
	if err != nil {
		panic(err)
	}

	c := pb.NewHelloClient(conn)
	s, err := c.Hello(context.Background(), &pb.String{Str: "haivan"})
	if err != nil {
		panic(err)
	}

	fmt.Println("ret", s.Str)
}
