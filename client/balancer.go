// Package client defines a roundrobin balancer. Roundrobin balancer is
// installed as one of the default balancers in gRPC, users don't need to
// explicitly install this balancer.

package client

import (
	"fmt"
	"strings"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
)

// Name is the name of partitioner balancer.
const Name = "partitioner"

func init() {
	balancer.Register(&parBuilder{name: Name, pickerBuilder: &pPickerBuilder{}})
}

// partitioner picker builder
type pPickerBuilder struct{}

func (*pPickerBuilder) Build(pars []string, readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	fmt.Printf("partitionerPicker: newPicker called with readySCs: %v\n", readySCs)
	var scs []balancer.SubConn

	subConnHost := make(map[string]balancer.SubConn)
	for a, sc := range readySCs {
		scs = append(scs, sc)
		subConnHost[a.Addr] = sc
	}
	return &parPicker{subConns: scs, partitions: pars, subConnHost: subConnHost}
}

type parPicker struct {
	*sync.Mutex
	// subConns is the snapshot of the roundrobin balancer when this picker was
	// created. The slice is immutable. Each Get() will do a round robin
	// selection from it and return the selected SubConn.
	subConns []balancer.SubConn

	next int

	partitions  []string
	subConnHost map[string]balancer.SubConn
}

func (me *parPicker) Pick(ctx context.Context, opts balancer.PickOptions) (balancer.SubConn, func(balancer.DoneInfo), error) {
	if len(p.subConns) <= 0 {
		return nil, nil, balancer.ErrNoSubConnAvailable
	}

	md, _ := metadata.FromOutgoingContext(ctx)
	pkey := strings.Join(md[PartitionKey], "")

	if pkey == "" { // round robin
		me.Lock()
		sc := me.subConns[me.next]
		me.next = (me.next + 1) % len(me.subConns)
		me.Unlock()
		return sc, nil, nil
	}

	// hashing key to find the partition number
	ghash.Write([]byte(pkey))
	par := ghash.Sum32() % uint32(len(me.partitions))
	ghash.Reset()

	me.Lock()
	host := me.partitions[par]
	sc := me.subConnHost[host]
	me.Unlock()
	if sc == nil {
		return nil, nil, balancer.ErrNoSubConnAvailable
	}
	return sc, nil, nil
}
