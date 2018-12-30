package worker

import (
	"context"
	"fmt"
	//	"github.com/subiz/errors"
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
	BLOCKED = -1
	NORMAL  = 1

	PartitionKey = "partitionkey"
)

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

type partition struct {
	worker_id   string
	worker_host string
	state       int // OFF, NORMAL, BLOCKED,

	prepare_worker_id   string
	prepare_worker_host string
}

type Worker struct {
	*sync.Mutex
	id     string // worker id
	config *pb.Configuration

	partitions []partition

	conn map[string]*grpc.ClientConn

	version        string
	handlerOutType map[string]reflect.Type // key is method name (not fullmethod), value is type (not pointer) of return value
	cluster        string
	term           int
	coor           pb.CoordinatorClient
}

// findConfig lookups worker configuration by worker ID in global configuration
// return nil if not found
func findConfig(conf *pb.Configuration, id string) *pb.WorkerConfiguration {
	for _, w := range conf.GetWorkers() {
		if w.GetId() == id {
			return w
		}
	}
	return nil
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

func (me *Worker) fetchConfig() {
	var conf *pb.Configuration
	var err error
	for {
		conf, err = me.coor.GetConfig(context.Background(), &pb.Empty{})
		if err != nil {
			fmt.Printf("ERR #234FSDOUD OUTDATED %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		err := me.validateRequest(conf.GetVersion(), conf.GetCluster(), int(conf.GetTerm()))
		if err != nil {
			fmt.Printf("ERR #234FSDOUD OUTDATED %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	me.Lock()
	me.config = conf
	me.term = int(conf.GetTerm())
	me.partitions = make([]partition, conf.GetTotalPartitions())
	for _, w := range conf.GetWorkers() {
		for _, p := range w.GetPartitions() {
			me.partitions[p].worker_id = w.GetId()
			me.partitions[p].worker_host = w.GetHost()
			me.partitions[p].prepare_worker_id = ""
			me.partitions[p].prepare_worker_host = ""
			me.partitions[p].state = NORMAL
		}
	}
	me.Unlock()
}

func NewWorker(cluster, id, coordinator string) *Worker {
	me := &Worker{
		Mutex:          &sync.Mutex{},
		version:        "1.0.0",
		cluster:        cluster,
		conn:           make(map[string]*grpc.ClientConn),
		handlerOutType: make(map[string]reflect.Type),
		term:           0,
	}

	cconn, err := dialGrpc(coordinator)
	if err != nil {
		panic(err)
	}
	me.coor = pb.NewCoordinatorClient(cconn)
	me.fetchConfig()
	return me
}

func (me *Worker) validateRequest(version, cluster string, term int) error {
	if version != me.version {
		//return errors.New(400, errors.E_invalid_version, "only support version "+me.version)
	}

	if cluster != me.cluster {
		//return errors.New(400, errors.E_invalid_cluster, "cluster should be "+me.cluster+" not "+cluster)
	}

	if term < me.term {
		//return errors.New(400, errors.E_invalid_term, "term should be %d, not %d", me.term, term)
	}
	return nil
}

func (me *Worker) GetConfig(cluster, version string, term int) (*pb.Configuration, error) {
	err := me.validateRequest(version, cluster, term)
	if err != nil {
		return nil, err
	}

	return me.config, nil
}

func (me *Worker) Prepare(conf *pb.Configuration) error {
	err := me.validateRequest(conf.GetVersion(), conf.GetCluster(), int(conf.GetTerm()))
	if err != nil {
		return err
	}

	me.Lock()
	for _, w := range conf.GetWorkers() {
		for _, p := range w.GetPartitions() {
			me.partitions[p].prepare_worker_id = w.GetId()
			me.partitions[p].prepare_worker_host = w.GetHost()

			if me.partitions[p].prepare_worker_id != me.partitions[p].worker_id {
				me.partitions[p].state = BLOCKED
			}
		}
	}
	me.Unlock()

	go me.fetchConfig()
	return nil
}

func analysis(server interface{}) map[string]reflect.Type {
	m := make(map[string]reflect.Type)
	t := reflect.TypeOf(server).Elem()
	for i := 0; i < t.NumMethod(); i++ {
		methodType := t.Method(i).Type
		if methodType.NumOut() != 2 || methodType.NumIn() != 2 {
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

func (me *Worker) forward(ctx context.Context, host string, in interface{}, serverinfo *grpc.UnaryServerInfo) (interface{}, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	outctx := metadata.NewOutgoingContext(context.Background(), md)

	cc := me.conn[host]
	if cc == nil {
		var err error
		cc, err = dialGrpc(host)
		if err != nil {
			return nil, err
		}
		me.conn[host] = cc
	}

	out := reflect.New(me.handlerOutType[serverinfo.FullMethod]).Interface()

	var header, trailer metadata.MD
	err := cc.Invoke(outctx, serverinfo.FullMethod, in, out, grpc.Header(&header), grpc.Trailer(&trailer))
	grpc.SendHeader(ctx, header)
	grpc.SetTrailer(ctx, trailer)

	return out, err
}

func (me *Worker) Intercept(ctx context.Context, in interface{}, sinfo *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (out interface{}, err error) {
	md, _ := metadata.FromIncomingContext(ctx)
	pkey := strings.Join(md[PartitionKey], "")

	if pkey == "" {
		return handler(ctx, in)
	}

	parindex := hash(pkey) % 1000
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

	return me.forward(ctx, par.worker_host, in, sinfo)
}
