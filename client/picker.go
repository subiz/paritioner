package client

import (
	"context"
	"fmt"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/metadata"
	"math/rand"
	"strings"
	"sync"
)

func NewParPicker() *parPicker { return &parPicker{Mutex: &sync.Mutex{}} }

// parPicker is shorten for partition picker. Its being used to pick the host
// for a GRPC request based on partition key
type parPicker struct {
	*sync.Mutex
	partitions []string
	subConns   map[string]balancer.SubConn
}

// Update is called by balancer to updates picker's current partitions and subConns
func (me *parPicker) Update(pars []string, subConns map[string]balancer.SubConn) {
	me.Lock()
	defer me.Unlock()
	fmt.Printf("partitionerPicker: newPicker called with readySCs: %v\n", pars)

	// copy subConns
	me.subConns = make(map[string]balancer.SubConn)
	for host, sc := range subConns {
		me.subConns[host] = sc
	}

	// copy partitions
	me.partitions = make([]string, len(pars))
	copy(me.partitions, pars)
}

// Pick is called before each GRPC request to determind wich subconn to send
// the request to
// We find worker host by hashing partition key to a partition number then
// look it up in the current partition map
// worker host might not be the correct one since partition map can be
// unsynchronized (by network lantences or network partitioned). When it
// happend, request will be redirected one or more times in worker side till
// it arrive the correct host. This may look ineffiecient but we don't have to
// guarrantee consistent at client-side, make client-side code so much simpler
// when the  network is stable and no worker member changes (most of the time),
// partition map is in-sync, clients are now sending requests directly to the
// correct host without making any additional redirection.
func (me *parPicker) Pick(ctx context.Context, opts balancer.PickOptions) (balancer.SubConn, func(balancer.DoneInfo), error) {
	if len(me.subConns) <= 0 {
		return nil, nil, balancer.ErrNoSubConnAvailable
	}

	md, _ := metadata.FromOutgoingContext(ctx)
	pkey := strings.Join(md[PartitionKey], "")

	if pkey == "" { // random host
		pkey = fmt.Sprintf("%d", rand.Int())
	}

	// hashing key to find the partition number
	ghash.Write([]byte(pkey))
	par := ghash.Sum32() % uint32(len(me.partitions))
	ghash.Reset()

	me.Lock()
	sc := me.subConns[me.partitions[par]]
	me.Unlock()
	if sc == nil {
		return nil, nil, balancer.ErrNoSubConnAvailable
	}
	return sc, nil, nil
}
