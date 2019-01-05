package main

import (
	"github.com/subiz/errors"
	"sync"
	"time"
)

type MultiChan struct {
	*sync.Mutex
	chans        map[string]chan interface{}
	lastModifies map[string]int64
}

func NewMultiChan() *MultiChan {
	m := &MultiChan{
		Mutex:        &sync.Mutex{},
		chans:        make(map[string]chan interface{}),
		lastModifies: make(map[string]int64),
	}
	go m.cleanLoop(5 * time.Minute)
	return m
}

func (me *MultiChan) Send(chanid string, msg interface{}, timeout time.Duration) error {
	me.Lock()
	ch, ok := me.chans[chanid]
	if !ok {
		ch = make(chan interface{})
		me.chans[chanid] = ch
	}
	me.lastModifies[chanid] = time.Now().Unix()
	me.Unlock()

	if timeout <= 0 {
		timeout = 1 * time.Millisecond
	}
	select {
	case ch <- msg:
		return nil
	case <-time.After(timeout):
		return errors.New(500, errors.E_send_to_channel_timeout, chanid)
	}
}

func (me *MultiChan) Recv(chanid string, timeout time.Duration) (interface{}, error) {
	me.Lock()
	ch, ok := me.chans[chanid]
	if !ok {
		ch = make(chan interface{})
		me.chans[chanid] = ch
	}
	me.lastModifies[chanid] = time.Now().Unix()
	me.Unlock()

	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	select {
	case msg := <-ch:
		return msg, nil
	case <-time.After(timeout):
		return nil, errors.New(500, errors.E_partition_rebalance_timeout)
	}
}

func (me *MultiChan) cleanLoop(clientinterval time.Duration) {
	for {
		me.Lock()
		for chanid, lm := range me.lastModifies {
			if time.Since(time.Unix(lm, 0)) < 3*time.Minute {
				continue
			}

			delete(me.lastModifies, chanid)
			delete(me.chans, chanid)
		}
		me.Unlock()
		time.Sleep(clientinterval)
	}
}
