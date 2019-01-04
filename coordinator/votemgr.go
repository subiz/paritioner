package main

import (
	"fmt"
	"github.com/subiz/errors"
	"sync"
	"time"
)

type VoteMgr struct {
	*sync.Mutex
	elections map[string]chan bool
}

func makeElectionId(workerid string, term int32) string {
	return fmt.Sprintf("%s|%d", workerid, term)
}

func (me *VoteMgr) Vote(workerid string, term int32, accept bool) {
	me.Lock()
	eId := makeElectionId(workerid, term)
	vote, ok := me.elections[eId]
	if !ok { // no election to vote
		me.Unlock()
		return
	}

	me.Unlock()

	// send in channel with timeout, avoiding gorotine wasting
	// we have to force close the channel in order to break send operation
	done := make(chan bool)
	go safe(func() {
		vote <- accept
		done <- true
	})
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		safe(func() {
			close(vote)
			close(done)
		})
	}
}

func NewVoteMgr() *VoteMgr {
	return &VoteMgr{Mutex: &sync.Mutex{}, elections: make(map[string]chan bool)}
}

func (me *VoteMgr) Wait(workerid string, term int32) error {
	me.Lock()
	eId := makeElectionId(workerid, term)
	if _, ok := me.elections[eId]; ok {
		return errors.New(500, errors.E_unknown) //error.E_duplicated_partition_term, "call wait for vote twice")
	}
	c := make(chan bool)
	me.elections[eId] = c
	me.Unlock()

	ticker := time.NewTicker(40 * time.Second)
	for {
		select {
		case accepted, more := <-c:
			if !more {
				return errors.New(500, errors.E_partition_rebalance_timeout)
			}

			if accepted {
				return nil
			}
			return errors.New(500, errors.E_partition_rebalance_timeout, "worker donot accept")
		case <-ticker.C:
			return errors.New(500, errors.E_partition_rebalance_timeout)
		}
	}
}
