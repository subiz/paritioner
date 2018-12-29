package worker

import (
	"reflect"
	"hash/fnv"
	pb "github.com/subiz/header/partitioner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/balancer/roundrobin"
)

const (
	MISS=0
	BLOCKED=-1
	NORMAL=1

	PartitionKey = "partitionkey"
)

func hash(s string) uint32 {
        h := fnv.New32a()
        h.Write([]byte(s))
        return h.Sum32()
}

type Worker struct {
	*sync.Mutex
	Config *pb.WorkerConfiguration
	partitions []int
	partitionLocs []string
	blocked_partition []bool
	conn map[string]*grpc.ClientConn
	GlobalConfig  *pb.Configuration // intermediate configuration
	Version string
	handlerOutType map[string]reflect.Type
	Cluster string
	Term    int32
	coor    CoordinatorMgr
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

type CoordinatorMgr interface {
	Config() *pb.Configuration
}

func (w *Worker) CorrectJob(key string)

func (w *Worker) Run() {

}

func (me *Worker) Commit() {

}

func (me *Worker) Abort() {

}

func  analysis(server interface{}) map[string]reflect.Type {
	m := make(map[string]reflect.Type)
	t := reflect.TypeOf(server).Elem()
	var s []string
	for i := 0; i < t.NumMethod(); i++ {
		methodType := t.Method(i).Type
		if methodType.NumOut() != 2 || methodType.NumIn() != 2 {
			continue
		}

		if methodType.Out(1).Kind() != reflect.Ptr || methodType.Out(1).Name() != "context" {
			contine
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



func (me *Worker) forward(ctx context.Context,	host string, in interface{}, serverinfo *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	out := new(Account)
	md, _ := metadata.FromIncomingContext(ctx)
	outctx := metadata.NewOutgoingContext(context.Background(), md)

	cc := me.conn[host]
	if cc == nil {
		cc, err = dialGrpc(host)
		if err != nil {
			return err
		}
		me.conn[host] = cc
	}

	out := me.makeOut(serverinfo.FullMethod)

	var header, trailer metadata.MD
	err := cc.Invoke(outctx, serverinfo.FullMethod, in, out, grpc.Header(&header),  grpc.Trailer(&trailer))
	grpc.SendHeader(ctx, header)
	grpc.SetTrailer(ctx, trailer)

	return out, err
}

func (me *Worker) Intercept(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (ret interface{}, err error) {
	md, _ := metadata.FromIncomingContext(ctx)
	pkey := strings.Join(md[PartitionKey], "")


	// busy waiting until we back to normal state

	if pkey == "" {
		return handler(ctx, req)
	}
	par := hash(pkey) % 1000
	if me.partitions[par]==NORMAL { // correct partition
		return handler(ctx, req)
	}

	if me.partitions[par]==MISS {
		// lookup correct node
		return handler(ctx, req)
	}



	func() {
		defer func() {
			if r := recover(); r != nil {
				e, ok := r.(error)
				if ok {
					err = errors.Wrap(e, 500, errors.E_unknown)
				}

				err = errors.New(500, errors.E_unknown, fmt.Sprintf("%v", e))
			}
		}()
		ret, err = handler(ctx, req)
	}()
	if err != nil {
		e, ok := err.(*errors.Error)
		if !ok {
			e = errors.Wrap(err, 500, errors.E_unknown)
		}
		md := metadata.Pairs(PanicKey, e.Error())
		grpc.SendHeader(ctx, md)
	}
	return ret, err
}
