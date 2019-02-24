package client

import (
	"math/rand"
	"time"

	"google.golang.org/grpc/balancer"
)

// Name is the name of partitioner balancer.
const Name = "partitioner"

func init() {
	rand.Seed(time.Now().UnixNano())
	balancer.Register(&parBuilder{})
}
