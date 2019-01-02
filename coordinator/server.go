package main

import (
	"context"
	"fmt"
	"github.com/subiz/errors"
	"github.com/subiz/goutils/log"
	pb "github.com/subiz/header/partitioner"
	"google.golang.org/grpc"
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

func NewServer(kubeservice string, db *DB) *Server {
	s := &Server{Mutex: &sync.Mutex{}}
	s.kubeservice = kubeservice
	s.cluster = kubeservice
	s.dialLock = &sync.Mutex{}
	s.db = db
	s.workers = make(map[string]pb.WorkerClient)
	s.coor = NewCoordinator(kubeservice, db, s)
	var err error
	s.hosts, err = db.LoadHosts(kubeservice)
	if err != nil {
		panic(err)
	}
	go s.lookupDNS()
	return s
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

func (me *Server) Join(id, host string) error {
	me.Lock()
	defer me.Unlock()

	if err := me.db.SaveHost(me.cluster, id, host); err != nil {
		return err
	}
	me.hosts[id] = host
	return nil
}

func (me *Server) GetConfig() *pb.Configuration {
	conf := me.coor.GetConfig()

	// refill hosts info since coor only store worker ids
	hosts := make(map[string]string)
	me.Lock()
	for workerid, _ := range conf.GetPartitions() {
		hosts[workerid] = me.hosts[workerid]
	}
	me.Unlock()
	conf.Hosts = hosts
	return conf
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
					err = errors.New(500, errors.E_unknown, r, workerid, host)
				}
			}()

			conn, er := dialGrpc(host)
			if er != nil {
				err = errors.Wrap(er, 500, errors.E_unknown, workerid, host)
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

	if _, err := w.Prepare(context.Background(), conf); err != nil {
		return errors.Wrap(err, 500, errors.E_unknown, workerid, host)
	}
	return nil
}

func dialGrpc(service string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	// Enabling WithBlock tells the client to not give up trying to find a server
	opts = append(opts, grpc.WithBlock())
	// However, we're still setting a timeout so that if the server takes too long, we still give up
	opts = append(opts, grpc.WithTimeout(1*time.Second))
	return grpc.Dial(service, opts...)
}
