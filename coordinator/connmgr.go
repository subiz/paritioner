package main

import (
	"github.com/thanhpk/randstr"
	"sync"
)

// ConnMgr is used to manage connections
type ConnMgr struct {
	*sync.Mutex
	workerPulls map[string]WorkerConn
}

type WorkerConn struct {
	Payload   interface{}
	conn_id   string
	worker_id string
	exitc     chan error
}

func NewConnMgr() *ConnMgr {
	return &ConnMgr{Mutex: &sync.Mutex{}, workerPulls: make(map[string]WorkerConn)}
}

func (me *ConnMgr) Pull(workerid string, payload interface{}) {
	me.Lock()
	if conn, ok := me.workerPulls[workerid]; ok {
		safe(func() { close(conn.exitc) })
		delete(me.workerPulls, workerid)
	}

	c := make(chan error)
	me.workerPulls[workerid] = WorkerConn{
		conn_id:   randstr.Hex(16),
		worker_id: workerid,
		Payload:   payload,
		exitc:     c,
	}
	me.Unlock()
	<-c
}

func (me *ConnMgr) Get(workerid string) (conn WorkerConn, ok bool) {
	me.Lock()
	defer me.Unlock()
	c, ok := me.workerPulls[workerid]
	return c, ok
}

func (me *ConnMgr) Remove(conn WorkerConn) {
	me.Lock()

	me.Unlock()

	oldconn, ok := me.workerPulls[conn.worker_id]
	if !ok { // already deleted
		return
	}

	if oldconn.conn_id == conn.conn_id {
		delete(me.workerPulls, conn.worker_id)
	}
}
