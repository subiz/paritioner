package client

import (
	"context"
	"math/rand"
	"strconv"
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/metadata"
)

// parPicker is shorten for partition picker. Its used to pick the right host
// for a GRPC request based on partition key
type parPicker struct {
	*sync.Mutex
	partitions []string
	subConns   map[string]balancer.SubConn
}

// NewParPicker creates a new parPicker object
func NewParPicker() *parPicker { return &parPicker{Mutex: &sync.Mutex{}} }

// Update is called by balancer to updates picker's current partitions and subConns
func (me *parPicker) Update(pars []string, subConns map[string]balancer.SubConn) {
	me.Lock()
	defer me.Unlock()

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
	mdpar := md[PartitionKey]
	var pkey string
	if len(mdpar) > 0 {
		pkey = mdpar[0]
	}

	if pkey == "" { // random host
		pkey = strconv.Itoa(rand.Int())
	}

	me.Lock()
	// hashing key to find the partition number
	ghash.Write([]byte(pkey))
	par := ghash.Sum32() % uint32(len(me.partitions))
	ghash.Reset()
	sc := me.subConns[me.partitions[par]]
	me.Unlock()

	if sc == nil {
		return nil, nil, balancer.ErrNoSubConnAvailable
	}
	return sc, nil, nil
}
