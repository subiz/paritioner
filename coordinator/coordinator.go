package coordinator

import (
	"context"
	"fmt"
	"google.golang.org/grpc/balancer/roundrobin"
	// "github.com/subiz/errors"
	"github.com/golang/protobuf/proto"
	pb "github.com/subiz/header/partitioner"
	"google.golang.org/grpc"
	"sync"
	"time"
)

// Coor is a coordinator implementation
type Coor struct {
	*sync.Mutex
	db      *DB
	config  *pb.Configuration // intermediate configuration, TODO: this should never benill
	version string
	cluster string

	dialLock *sync.Mutex
	workers  map[string]pb.WorkerClient
}

func NewCoordinator(cluster string, db *DB) *Coor {
	me := &Coor{Mutex: &sync.Mutex{}}
	me.version = "1.0.0"
	me.cluster = cluster
	me.dialLock = &sync.Mutex{}
	conf, err := me.db.Load()
	if err != nil {
		panic(err)
	}
	me.config = conf
	return me
}

func (me *Coor) validateRequest(version, cluster string, term int32) error {
	if version != me.version {
		//return errors.New(400, errors.E_invalid_version, "only support version "+me.Version)
	}

	if cluster != me.cluster {
		//return errors.New(400, errors.E_invalid_cluster, "cluster should be "+me.Cluster+" not "+cluster)
	}

	if term < me.config.GetTerm() {
		//return errors.New(400, errors.E_invalid_term, "term should be %d, not %d", me.Term, term)
	}
	return nil
}

func (me *Coor) GetConfig(ctx context.Context, em *pb.Empty) (*pb.Configuration, error) {
	me.Lock()
	defer me.Unlock()
	return me.config, nil
}

func (me *Coor) Join(ctx context.Context, join *pb.JoinRequest) (*pb.Configuration, error) {
	me.Lock()
	defer me.Unlock()

	err := me.validateRequest(join.GetVersion(), join.GetCluster(), join.GetTerm())
	if err != nil {
		return nil, err
	}

	found := false
	for _, w := range me.config.GetWorkers() {
		if w.Id != join.GetId() {
			continue
		}
		found = true
		w.Host = join.Host
	}

	me.config.Term = me.config.GetNextTerm()
	me.config.NextTerm = me.config.GetNextTerm() + 1

	if found {
		if err := me.db.Store(me.config); err != nil {
			return nil, err
		}
	}
	return me.config, nil
}

// make sure all other nodes are died
// this protocol assume that nodes are all nodes that survived
func (me *Coor) Change(newWorkers []string) {
	me.Lock()
	defer me.Unlock()

	workerM := make(map[string][]int32)
	for _, w := range me.config.GetWorkers() {
		workerM[w.GetId()] = w.GetPartitions()
	}

	workerM = balance(workerM, newWorkers)
	newConfig := proto.Clone(me.config).(*pb.Configuration)
	newConfig.Workers = make([]*pb.WorkerConfiguration, 0)

	for _, w := range newWorkers {
		host := ""
		for _, oldW := range me.config.GetWorkers() {
			if oldW.GetId() == w {
				host = oldW.GetHost()
			}
		}

		wc := &pb.WorkerConfiguration{Id: w, Host: host, Partitions: workerM[w]}
		newConfig.Workers = append(newConfig.Workers, wc)
	}

	newConfig.Term = newConfig.NextTerm
	newConfig.NextTerm++

	me.config.NextTerm++
	if err := me.db.Store(me.config); err != nil {
		fmt.Printf("ERR #DB093FDPW db, %v\n", err)
		return
	}

	responseC := make(chan bool)
	for _, wc := range newConfig.Workers {
		go func(wc *pb.WorkerConfiguration) {
			responseC <- me.requestWorker(wc.GetHost(), newConfig)
		}(wc)
	}

	ticker := time.NewTicker(40 * time.Second)

	total := 0
	for {
		select {
		case res := <-responseC:
			if !res {
				goto failed
			}

			total++
			if total == len(me.config.GetWorkers()) {
				// success
				//me.config = newConfigs
				return
			}
		case <-ticker.C:
			fmt.Printf("TIMEOUT, got only %d respones", total)
			goto failed
		}
	}

failed:
}

func dialGrpc(service string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	// Enabling WithBlock tells the client to not give up trying to find a server
	opts = append(opts, grpc.WithBlock())
	// However, we're still setting a timeout so that if the server takes too long, we still give up
	opts = append(opts, grpc.WithTimeout(10*time.Second))
	opts = append(opts, grpc.WithBalancerName(roundrobin.Name))
	//opts = append(opts, grpc.WithBalancer(grpc.RoundRobin(res)))
	return grpc.Dial(service, opts...)
}

func (me *Coor) requestWorker(host string, p *pb.Configuration) bool {
	me.dialLock.Lock()
	w, ok := me.workers[host]
	if !ok {
		conn, err := dialGrpc(host)
		if err != nil {
			me.dialLock.Unlock()
			return false
		}
		w = pb.NewWorkerClient(conn)
		me.workers[host] = w
		me.dialLock.Unlock()
	}

	_, err := w.Request(context.Background(), p)
	if err != nil {
		fmt.Printf("ERR #DJKWPK85JD, grpc err %v\n", err)
		return false
	}

	return true
}
