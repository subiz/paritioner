package client

import (
	"context"
	pb "github.com/subiz/partitioner/header"
	"google.golang.org/grpc"
	"testing"
	"time"
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
				Host:       "localhost:10001",
				Id:         "1",
				Partitions: []int32{0, 1, 2, 3, 4},
			},
			"2": &pb.WorkerInfo{
				Id:         "2",
				Partitions: []int32{5, 6, 7, 8, 9},
				Host:       "localhost:10002",
			},
		},
	}
	go RunWorker("10001", "1", conf)
	go RunWorker("10002", "2", conf)

	client := createClient("localhost:10001")
	w1c, w2c := 0, 0
	for i := 0; i < 10000; i++ {
		out, err := client.Hello(context.Background(), &pb.GetConfigRequest{})
		if err != nil {
			t.Fatal(err)
		}
		if out.Id == "1" {
			w1c++
		} else if out.Id == "2" {
			w2c++
		}
	}

	println("OUT", w1c, w2c)
}
