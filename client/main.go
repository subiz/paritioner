package client

import (
	"google.golang.org/grpc/balancer"
	"math/rand"
	"sync"
	"time"
)

// Name is the name of partitioner balancer.
const Name = "partitioner"

func init() {
	rand.Seed(time.Now().UnixNano())
	balancer.Register(&parBuilder{})
}

// parBuilder is shorten for partition load balancer builder. Its used by GRPC
// to build a new balancer, which is in this case, a ParBalancer
// This struct implements GRPC interface balancer.Builder
type parBuilder struct{}

// Build implements GRPC method Build in balancer.Builder
func (me *parBuilder) Build(cc balancer.ClientConn,
	opt balancer.BuildOptions) balancer.Balancer {
	return &parBalancer{Mutex: &sync.Mutex{}, cc: cc, picker: NewParPicker()}
}

// Name implements GRPC method Name in balancer.Builder
func (me *parBuilder) Name() string { return Name }
