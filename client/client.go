package client

import (
	"context"
	"fmt"
	pb "github.com/subiz/header/partitioner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/metadata"
	"hash/fnv"
	"strings"
	"sync"
	"time"
)

const (
	PartitionKey = "partitionkey"
)

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

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
	opts = append(opts, grpc.WithTimeout(10*time.Second))
	//res, err := naming.NewDNSResolverWithFreq(1 * time.Second)
	//if err != nil {
	//	return nil, err
	//}
	opts = append(opts, grpc.WithBalancerName(roundrobin.Name))
	//opts = append(opts, grpc.WithBalancer(grpc.RoundRobin(res)))
	return grpc.Dial(service, opts...)
}

func NewClient(service string) *Client {
	// connect to grpc cluster

	pconn, err := dialGrpc(service)
	if err != nil {
		panic(err)
	}
	// parittioner client
	pclient := pb.NewWorkerClient(pconn)
	me := &Client{RWMutex: &sync.RWMutex{}}

	pars, err := fetchPartitions(pclient)
	if err != nil {
		panic(err)
	}

	me.partitions = pars

	// update cluster state every 30s
	go me.fetchLoop(pclient)
	return me
}

func fetchPartitions(pclient pb.WorkerClient) ([]string, error) {
	conf, err := pclient.GetConfig(context.Background(), &pb.Cluster{})
	if err != nil {
		return nil, err
	}

	partitions := make([]string, conf.GetTotalPartitions())
	for workerid, pars := range conf.GetPartitions() {
		for _, parNum := range pars.GetPartitions() {
			partitions[parNum] = conf.GetHosts()[workerid]
		}
	}
	return partitions, nil
}

func (me *Client) fetchLoop(pclient pb.WorkerClient) {
	for {
		var pars []string
		var err error
		func() {
			defer func() { recover() }()
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

// forward to the right host
func (me *Client) clientInterceptor(ctx context.Context, method string, in interface{}, out interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md, _ := metadata.FromOutgoingContext(ctx)
	pkey := strings.Join(md[PartitionKey], "")

	if pkey == "" {
		return invoker(ctx, method, in, out, cc, opts...)
	}

	par := hash(pkey) % 1000
	host := me.partitions[par]
	cc = me.conn[host]
	if cc == nil {
		var err error
		me.Lock()
		cc, err = func() (*grpc.ClientConn, error) {
			defer func() { recover() }()
			return dialGrpc(host)
		}()
		if err != nil {
			me.Unlock()
			return err
		}
		me.conn[host] = cc
	}

	return cc.Invoke(ctx, method, in, out, opts...)
}

func (me *Client) WithInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(me.clientInterceptor)
}
