package main

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/subiz/errors"
	pb "github.com/subiz/partitioner/header"
	"sync"
	"time"
)

// Coor is a coordinator implementation
type Coor struct {
	*sync.Mutex
	db         *DB
	config     *pb.Configuration // intermediate configuration, TODO: this never be nil
	workerComm WorkerComm
	//	hosts      map[string]string
}

// Workers communator, used to send signal (message) to workers
type WorkerComm interface {
	Prepare(cluster, workerid string, conf *pb.Configuration) error
}

func NewCoordinator(cluster string, db *DB, workerComm WorkerComm) *Coor {
	me := &Coor{Mutex: &sync.Mutex{}}
	me.workerComm = workerComm
	me.db = db
	conf, err := me.db.Load(cluster)
	if err != nil {
		panic(err)
	}
	me.config = conf

	// init config
	if conf.GetCluster() == "" || conf.GetTotalPartitions() == 0 {
		me.config = &pb.Configuration{
			Version:         "1.0.0",
			Cluster:         cluster,
			Term:            0,
			NextTerm:        1,
			TotalPartitions: 1024,
		}
		if err := me.db.Store(me.config); err != nil {
			panic(err)
		}
	}

	return me
}

func (me *Coor) Join(w *pb.WorkerHost) error {
	me.Lock()
	defer me.Unlock()
	if err := me.db.SaveHost(me.config.Cluster, w.GetId(), w.GetHost()); err != nil {
		return err
	}

	// me.hosts[id] = host
}

func (me *Coor) validateRequest(version, cluster string, term int32) error {
	if version != me.config.Version {
		return errors.New(400, errors.E_invalid_partition_version,
			"only support version "+me.config.Version)
	}

	if cluster != me.config.Cluster {
		return errors.New(400, errors.E_invalid_partition_cluster,
			"cluster should be "+me.config.Cluster+" not "+cluster)
	}

	if term < me.config.Term {
		return errors.New(400, errors.E_invalid_partition_term,
			"term should be %d, not %d", me.config.Term, term)
	}
	return nil
}

func (me *Coor) GetConfig() *pb.Configuration {
	me.Lock()
	defer me.Unlock()
	return me.config
}

func (me *Coor) transition(newConf *pb.Configuration) error {
	me.Lock()
	defer me.Unlock()

	// ignore no change
	newb, _ := newConf.MarshalJSON()
	oldb, _ := me.config.MarshalJSON()
	if bytes.Compare(newb, oldb) == 0 {
		return nil
	}

	newConf.Term, newConf.NextTerm = me.config.NextTerm, me.config.NextTerm+1

	// save next term to database
	me.config.NextTerm++
	if err := me.db.Store(me.config); err != nil {
		return err
	}

	responseC := make(chan error)
	for _, id := range newWorkers {
		go func(id string) {
			err := me.workerComm.Prepare(me.config.Cluster, id, newConf)
			select {
			case responseC <- err:
			default:
			}
		}(id)
	}

	ticker := time.NewTicker(40 * time.Second)
	numVotes := 0
	for {
		select {
		case err := <-responseC:
			if err != nil {
				return errors.Wrap(err, 500, errors.E_error_from_partition_peer)
			}
			numVotes++
			if numVotes == len(newWorkers) { // successed
				fmt.Printf("SUCCESS %v\n", newPars)
				if err := me.db.Store(newConfig); err != nil {
					return err
				}
				me.config = newConfig
				return nil
			}
		case <-ticker.C:
			return errors.New(500, errors.E_partition_rebalance_timeout)
		}
	}

}

// make sure all other nodes are died
// this protocol assume that nodes are all nodes that survived
func (me *Coor) ChangeWorkers(newWorkers []string) error {

	// partitions map, key is worker's ID, value is partitions number that is assigned for the worker
	partitionM := make(map[string][]int32)
	for workerid, w := range me.config.GetPartitions() {
		partitionM[workerid] = w.GetPartitions()
	}

	if len(partitionM) == 0 {
		allpars := make([]int32, 0)
		for i := int32(0); i < me.config.GetTotalPartitions(); i++ {
			allpars = append(allpars, i)
		}
		partitionM["_"] = allpars
	}

	partitionM = balance(partitionM, newWorkers)
	newPars := make(map[string]*pb.WorkerPartitions)
	for id, pars := range partitionM {
		newPars[id] = &pb.WorkerPartitions{Partitions: pars}
	}

	newConfig := proto.Clone(me.config).(*pb.Configuration)
	newConfig.Partitions = newPars
	// newConfig.Hosts = newHosts
	return me.transition(newConfig)

}
