package client

import (
	"context"
	"fmt"
	"github.com/subiz/errors"
	pb "github.com/subiz/header/partitioner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Name is the name of partitioner balancer.
const Name = "partitioner"

func init() {
	rand.Seed(time.Now().UnixNano())
	balancer.Register(&parBuilder{name: Name})
}

func NewParPicker() *parPicker { return &parPicker{Mutex: &sync.Mutex{}} }

type parPicker struct {
	*sync.Mutex
	partitions []string
	subConns   map[string]balancer.SubConn
}

func (me *parPicker) Update(pars []string, subConns map[string]balancer.SubConn) {
	me.Lock()
	defer me.Unlock()
	fmt.Printf("partitionerPicker: newPicker called with readySCs: %v\n", pars)

	// copy subConns
	me.subConns = make(map[string]balancer.SubConn)
	for host, sc := range subConns {
		me.subConns[host] = sc
	}

	me.partitions = pars
}

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

type parBuilder struct {
	name string
}

func (me *parBuilder) Build(cc balancer.ClientConn, opt balancer.BuildOptions) balancer.Balancer {
	return &parBalancer{
		Mutex:    &sync.Mutex{},
		cc:       cc,
		subConns: make(map[string]balancer.SubConn),
		picker:   NewParPicker(),
	}
}

func (me *parBuilder) Name() string { return me.name }

// partitioner balancer
type parBalancer struct {
	*sync.Mutex
	cc balancer.ClientConn

	subConns map[string]balancer.SubConn

	picker     balancer.Picker
	conn       map[string]*grpc.ClientConn
	partitions []string

	// tells that fetchLoop is running or not
	fetching bool
}

// fetchLoop synchronizes partition in every 30 sec
func (me *parBalancer) fetchLoop(service string) {
	if me.fetching {
		return
	}

	me.fetching = true
	pconn, err := dialGrpc(service)
	if err != nil {
		panic(err)
	}

	pclient := pb.NewWorkerClient(pconn)
	// update cluster state every 30s
	for {
		pars, err := fetchPartitions(pclient)
		if err != nil {
			fmt.Printf("ERR#FS94GPOFD fetching partition: %v\n", err)
		}

		if len(pars) == 0 {
			continue
		}

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
				// Keep the state of this sc in b.scStates until sc's state becomes Shutdown.
				// The entry will be deleted in HandleSubConnStateChange.
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

	a := addrs[0]
	me.fetchLoop(a.Addr)
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
// partition map is an array tells which worker will handle which partition
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

	conf, err := pclient.GetConfig(context.Background(), &pb.Cluster{})
	if err != nil {
		return nil, err
	}

	// convert configuration fetched from worker to partition map
	partitions := make([]string, conf.GetTotalPartitions())
	for workerid, pars := range conf.GetPartitions() {
		for _, parNum := range pars.GetPartitions() {
			partitions[parNum] = conf.GetHosts()[workerid]
		}
	}
	return partitions, nil
}
