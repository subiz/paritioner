package client

import (
	"context"
	"fmt"
	pb "github.com/subiz/header/partitioner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"hash/fnv"
	"strings"
	"sync"
	"time"
)

const (
	// GRPC Metadata key to store partition key attached to each GRPC request
	// the key will be hashed to find the correct worker that handle the request
	PartitionKey = "partitionkey"
)

// global hashing util, used to hash key to partition number
var ghash = fnv.New32a()

// Client is used to partition request through workers in cluster
// user uses this struct by create an GRPC interceptor, the interceptor will
// trigger right before the request is sent, redirecting the request to assigned
// worker
type Client struct {
	*sync.RWMutex
	conn       map[string]*grpc.ClientConn
	partitions []string
}

func dialGrpc(service string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	// Enabling WithBlock tells the client to not give up trying to find a server
	opts = append(opts, grpc.WithBlock())
	// However, we're still setting a timeout so that if the server takes too long, we still give up
	opts = append(opts, grpc.WithTimeout(2*time.Second))
	return grpc.Dial(service, opts...)
}

// NewInterceptor create new GRPC interceptor to partition request
// service is one of workers host, e.g: web-0.web:8080. This host will be used
// to discover all workers in cluster.
func NewInterceptor(service string) grpc.DialOption {
	me := &Client{RWMutex: &sync.RWMutex{}}
	me.conn = make(map[string]*grpc.ClientConn)

	pconn, err := dialGrpc(service)
	if err != nil {
		panic(err)
	}

	pclient := pb.NewWorkerClient(pconn)

	pars, err := fetchPartitions(pclient)
	if err != nil {
		panic(err)
	}

	me.partitions = pars

	// update cluster state every 30s
	go me.fetchLoop(pclient)

	return grpc.WithUnaryInterceptor(me.clientInterceptor)
}

// fetchPartitions calls worker api to return the latest partition map
// partition map is an array tells which worker will handle which partition
// partition number is ordinal index of array, each element contains worker
// host
// eg: ["worker-0:8081", "worker-1:8081", "worker-0:8081"] has 3 partitions:
// {0, 1, 2}, partition 0 and 2 are handled by worker 0 at host worker-0:8081
// partition 1 is handled by worker 1 at host worker-1:8081
func fetchPartitions(pclient pb.WorkerClient) ([]string, error) {
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

// clientInterceptor is an GRPC client interceptor, it redirects the request to
// the correct worker host
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
func (me *Client) clientInterceptor(ctx context.Context, method string, in interface{}, out interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md, _ := metadata.FromOutgoingContext(ctx)
	pkey := strings.Join(md[PartitionKey], "")

	if pkey == "" {
		return invoker(ctx, method, in, out, cc, opts...)
	}

	// hashing key to find the partition number
	ghash.Write([]byte(pkey))
	par := ghash.Sum32() % uint32(len(me.partitions))
	ghash.Reset()

	host := me.partitions[par]
	me.Lock()
	cc = me.conn[host]
	if cc == nil {
		var err error
		cc, err = func() (*grpc.ClientConn, error) {
			defer func() { recover() }()
			return dialGrpc(host) // TODO: this could cause bottle neck if user send
			// many requests to a not-responding host
		}()
		if err != nil {
			me.Unlock()
			return err
		}
		me.conn[host] = cc
		me.Unlock()
	}
	return cc.Invoke(ctx, method, in, out, opts...)
}

// fetchLoop synchronizes partition in every 30 sec
func (me *Client) fetchLoop(pclient pb.WorkerClient) {
	for {
		var pars []string
		func() { // wrap in a function to prevent panicing
			defer func() { recover() }()
			var err error
			pars, err = fetchPartitions(pclient)
			if err != nil {
				fmt.Printf("ERR#FS94GPOFD fetching partition: %v\n", err)
			}
		}()

		if len(pars) == 0 {
			continue
		}

		me.Lock()
		me.partitions = pars
		me.Unlock()
		time.Sleep(30 * time.Second)
	}
}
