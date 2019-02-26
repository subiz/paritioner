package main

import (
	pb "github.com/subiz/partitioner/header"
)

type DBMock struct {
	store map[string]pb.Configuration
}

func NewDBMock() *DBMock {
	return &DBMock{store: make(map[string]pb.Configuration)}
}

func (me *DBMock) Load(cluster string) (*pb.Configuration, error) {
	conf := me.store[cluster]
	return &conf, nil
}

func (me *DBMock) Store(conf *pb.Configuration) error {
	me.store[conf.Cluster] = *conf
	return nil
}
