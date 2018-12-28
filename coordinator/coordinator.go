package coordinator

import (
	"fmt"
	"github.com/subiz/errors"
	pb "github.com/subiz/header/partitioner"
	"sync"
	"time"
)

// Coor is a coordinator implementation
type Coor struct {
	*sync.Mutex
	db *DB

	Config    *pb.Configuration // intermediate configuration
	Version   string
	Cluster   string
	Term      int32
	workerMgr WorkerMgr
}

type WorkerMgr interface {
	Request(old, new *pb.WorkerConfiguration) bool
	Abort(w *pb.WorkerConfiguration)
	Apply(w *pb.WorkerConfiguration)
}

// Run starts the coordinator
func (me *Coor) Run(db *DB) {
	conf, err := me.db.Load()
	if err != nil {
		panic(err)
	}
	me.Config = conf
}

func (me *Coor) validateRequest(version, cluster string, term int32) error {
	if version != me.Version {
		return errors.New(400, errors.E_invalid_version, "only support version "+me.Version)
	}

	if cluster != me.Cluster {
		return errors.New(400, errors.E_invalid_cluster, "cluster should be "+me.Cluster+" not "+cluster)
	}

	if term != me.Term {
		return errors.New(400, errors.E_invalid_term, "term should be %d, not %d", me.Term, term)
	}
	return nil
}

func (me *Coor) Join(join *pb.JoinClusterRequest) error {
	me.Lock()
	defer me.Unlock()

	err := me.validateRequest(join.GetVersion(), join.GetCluster(), join.GetTerm())
	if err != nil {
		return err
	}

	found := false
	for _, w := range me.Config.GetWorkers() {
		if w.Id != join.GetId() {
			continue
		}
		found = true
		w.Host = join.Host
	}

	if found {
		return me.db.Store(me.Config)
	}
	return nil
}

func (me *Coor) Change(nodes []string) {
	me.Lock()
	defer me.Unlock()

	// balance(workers, nodes)
	newConfigs := me.Config.GetWorkers()

	responseC := make(chan bool)
	for i, w := range me.Config.GetWorkers() {
		go func(newConfig *pb.WorkerConfiguration) {
			responseC <- me.workerMgr.Request(w, newConfig)
		}(newConfigs[i])
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
			if total == len(me.Config.GetWorkers()) {
				// success
				for _, w := range newConfigs {
					me.workerMgr.Apply(w)
				}
				return
			}
		case <-ticker.C:
			fmt.Printf("TIMEOUT, got only %d respones", total)
			goto failed
		}
	}

	failed:
	for _, w := range me.Config.GetWorkers() {
		me.workerMgr.Abort(w)
	}
}
