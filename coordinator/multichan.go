package main

import (
	"github.com/subiz/errors"
	"sync"
	"time"
)

type MultiChan struct {
	*sync.Mutex
	chans map[string]chan interface{}
}

func (me *MultiChan) Send(chanid string, msg interface{}) {
	me.Lock()

	ch, ok := me.chans[chanid]
	if !ok { // no election to vote
		me.Unlock()
		return
	}

	me.Unlock()

	// send in channel with timeout, avoiding gorotine wasting
	// we have to force close the channel in order to break send operation
	done := make(chan bool)
	go safe(func() {
		ch <- msg
		done <- true
	})
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		safe(func() {
			close(ch)
			close(done)
		})
	}
}

func NewMultiChan() *MultiChan {
	return &MultiChan{Mutex: &sync.Mutex{}, chans: make(map[string]chan interface{})}
}

func (me *MultiChan) Recv(chanid string) (interface{}, error) {
	me.Lock()
	c, ok := me.chans[chanid]
	if !ok {
		c = make(chan interface{})
		me.chans[chanid] = c
		me.Unlock()
		// return nil, errors.New(500, errors.E_unknown) //error.E_duplicated_partition_term, "call wait for vote twice")
	}

	defer func() {
		// clear resource when done, since chan are designed to use one time
		// only
		me.Lock()
		safe(func() { close(c) })
		delete(me.chans, chanid)
		me.Unlock()
	}()

	ticker := time.NewTicker(40 * time.Second)
	select {
	case msg, more := <-c:
		if !more {
			return nil, errors.New(500, errors.E_partition_rebalance_timeout)
		}
		return msg, nil
	case <-ticker.C:
		return nil, errors.New(500, errors.E_partition_rebalance_timeout)
	}
}
