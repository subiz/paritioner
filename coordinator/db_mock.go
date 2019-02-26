package main

import (
	pb "github.com/subiz/partitioner/header"
	"sync"
)

// DBMock is a mock of DB struct for testing purpose only
type DBMock struct {
	*sync.Mutex
	store map[string]pb.Configuration
}

// NewDBMock creates a new DBMock object
func NewDBMock() *DBMock {
	return &DBMock{
		Mutex: &sync.Mutex{},
		store: make(map[string]pb.Configuration),
	}
}

// Load lookups configuration for cluster in the database
func (me *DBMock) Load(cluster string) (*pb.Configuration, error) {
	me.Lock()
	defer me.Unlock()
	conf := me.store[cluster]
	return &conf, nil
}

// Store persists configuration to the database
func (me *DBMock) Store(conf *pb.Configuration) error {
	me.Lock()
	defer me.Unlock()
	me.store[conf.Cluster] = *conf
	return nil
}
