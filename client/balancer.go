package client

import (
	"context"
	"fmt"
	"github.com/subiz/errors"
	pb "github.com/subiz/partitioner/header"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"
	"sync"
	"time"
)

// parBalancer is shorten for partitioner balancer
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
	// loop once only
	me.Lock()
	if me.fetching {
		me.Unlock()
		return
	}
	me.fetching = true
	me.Unlock()

	// trying to connect to a worker
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
