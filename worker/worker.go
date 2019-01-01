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
	worker_id           string
	worker_host         string
	state               int // OFF, NORMAL, BLOCKED,
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
}

func dialGrpc(service string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	// Enabling WithBlock tells the client to not give up trying to find a server
	opts = append(opts, grpc.WithBlock())
	// However, we're still setting a timeout so that if the server takes too long, we still give up
	opts = append(opts, grpc.WithTimeout(10*time.Second))
	opts = append(opts, grpc.WithBalancerName(roundrobin.Name))
	return grpc.Dial(service, opts...)
}

func (me *Worker) fetchConfig() {
	var conf *pb.Configuration
	var err error
	for {
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
	me.Unlock()
}

func NewWorker(host string, cluster, id, coordinator string) *Worker {
	me := &Worker{
		Mutex:   &sync.Mutex{},
		version: "1.0.0",
		cluster: cluster,
		term:    0,
		id:      id,
		host:    host,
	}

	cconn, err := dialGrpc(coordinator)
	if err != nil {
		panic(err)
	}
	me.coor = pb.NewCoordinatorClient(cconn)
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

	me.fetchConfig()
	return me
}

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

func (me *Worker) GetHost(ctx context.Context, cluster *pb.Cluster) (*pb.WorkerHost, error) {
	return &pb.WorkerHost{Cluster: me.cluster, Host: me.host, Id: me.id}, nil
}

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
	return nil, nil
}

func analysis(server interface{}) map[string]reflect.Type {
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


func (me *Worker) CreateIntercept(mgr interface{}) grpc.UnaryServerInterceptor {
	outTypeM := analysis(mgr)
	conn := &sync.Map{}
	dialMutex := &sync.Mutex{}
	return func(ctx context.Context, in interface{}, sinfo *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (out interface{}, err error) {
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

		dialMutex.Lock()
		cci, ok := conn.Load(par.worker_host)
		if !ok {
			var err error
			cci, err = dialGrpc(par.worker_host)
			if err != nil {
				dialMutex.Unlock()
				return nil, err
			}
			conn.Store(par.worker_host, cci)
			dialMutex.Unlock()
		}
		return forward(cci.(*grpc.ClientConn), outTypeM, ctx, in, sinfo)
	}
}

// key is method name (not fullmethod), value is type (not pointer) of return value
func forward(cc *grpc.ClientConn, outType map[string]reflect.Type, ctx context.Context, in interface{}, serverinfo *grpc.UnaryServerInfo) (interface{}, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	outctx := metadata.NewOutgoingContext(context.Background(), md)

	out := reflect.New(outType[serverinfo.FullMethod]).Interface()

	var header, trailer metadata.MD
	err := cc.Invoke(outctx, serverinfo.FullMethod, in, out, grpc.Header(&header), grpc.Trailer(&trailer))
	grpc.SendHeader(ctx, header)
	grpc.SetTrailer(ctx, trailer)

	return out, err
}
