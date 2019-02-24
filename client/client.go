package client

import (
	"context"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "github.com/subiz/partitioner/header"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
	// string: id
	conn map[string]*grpc.ClientConn

	port string

	// partitions id
	partitions []string
}

// dialGRPC makes a connection to GRPC addr (e.g: grpc.subiz.com:8080)
func dialGrpc(addr string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	// Enabling WithBlock tells the client to not give up trying to find a server
	opts = append(opts, grpc.WithBlock())
	// However, we're still setting a timeout so that if the server takes too long, we still give up
	opts = append(opts, grpc.WithTimeout(2*time.Second))
	return grpc.Dial(addr, opts...)
}

// NewInterceptor create new GRPC interceptor to partition request
// service is one of workers host, e.g: web-0.web. This host will be used
// to discover all workers in cluster.
func NewInterceptor(service string, port int) grpc.DialOption {
	me := &Client{RWMutex: &sync.RWMutex{}}
	me.conn = make(map[string]*grpc.ClientConn)
	me.port = strconv.Itoa(port)
	pconn, err := dialGrpc(service + ":" + me.port)
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

// clientInterceptor is an GRPC client interceptor, it redirects the request to
// the correct worker host
// We find worker host by hashing partition key to a partition number then
// look it up in the current partition map
// worker host might not be the correct one since partition map can be
// unsynchronized (by network lantences or network partitioned). When it
// happend, request will be redirected one or more times in worker side till
// it arrive the correct host. This may look ineffiecient but we don't have to
// guarrantee consistent at client-side, make client-side code so much simpler
// when the network is stable and no worker member changes (most of the time),
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

	id := me.partitions[par]
	me.Lock()
	cc = me.conn[id]
	if cc == nil {
		var err error
		cc, err = func() (*grpc.ClientConn, error) {
			defer func() { recover() }()
			return dialGrpc(id + ":" + me.port) // TODO: this could cause bottle neck if user send
			// many requests to a not-responding host
		}()
		if err != nil {
			me.Unlock()
			return err
		}
		me.conn[id] = cc
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
