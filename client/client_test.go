package client

import (
	"time"
	"context"
	pb "github.com/subiz/partitioner/header"
	"google.golang.org/grpc"
	"testing"
)

func createClient(workerhost string) pb.HelloClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	opts = append(opts, grpc.WithTimeout(10*time.Second))
	opts = append(opts, grpc.WithBalancerName(Name))
	conn, err := grpc.Dial(workerhost, opts...)
	if err != nil {
		panic(err)
	}

	return pb.NewHelloClient(conn)
}

func TestBalancer(t *testing.T) {
	conf := &pb.Configuration{
		Cluster:         "cluster",
		Term:            2,
		TotalPartitions: 10,
		Workers: map[string]*pb.WorkerInfo{
			"1": &pb.WorkerInfo{
				Id:         "1",
				Partitions: []int32{0, 1, 2, 3, 4},
			},
			"2": &pb.WorkerInfo{
				Id:         "2",
				Partitions: []int32{5, 6, 7, 8, 9},
			},
		},
	}
	go RunWorker("10001", "1", conf)
	go RunWorker("10002", "2", conf)

	client := createClient("localhost:10001")
	out, err := client.Hello(context.Background(), &pb.GetConfigRequest{})
	if err != nil {
		t.Fatal(err)
	}

	println("OUT", out.Id)
}
