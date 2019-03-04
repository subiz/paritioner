package coordinator

import (
	"bytes"
	"github.com/golang/protobuf/proto"
	"github.com/subiz/errors"
	pb "github.com/subiz/partitioner/header"
	"sync"
	"time"
)

type IDB interface {
	// Store persists configuration to a persistent storage
	Store(conf *pb.Configuration) error

	// Load retrives configuration for cluster from a persistent storage
	Load(cluster string) (*pb.Configuration, error)
}

// Coor is a coordinator implementation
type Coor struct {
	*sync.Mutex

	// used to save (load) configuration to (from) a database
	db IDB

	// hold the intermediate state of the cluster
	config *pb.Configuration

	// used to communicate with workers
	workerComm WorkerComm
}

// Workers communator, used to send signal (message) to workers
type WorkerComm interface {
	Prepare(host string, conf *pb.Configuration) error
}

// NewCoordinator creates a new coordinator
func NewCoordinator(cluster string, db IDB, workerComm WorkerComm) *Coor {
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
			Version:         VERSION,
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

// Leave is called by a worker to tell coordinator that its no longer want
// to be a member of the cluster.
// This function causes a transition, so its may block if coordinator is
// in middle of another transition.
// Coordinator before accept the request must notify all old workers and
// waiting for theirs votes. If there is a workers that doesn't give its vote
// on time, coordinator must deny the request.
// If coordinator accepts the request, it must redistribute all partitions
// handled by the leaving worker to its remainding workers.
func (me *Coor) Leave(req *pb.LeaveRequest) error {
	me.Lock()
	defer me.Unlock()

	err := me.validateRequest(req.Version, req.Cluster, req.Term)
	if err != nil {
		return err
	}

	// ignore if not in the cluster
	_, found := me.config.Workers[req.Id]
	if !found {
		return nil
	}

	// partitions map, key is worker's ID, value is partitions number that is assigned for the worker
	partitionMap := make(map[string][]int32)
	// all current worker ids except the leaving one
	newWorkerIds := make([]string, 0)
	for workerid, w := range me.config.GetWorkers() {
		if workerid != req.Id {
			newWorkerIds = append(newWorkerIds, workerid)
		}
		partitionMap[workerid] = w.GetPartitions()
	}

	partitionMap = balance(me.config.TotalPartitions, partitionMap,
		newWorkerIds)

	newWorkers := make(map[string]*pb.WorkerInfo)
	for id, pars := range partitionMap {
		newWorkers[id] = me.config.Workers[id]
		newWorkers[id].Partitions = pars
	}

	newConfig := proto.Clone(me.config).(*pb.Configuration)
	newConfig.Workers = newWorkers
	return me.transition(newConfig, newWorkerIds, "")
}

// Join is called by a worker to tell coordinator that it want to be a
// member of the cluster.
// This function causes a transition, so its may block if coordinator is
// in middle of another transition.
// Coordinator before accept the request must notify all old workers and
// waiting for theirs votes. If there is a workers that doesn't give its vote
// note that, coordinator dont notify the joining worker
// on time, coordinator must deny the request.
// If coordinator accepts the request, it must redistribute the partitions
// evenly across all its workers
func (me *Coor) Join(req *pb.JoinRequest) error {
	me.Lock()
	defer me.Unlock()
	err := me.validateRequest(req.Version, req.Cluster, req.Term)
	if err != nil {
		return err
	}

	// ignore if already
	_, found := me.config.Workers[req.Id]
	if found {
		return nil
	}

	// partitions map, key is worker's ID, value is partitions number that is assigned for the worker
	partitionMap := make(map[string][]int32)
	newWorkerIds := make([]string, 0)
	for workerid, w := range me.config.Workers {
		newWorkerIds = append(newWorkerIds, workerid)
		partitionMap[workerid] = w.Partitions
	}
	newWorkerIds = append(newWorkerIds, req.Id)

	partitionMap = balance(me.config.TotalPartitions, partitionMap,
		newWorkerIds)

	newWorkers := make(map[string]*pb.WorkerInfo)
	for id, pars := range partitionMap {
		newWorkers[id] = me.config.Workers[id]
		if id == req.Id {
			newWorkers[id] = &pb.WorkerInfo{
				Id: id,
				Host: req.Host,
				NotifyHost: req.NotifyHost,
			}
		}
		newWorkers[id].Partitions = pars
	}
	newConfig := proto.Clone(me.config).(*pb.Configuration)
	newConfig.Workers = newWorkers
	return me.transition(newConfig, newWorkerIds, req.Id)
}

// validateRequest make sures that version, cluster and term of the request
// match the coordinator. So we know that the request is compatible, upto data
// and in the right place
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

// GetConfig returns the current configuration of the coordinator
// This function block while the coordinator in middle of a transition
func (me *Coor) GetConfig() *pb.Configuration {
	me.Lock()
	defer me.Unlock()
	return me.config
}

// TODO: should remove newWorkers param since newConf already has it
func (me *Coor) transition(newConf *pb.Configuration, newWorkers []string, noVoter string) error {
	// ignore no change
	newb, _ := proto.Marshal(newConf)
	oldb, _ := proto.Marshal(me.config)
	if bytes.Compare(newb, oldb) == 0 {
		return nil
	}

	newConf.Term, newConf.NextTerm = me.config.NextTerm, me.config.NextTerm+1

	// save next term to database
	me.config.NextTerm++
	if err := me.db.Store(me.config); err != nil {
		return err
	}

	requiredVotes := len(newWorkers)
	if noVoter != "" && requiredVotes > 0 {
		requiredVotes--
	}
	if requiredVotes == 0 {
		if err := me.db.Store(newConf); err != nil {
			return err
		}
		me.config = newConf
		return nil
	}

	responseC := make(chan error)
	for _, id := range newWorkers {
		if id == noVoter {
			continue
		}
		go func(id string) {
			err := me.workerComm.Prepare(newConf.Workers[id].NotifyHost, newConf)
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
			if numVotes == requiredVotes { // successed
				// fmt.Printf("SUCCESS %v\n", newPars)
				if err := me.db.Store(newConf); err != nil {
					return err
				}
				me.config = newConf
				return nil
			}
		case <-ticker.C:
			return errors.New(500, errors.E_partition_rebalance_timeout)
		}
	}
}

const VERSION = "1.0.0"
