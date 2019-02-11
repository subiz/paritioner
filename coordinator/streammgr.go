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
// for each worker, user uses Pull method to maintain the pulling.
// after worker leave, user should call Remove to clean up the resource
type StreamMgr struct {
	*sync.Mutex

	// holds all connections inside a map
	// used to lookup connection details using worker's ID
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
}

// NewStreamMgr creates a new StreamMgr object
func NewStreamMgr() *StreamMgr {
	return &StreamMgr{Mutex: &sync.Mutex{}, pulls: make(map[string]workerConn)}
}

// Pull adds new connection to manager.
// This method BLOCKS THE EXECUTION and keep the stream open until user calls
// another Pull with the same workerid. This represenses a single long polling
// request for each worker id
// User can creates as many connections for a worker as they want but the
// manager only hold the last one. It makes sure that in anytime, there are
// ATMOST one long pulling request for a worker.
// The execution can also be unblock by calling Remove
// pull exits when:
// - stream broken
// - client cancelled
// This method is thread-safe
func (me *StreamMgr) Pull(workerid string, stream Stream) {
	me.Lock()

	// drop the last long pulling request if existed
	if conn, ok := me.pulls[workerid]; ok {
		// we blocked the last pulling by receiving from an empty channel.
		// here we unblock the execution (so the pulling can die)
		conn.exitc <- true
	}

	c := make(chan bool)
	me.pulls[workerid] = workerConn{
		conn_id:     randstr.Hex(16),
		worker_id:   workerid,
		stream:      stream,
		exitc:       c,
	}
	me.Unlock()

	<-c // block
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

	// connection is outdated, some one has already closed the connection
	if oldconn.conn_id != conn.conn_id {
		return
	}

	delete(me.pulls, conn.worker_id)
	// sending to exitc to unblock the execution, releasing the poll
	conn.exitc <- true
}

// Send sends msg to worker's stream
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
