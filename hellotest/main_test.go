package main

import (
	"context"
	"strconv"
	"fmt"
	"github.com/subiz/partitioner/client"
	coor "github.com/subiz/partitioner/coordinator"
	pb "github.com/subiz/partitioner/header"
	"github.com/subiz/partitioner/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"testing"
	"time"
)

const (
	host = ":50051"
)

// server is used to implement partitioner.HelloServer.
type server struct {
	name string
}

// Hello implements partitioner.HelloServer
func (s *server) Hello(ctx context.Context, in *pb.GetConfigRequest) (*pb.WorkerInfo, error) {
	log.Printf("Received: %v", in.Cluster)
	return &pb.WorkerInfo{Host: "Hello " + in.Cluster + "from" + s.name}, nil
}

func TestWorker(t *testing.T) {
	var cluster = "cluster1" + strconv.FormatInt(time.Now().UnixNano(), 10)
	go coor.RunCoordinator([]string{"172.17.0.2:9042"}, []string{cluster},
		"10000", "cassandra", "cassandra")

	go runWorker("10001", "worker1", cluster, "localhost:10000")
	go runWorker("10002", "worker2", cluster, "localhost:10000")

	time.Sleep(10 * time.Second)
	println("CREATING CLIENT")
	c := createClient("localhost:10001")
	key := "123"
	ct := metadata.AppendToOutgoingContext(context.Background(),
		"partitionkey", key)

	s, err := c.Hello(ct, &pb.GetConfigRequest{Cluster: "thanh"})
	if err != nil {
		panic(err)
	}

	fmt.Println("ret", s.Host)
	// send
}

func runWorker(port, id, cluster, coordinatorhost string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}

	w := worker.NewWorker(":"+port, cluster, id, coordinatorhost)

	interceptor := grpc.UnaryInterceptor(w.CreateIntercept(pb.NewHelloClient(nil)))
	grpcserver := grpc.NewServer(interceptor)
	pb.RegisterWorkerServer(grpcserver, w)
	pb.RegisterHelloServer(grpcserver, &server{})

	reflection.Register(grpcserver)
	println("WORKERURNINNG ART" ,port)
	if err := grpcserver.Serve(lis); err != nil {
		panic(err)
	}
}

func createClient(workerhost string) pb.HelloClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(10*time.Second))
	opts = append(opts, grpc.WithBalancerName(client.Name))
	conn, err := grpc.Dial(workerhost, opts...)
	if err != nil {
		panic(err)
	}

	return pb.NewHelloClient(conn)
}
