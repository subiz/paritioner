package client

import (
	"context"
	"hash/fnv"
	"math/rand"
	"strconv"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/metadata"
)

// parPicker is shorten for partition picker. Its used to pick the right host
// for a GRPC request based on partition key
type parPicker struct {
	partitions []string
	subConns   map[string]balancer.SubConn
}

// NewParPicker creates a new parPicker object
func NewParPicker(pars []string, subConns map[string]balancer.SubConn) *parPicker {
	me := &parPicker{}

	// copy subConns
	me.subConns = make(map[string]balancer.SubConn)
	for host, sc := range subConns {
		me.subConns[host] = sc
	}

	// copy partitions
	me.partitions = make([]string, len(pars))
	copy(me.partitions, pars)
	return me
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

	// hashing key to find the partition number
	g_pickerhash.Write([]byte(pkey))
	par := g_pickerhash.Sum32() % uint32(len(me.partitions))
	g_pickerhash.Reset()
	sc := me.subConns[me.partitions[par]]

	if sc == nil {
		return nil, nil, balancer.ErrNoSubConnAvailable
	}
	return sc, nil, nil
}

// global hashing util, used to hash key to partition number
// must lock before use
var g_pickerhash = fnv.New32a()

// GRPC Metadata key to store partition key attached to each GRPC request
// the key will be hashed to find the correct worker that handle the request
const PartitionKey = "partitionkey"
