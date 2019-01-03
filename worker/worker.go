package worker

import (
	"context"
	"fmt"
	"github.com/subiz/errors"
	pb "github.com/subiz/header/partitioner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/metadata"
	"hash/fnv"
	"reflect"
	"strings"
	"sync"
	"time"
)

const (
	// represense blocked state of an partition mean...
	BLOCKED = -1
	NORMAL  = 1

	PartitionKey = "partitionkey"
)

// global hashing util, used to hash key to partition number
var ghash = fnv.New32a()

type partition struct {
	worker_id           string
	worker_host         string
	state               int // NORMAL, BLOCKED,
	prepare_worker_id   string
	prepare_worker_host string
}

type Worker struct {
	*sync.Mutex
	id         string
	partitions []partition
	version    string
	cluster    string
	term       int
	coor       pb.CoordinatorClient
	host       string
	onUpdates  []func([]int32)
	conn       map[string]*grpc.ClientConn
	dialMutex  *sync.Mutex
}

// NewWorker creates and starts a new Worker object
// parameters:
//  host: grpc host address for coordinator or other workers to connect
//    to this worker. E.g: web-1.web:8080
//  cluster: ID of cluster. E.g: web
//  id: ID of the worker, this must be unique for each workers inside
//    cluster. E.g: web-1
//  coordinator: host address of coordinator, used to listen changes
//    and making transaction. E.g: coordinator:8021
func NewWorker(host string, cluster, id, coordinator string) *Worker {
	me := &Worker{
		Mutex:     &sync.Mutex{},
		version:   "1.0.0",
		cluster:   cluster,
		term:      0,
		id:        id,
		host:      host,
		conn:      make(map[string]*grpc.ClientConn),
		dialMutex: &sync.Mutex{},
	}

	cconn, err := dialGrpc(coordinator)
	if err != nil {
		panic(err)
	}
	me.coor = pb.NewCoordinatorClient(cconn)
	fmt.Println("JOINING THE CLUSTER...")
	for {
		_, err := me.coor.Join(context.Background(), &pb.WorkerHost{
			Cluster: cluster,
			Host:    host,
			Id:      id,
		})
		if err == nil {
			break
		}
		fmt.Printf("ERR while joining cluster %v. Retry in 2 secs\n", err)
		time.Sleep(2 * time.Second)
	}

	fmt.Println("JOINED")
	me.fetchConfig()
	return me
}

// OnUpdate registers a subscriber to workers.
// The subscriber get call synchronously when worker's partitions is changed
// worker will pass list of its new partitions through subscriber parameter
func (me *Worker) OnUpdate(f func([]int32)) {
	me.Lock()
	defer me.Unlock()
	me.onUpdates = append(me.onUpdates, f)
}

// fetchConfig updates worker partition map to the current configuration of
// coordinator.
// Note: this function may take up to ten of seconds to execute since the
// cluster is in middle of rebalance process. During that time, coordinator
// would be blocking any requests.
func (me *Worker) fetchConfig() {
	var conf *pb.Configuration
	var err error
	for {
		fmt.Println("FETCHING CONFIG")
		conf, err = me.coor.GetConfig(context.Background(), &pb.Cluster{Id: me.cluster})
		if err != nil {
			fmt.Printf("ERR #234FOISDOUD config %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		err := me.validateRequest(conf.GetVersion(), conf.GetCluster(), int(conf.GetTerm()))
		if err != nil {
			fmt.Printf("ERR #234DDFSDOUD OUTDATED %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	me.Lock()
	// rebuild partition map using coordinator's configuration
	me.term = int(conf.GetTerm())
	me.partitions = make([]partition, conf.GetTotalPartitions())
	for workerid, pars := range conf.GetPartitions() {
		for _, p := range pars.GetPartitions() {
			me.partitions[p].worker_id = workerid
			me.partitions[p].worker_host = conf.GetHosts()[workerid]

			me.partitions[p].prepare_worker_id = ""
			me.partitions[p].prepare_worker_host = ""
			me.partitions[p].state = NORMAL
		}
	}

	// notify subscriber about the change
	// this must done synchronously since order of changes might be
	// critical with some subscribers
	ourpars := make([]int32, 0)
	for p, par := range me.partitions {
		if par.worker_id == me.id {
			ourpars = append(ourpars, int32(p))
		}
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(me.onUpdates))
	for _, f := range me.onUpdates {
		func() { // wrap in a function to prevent panicing
			defer func() {
				recover()
				wg.Done()
			}()
			f(ourpars)
		}()
	}
	wg.Wait()
	me.Unlock()
}

// validateRequest makes sure version, cluster and term are upto date and
//   is sent to correct cluster
func (me *Worker) validateRequest(version, cluster string, term int) error {
	if version != me.version {
		return errors.New(400, errors.E_invalid_partition_version, "only support version "+me.version)
	}

	if cluster != me.cluster {
		return errors.New(400, errors.E_invalid_partition_cluster, "cluster should be "+me.cluster+" not "+cluster)
	}

	if term < me.term {
		return errors.New(400, errors.E_invalid_partition_term, "term should be %d, not %d", me.term, term)
	}
	return nil
}

// GetConfig is a GRPC handler, it returns current configuration of the cluster
func (me *Worker) GetConfig(ctx context.Context, cluster *pb.Cluster) (*pb.Configuration, error) {
	cluster.Id = me.cluster
	return me.coor.GetConfig(ctx, cluster)
}

func (me *Worker) Prepare(ctx context.Context, conf *pb.Configuration) (*pb.Empty, error) {
	err := me.validateRequest(conf.GetVersion(), conf.GetCluster(), int(conf.GetTerm()))
	if err != nil {
		return nil, err
	}

	me.Lock()
	for workerid, pars := range conf.GetPartitions() {
		for _, p := range pars.GetPartitions() {
			me.partitions[p].prepare_worker_id = workerid
			me.partitions[p].prepare_worker_host = conf.GetHosts()[workerid]

			if me.partitions[p].prepare_worker_id != me.partitions[p].worker_id {
				me.partitions[p].state = BLOCKED
			}
		}
	}
	me.Unlock()

	go me.fetchConfig()
	return &pb.Empty{}, nil
}

// analysisReturnType returns all return types for every GRPC method in server handler
// the returned map takes full method name (i.e., /package.service/method) as key, and the return type as value
// For example, with handler
//   (s *server) func Goodbye() string {}
//   (s *server) func Ping(_ context.Context, _ *pb.Ping) (*pb.Pong, error) {}
//   (s *server) func Hello(_ context.Context, _ *pb.Empty) (*pb.String, error) {}
// this function detected 2 GRPC methods is Ping and Hello, it would return
// {"Ping": *pb.Pong, "Hello": *pb.Empty}
func analysisReturnType(server interface{}) map[string]reflect.Type {
	m := make(map[string]reflect.Type)
	t := reflect.TypeOf(server).Elem()
	for i := 0; i < t.NumMethod(); i++ {
		methodType := t.Method(i).Type
		if methodType.NumOut() != 2 || methodType.NumIn() < 2 {
			println("skip dd", t.Method(i).Name)
			continue
		}

		if methodType.Out(1).Kind() != reflect.Ptr || methodType.Out(1).Name() != "context" {
			continue
		}

		if methodType.Out(0).Kind() != reflect.Ptr || methodType.Out(1).Name() != "error" {
			continue
		}

		m[t.Method(i).Name] = methodType.Out(0).Elem()
		//println("d", reflect.New().Interface().(string))
		//println(t.Method(i).Type.Out(0).Name())
	}
	return m
}

// CreateIntercept makes a new GRPC server interceptor
func (me *Worker) CreateIntercept(mgr interface{}) grpc.UnaryServerInterceptor {
	returnedTypeM := analysisReturnType(mgr)
	return func(ctx context.Context, in interface{}, sinfo *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (out interface{}, err error) {
		md, _ := metadata.FromIncomingContext(ctx)
		pkey := strings.Join(md[PartitionKey], "")

		if pkey == "" {
			return handler(ctx, in)
		}

		ghash.Write([]byte(pkey))
		parindex := ghash.Sum32() % uint32(len(me.partitions)) // 1024
		ghash.Reset()

		me.Lock()
		par := me.partitions[parindex]
		me.Unlock()

		// block
		for par.state == BLOCKED {
			time.Sleep(5 * time.Second)
			me.Lock()
			par = me.partitions[parindex]
			me.Unlock()
		}

		if par.worker_id == me.id { // correct parittion
			return handler(ctx, in)
		}

		return me.forward(par.worker_host, sinfo.FullMethod, returnedTypeM[sinfo.FullMethod], ctx, in)
	}
}

// forward redirects a GRPC calls to another host, header and trailer are preserved
// parameters:
//   host: host address which will be redirected to
//   method: the full RPC method string, i.e., /package.service/method.
//   returnedType: type of returned value
//   in: value of input (in request) parameter
// this method returns output just like a normal GRPC call
func (me *Worker) forward(host, method string, returnedType reflect.Type, ctx context.Context, in interface{}) (interface{}, error) {

	// use cache host connection or create a new one
	me.dialMutex.Lock()
	cc, ok := me.conn[host]
	if !ok {
		var err error
		cc, err = dialGrpc(host)
		if err != nil {
			me.dialMutex.Unlock()
			return nil, err
		}
		me.conn[host] = cc
		me.dialMutex.Unlock()
	}

	md, _ := metadata.FromIncomingContext(ctx)
	outctx := metadata.NewOutgoingContext(context.Background(), md)

	out := reflect.New(returnedType).Interface()
	var header, trailer metadata.MD
	err := cc.Invoke(outctx, method, in, out, grpc.Header(&header), grpc.Trailer(&trailer))
	grpc.SendHeader(ctx, header)
	grpc.SetTrailer(ctx, trailer)

	return out, err
}

// dialGrpc makes a GRPC connection to service
func dialGrpc(service string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	// Enabling WithBlock tells the client to not give up trying to find a server
	opts = append(opts, grpc.WithBlock())
	// However, we're still setting a timeout so that if the server takes too long, we still give up
	opts = append(opts, grpc.WithTimeout(5*time.Second))
	opts = append(opts, grpc.WithBalancerName(roundrobin.Name))
	return grpc.Dial(service, opts...)
}
