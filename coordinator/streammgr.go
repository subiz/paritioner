package main

import (
	"github.com/subiz/errors"
	"github.com/thanhpk/randstr"
	"sync"
)

type Stream interface {
	SendMsg(m interface{}) error
}

// StreamMgr manages worker long polling connections
// for each worker, user uses Pull method to maintaing the pulling
// after worker leave, user should call Remove to clean up the resource
type StreamMgr struct {
	*sync.Mutex

	// holds all connections inside a map
	// key of the map is ID of worker, value of the map is the connection
	// details
	pulls map[string]workerConn
}

// WorkerConn represences a connection made from worker to coordinator
type workerConn struct {
	// Payload is used to hold arbitrary data which user want to attach
	// to this worker
	stream Stream

	// an unique string to differentiate  multiple connections of a worker
	conn_id string

	// ID of the worker
	worker_id string

	// a channel to remotely unhalt the blocking worker.Pull method
	exitc chan bool

	pingPayload interface{}
}

// NewStreamMgr creates a new StreamMgr object
func NewStreamMgr() *StreamMgr {
	return &StreamMgr{Mutex: &sync.Mutex{}, pulls: make(map[string]workerConn)}
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
// pull exits when:
// - stream broke
// - client canceled
func (me *StreamMgr) Pull(workerid string, stream Stream, pingPayload interface{}) {
	me.Lock()

	// droping the last long pulling request if existed
	if conn, ok := me.pulls[workerid]; ok {
		// we blocking the execution by makeing it receive from an empty channel.
		// By closing the channel, we unblock the execution (so the pulling can die)
		safe(func() { close(conn.exitc) })
		delete(me.pulls, workerid)
	}

	c := make(chan bool)
	me.pulls[workerid] = workerConn{
		conn_id:     randstr.Hex(16),
		worker_id:   workerid,
		stream:      stream,
		exitc:       c,
		pingPayload: pingPayload,
	}
	me.Unlock()
	// block until channel is closed
	// (no one is going to send to this channel)
	<-c
}

// Remove drops worker connection and it's pulling
// worker can make multiples connections, but the manager only keep
// the last one and dropped all others connections. Any attempt to
// delete previous connection is considered outdated and will be ignore
func (me *StreamMgr) remove(conn workerConn) {
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

func (me *StreamMgr) Send(workerid string, msg interface{}) error {
	me.Lock()
	conn, ok := me.pulls[workerid]
	me.Unlock()

	if !ok {
		return errors.New(500, errors.E_partition_node_have_not_joined_the_cluster,
			"worker stream not found", workerid)
	}

	if err := conn.stream.SendMsg(msg); err != nil {
		me.remove(conn)
		return err
	}
	return nil
}

func safe(f func()) {
	func() {
		defer func() { recover() }()
		f()
	}()
}
