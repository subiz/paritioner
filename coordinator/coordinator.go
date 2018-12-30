package coordinator

import (
	"context"
	"fmt"
	"sort"
	// "github.com/subiz/errors"
	pb "github.com/subiz/header/partitioner"
	"sync"
	"time"
)

// Coor is a coordinator implementation
type Coor struct {
	*sync.Mutex
	db      *DB
	config  *pb.Configuration // intermediate configuration
	version string
	cluster string
	worker  Worker
}

type Worker interface {
	Request(*pb.WorkerConfiguration) bool
}

func NewCoordinator(cluster string, db *DB) *Coor {
	me := &Coor{Mutex: &sync.Mutex{}}
	me.version = "1.0.0"
	me.cluster = cluster
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

	me.config.Term = me.config.GetTerm() + 1

	if found {
		if err := me.db.Store(me.config); err != nil {
			return nil, err
		}
	}
	return me.config, nil
}

func (me *Coor) Change(nodes []string) {
	me.Lock()
	defer me.Unlock()

	// balance(workers, nodes)
	newConfigs := me.config.GetWorkers()

	responseC := make(chan bool)
	for i, _ := range me.config.GetWorkers() {
		go func(newConfig *pb.WorkerConfiguration) {
			//responseC <- me.worker.Request(w, newConfig)
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

type elem struct {
	id      string
	pars    []int
	numPars int
}

type byLength []elem

func (s byLength) Len() int      { return len(s) }
func (s byLength) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byLength) Less(i, j int) bool {
	if len(s[i].pars) == len(s[j].pars) {
		return s[i].id < s[j].id
	}
	return len(s[i].pars) > len(s[j].pars)
}

func isIn(A []string, s string) bool {
	for _, a := range A {
		if a == s {
			return true
		}
	}
	return false
}
func balance(partitions map[string][]int, nodes []string) map[string][]int {
	if len(nodes) == 0 {
		return nil
	}

	elems := make([]elem, len(partitions))
	i := 0
	for id, pars := range partitions {
		elems[i] = elem{id: id, pars: pars}
		i++
	}

	for _, n := range nodes {
		found := false
		for _, e := range elems {
			if e.id == n {
				found = true
				break
			}
		}

		if !found {
			elems = append(elems, elem{id:n})
		}
	}

	sort.Sort(byLength(elems))

	// count total of partition
	totalPars := 0
	for _, pars := range partitions {
		totalPars += len(pars)
	}

	mod := totalPars % len(nodes)
	numWorkerPars := totalPars / len(nodes)
	i = 0
	for k, e := range elems {
		if !isIn(nodes, e.id) {
			continue
		}
		e.numPars = numWorkerPars
		if i < mod {
			e.numPars++
		}
		i++
		elems[k] = e
	}

	redurantPars := make([]int, 0)
	for k, e := range elems {
		if len(e.pars) > e.numPars { // have redurant job
			redurantPars = append(redurantPars, e.pars[e.numPars:]...)
			e.pars = e.pars[:e.numPars]
			elems[k] = e
		}
	}

	for k, e := range elems {
		if len(e.pars) < e.numPars { // have redurant job
			lenepars := len(e.pars)
			e.pars = append(e.pars, redurantPars[:e.numPars-lenepars]...)

			redurantPars = redurantPars[e.numPars-lenepars:]
			elems[k] = e
		}
	}

	partitions = make(map[string][]int)
	for _, e := range elems {
		if len(e.pars) > 0 {
			partitions[e.id] = e.pars
		}
	}
	return partitions
}
