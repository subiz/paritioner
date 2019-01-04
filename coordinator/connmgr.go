package main

import (
	"github.com/thanhpk/randstr"
	"sync"
)

// ConnMgr manages worker long polling connections
// for each worker, user uses Pull method to maintaing the pulling
// after worker leave, user should call Remove to clean up the resource
type ConnMgr struct {
	*sync.Mutex

	// holds all connections inside a map
	// key of the map is ID of worker, value of the map is the connection
	// details
	pulls map[string]WorkerConn
}

// WorkerConn represences a connection made from worker to coordinator
type WorkerConn struct {
	// Payload is used to hold arbitrary data which user want to attach
	// to this worker
	Payload interface{}

	// an unique string to differentiate  multiple connections of a worker
	conn_id string

	// ID of the worker
	worker_id string

	// a channel to remotely unhalt the blocking worker.Pull method
	exitc chan bool
}

// NewConnMgr creates a new ConnMgr object
func NewConnMgr() *ConnMgr {
	return &ConnMgr{Mutex: &sync.Mutex{}, pulls: make(map[string]WorkerConn)}
}

// Pull adds new connection to manager.
// This method BLOCKS THE EXECUTION to keeping the streaming open
// until user calls another Pull with
// the same workerid. This represenses a single long polling request for each
// worker id
// User can creates many connections for a worker as they want but the manager
// only keep the last one. It makes sure that in anytime, there are atmost one
// long pulling request for a worker.
// The execution can also be unblock by calling Remove
func (me *ConnMgr) Pull(workerid string, payload interface{}) {
	me.Lock()

	// droping the last long pulling request if existed
	if conn, ok := me.pulls[workerid]; ok {
		// we blocking the execution by makeing it receive from an empty channel.
		// By closing the channel, we unblock the execution (so the pulling can die)
		safe(func() { close(conn.exitc) })
		delete(me.pulls, workerid)
	}

	c := make(chan bool)
	me.pulls[workerid] = WorkerConn{
		conn_id:   randstr.Hex(16),
		worker_id: workerid,
		Payload:   payload,
		exitc:     c,
	}
	me.Unlock()
	// block until channel is closed
	// (no one is going to send to this channel)
	<-c
}

// Get lookups worker connection details by ID
func (me *ConnMgr) Get(workerid string) (conn WorkerConn, ok bool) {
	me.Lock()
	defer me.Unlock()
	c, ok := me.pulls[workerid]
	return c, ok
}

// Remove drops worker connection and it's pulling
// worker can make multiples connections, but the manager only keep
// the last one and dropped all others connections. Any attempt to
// delete previous connection is considered outdated and will be ignore
func (me *ConnMgr) Remove(conn WorkerConn) {
	me.Lock()
	defer me.Unlock()

	oldconn, ok := me.pulls[conn.worker_id]
	if !ok { // already deleted
		return
	}

	if oldconn.conn_id != conn.conn_id {
		// connection is outdated, some one has already closed the connection
		return
	}

	// closing exitc to unblock the execution, let the pulling die
	safe(func() { close(conn.exitc) })
	delete(me.pulls, conn.worker_id)
}
