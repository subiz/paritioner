package client

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/subiz/errors"
	pb "github.com/subiz/partitioner/header"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
)

// GRPC balancer requires 3 parts: Builer, Balancer and Picker
// GRPC uses Builder to create a new Balancer object
//
// Balancer calls Picker.Pick before every request to determinds which host
// to send the request to
//
//
//
//

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

// parBalancer is a GRPC balancer, its name is shorten for partitioner balancer
type parBalancer struct {
	*sync.Mutex
	cc balancer.ClientConn

	// subConn for each host
	subConns map[string]balancer.SubConn

	picker balancer.Picker

	// partition map
	partitions []string

	// use to make sure we only running fetching loop once
	fetching bool
}

// fetchLoop synchronizes partition in every 30 sec
//
func (me *parBalancer) fetchLoop(addr string) {
	// make sure that we only loop once
	// the first loop will set fetching=true
	me.Lock()
	if me.fetching {
		me.Unlock()
		return
	}
	me.fetching = true
	me.Unlock()

	// try to connect to a worker, retry automatically on error
	// block until success
	var pconn *grpc.ClientConn
	for {
		var err error
		pconn, err = dialGrpc(addr)
		if err != nil {
			fmt.Printf("ERR LKJSDLFKJ49FD cannot connect to addr %s, %v\n", addr, err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	pclient := pb.NewWorkerClient(pconn)

	// main loop
	for {
		pars, err := fetchPartitions(pclient)
		if err != nil {
			fmt.Printf("ERR#FS94GPOFD fetching partition: %v\n", err)
		}

		// making array of hosts using partition map
		hostM := make(map[string]bool, 0)
		for _, host := range pars {
			hostM[host] = true
		}
		hosts := make([]string, 0)
		for host := range hostM {
			hosts = append(hosts, host)
		}

		addrsSet := make(map[string]struct{})
		for _, host := range hosts {
			addrsSet[host] = struct{}{}
			if _, ok := me.subConns[host]; !ok {
				sc, err := me.cc.NewSubConn([]resolver.Address{{
					Addr: host,
					Type: resolver.Backend,
				}}, balancer.NewSubConnOptions{HealthCheckEnabled: false})
				if err != nil {
					fmt.Printf("ERR HSJKH849FDSF %v", err)
					continue
				}
				me.subConns[host] = sc
				sc.Connect()
			}
		}

		for h, sc := range me.subConns {
			// host h was removed by resolver.
			if _, ok := addrsSet[h]; !ok {
				me.cc.RemoveSubConn(sc)
				delete(me.subConns, h)
			}
		}

		me.Lock()
		me.partitions = pars
		me.Unlock()

		time.Sleep(30 * time.Second)
	}
}

// ignore since we dont rely in resolver addrs
func (me *parBalancer) HandleResolvedAddrs(addrs []resolver.Address, err error) {
	if err != nil {
		fmt.Printf("ERR base.baseBalancer: HandleResolvedAddrs called with error %v", err)
		return
	}

	// use first address only, the rest of cluster will be discover through
	// this address
	me.fetchLoop(addrs[0].Addr)
}

// also ignore since we have our own implementation
func (me *parBalancer) HandleSubConnStateChange(sc balancer.SubConn, s connectivity.State) {
	fmt.Printf("partitioner.parBalancer: handle SubConn state change: %p, %v\n", sc, s)
	if s == connectivity.Idle {
		sc.Connect()
	}
	me.cc.UpdateBalancerState(connectivity.Ready, me.picker)
}

// Close is a nop because base balancer doesn't have internal state to clean up,
// and it doesn't need to call RemoveSubConn for the SubConns.
func (_ *parBalancer) Close() {}

// fetchPartitions calls worker api to return the latest partition map
// Partition map is an array tells which worker will handle which partition
// partition number is ordinal index of array, each element contains worker
// host
// eg: ["worker-0:8081", "worker-1:8081", "worker-0:8081"] has 3 partitions:
// {0, 1, 2}, partition 0 and 2 are handled by worker 0 at host worker-0:8081
// partition 1 is handled by worker 1 at host worker-1:8081
func fetchPartitions(pclient pb.WorkerClient) (pars []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			if err, ok = r.(error); ok {
				err = errors.Wrap(err, 500, errors.E_unknown)
				return
			}
			err = errors.New(500, errors.E_unknown, r)
		}
	}()

	conf, err := pclient.GetConfig(context.Background(), &pb.GetConfigRequest{})
	if err != nil {
		return nil, err
	}

	// convert configuration fetched from worker to partition map
	partitions := make([]string, conf.GetTotalPartitions())
	for _, workerinfo := range conf.Workers {
		for _, parNum := range workerinfo.GetPartitions() {
			partitions[parNum] = workerinfo.Host
		}
	}
	return partitions, nil
}

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
	g_pickerhash.Write([]byte(pkey))
	par := g_pickerhash.Sum32() % uint32(len(me.partitions))
	g_pickerhash.Reset()
	sc := me.subConns[me.partitions[par]]
	me.Unlock()

	if sc == nil {
		return nil, nil, balancer.ErrNoSubConnAvailable
	}
	return sc, nil, nil
}

// global hashing util, used to hash key to partition number
// must lock before use
var g_pickerhash = fnv.New32a()
