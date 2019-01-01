package coordinator

import (
	"context"
	"fmt"
	"github.com/subiz/errors"
	"github.com/subiz/goutils/log"
	pb "github.com/subiz/header/partitioner"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"net"
	"sync"
	"time"
)

type Server struct {
	*sync.Mutex
	coor        *Coor
	dialLock    *sync.Mutex
	db          *DB
	kubeservice string
	hosts       map[string]string
	workers     map[string]pb.WorkerClient
	cluster     string
}

func RunServe(cluster, kubeservice string, cassandraSeeds []string, port string) error {
	me := &Server{Mutex: &sync.Mutex{}}
	me.kubeservice = kubeservice
	me.cluster = cluster
	me.dialLock = &sync.Mutex{}
	me.db = NewDB(cassandraSeeds, cluster)
	me.coor = NewCoordinator(cluster, me.db, me)
	var err error
	me.hosts, err = me.db.LoadHosts()
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterCoordinatorServer(grpcServer, me)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	return grpcServer.Serve(lis)
}

func (me *Server) lookupDNS() {
	for {
		func() {
			me.Lock()
			defer me.Unlock()

			defer func() { recover() }()
			ips, err := net.LookupIP(me.kubeservice)
			if err != nil {
				fmt.Printf("Could not get IPs: %v\n", err)
				return
			}

			fmt.Println("looking up dns, got", ips)
			conf := me.coor.GetConfig()
			if len(ips) == len(conf.GetPartitions()) { // no change
				return
			}
			workers := make([]string, 0)
			for i := 0; i < len(ips); i++ {
				workers = append(workers, fmt.Sprintf("%s-%d", me.cluster, i))
			}
			if err := me.coor.ChangeWorkers(workers); err != nil {
				log.Error(err)
				return
			}
		}()
		time.Sleep(2 * time.Second)
	}
}

func (me *Server) Join(ctx context.Context, host *pb.WorkerHost) (*pb.Empty, error) {
	me.Lock()
	defer me.Unlock()

	if err := me.db.SaveHost(host.GetId(), host.GetHost()); err != nil {
		return nil, err
	}
	me.hosts[host.GetId()] = host.GetHost()
	return &pb.Empty{}, nil
}

func (me *Server) Prepare(workerid string, conf *pb.Configuration) error {
	host, ok := me.hosts[workerid]
	if !ok {
		return errors.New(500, errors.E_partition_node_have_not_joined_the_cluster, workerid)
	}

	var err error
	me.dialLock.Lock()
	w, ok := me.workers[host]
	if !ok {
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = errors.New(500, errors.E_unknown, r)
				}
			}()

			conn, er := dialGrpc(host)
			if er != nil {
				err = er
				return
			}
			w = pb.NewWorkerClient(conn)
			me.workers[host] = w
		}()
	}
	me.dialLock.Unlock()
	if err != nil {
		return err
	}

	_, err = w.Prepare(context.Background(), conf)
	return err
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

func (me *Server) GetConfig(ctx context.Context, em *pb.Empty) (*pb.Configuration, error) {
	return me.coor.GetConfig(), nil
}
