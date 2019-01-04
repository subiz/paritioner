package main

import (
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
	coor    *Coor
	db      *DB
	cluster string
	mapLock *sync.Mutex
	connMgr *ConnMgr
	voteMgr *VoteMgr
}

type workerVote struct {
	term     int32
	voteChan chan bool
}

func NewServer(kubeservice string, db *DB) *Server {
	s := &Server{Mutex: &sync.Mutex{}, cluster: kubeservice, db: db, mapLock: &sync.Mutex{}}
	s.connMgr = NewConnMgr()
	s.voteMgr = NewVoteMgr()
	s.coor = NewCoordinator(kubeservice, db, s)
	go s.lookupDNS()
	return s
}

func (me *Server) lookupDNS() {
	for {
		func() {
			me.Lock()
			defer me.Unlock()

			defer func() { recover() }()

			_, addrs, err := net.LookupSRV("", "", me.cluster)
			for _, record := range addrs {
				fmt.Printf("Target: %s:%d\n", record.Target, record.Port)
			}
			if err != nil {
				fmt.Printf("Could not get IPs: %v\n", err)
			}

			ips, err := net.LookupIP(me.cluster)
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

func (me *Server) GetConfig() *pb.Configuration { return me.coor.GetConfig() }

func (me *Server) Prepare(workerid string, conf *pb.Configuration) error {
	conn, ok := me.connMgr.Get(workerid)
	if !ok {
		return errors.New(500, errors.E_partition_node_have_not_joined_the_cluster,
			"worker stream not found", workerid)
	}

	stream := conn.Payload.(pb.Coordinator_RebalanceServer)
	if err := stream.Send(conf); err != nil {
		me.connMgr.Remove(conn)
		return err
	}

	return me.voteMgr.Wait(workerid, conf.GetTerm())
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

func safe(f func()) {
	func() {
		defer func() { recover() }()
		f()
	}()
}

func (me *Server) Vote(workerid string, term int32, accept bool) { me.voteMgr.Vote(workerid, term, accept) }

func (me *Server) Pull(workerid string, stream pb.Coordinator_RebalanceServer) {
	me.connMgr.Pull(workerid, stream)
}
