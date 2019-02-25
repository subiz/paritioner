package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/golang/protobuf/proto"
	"github.com/subiz/errors"
	pb "github.com/subiz/partitioner/header"
	"time"
)

const (
	keyspace = "partitioner"
	tblConf  = "conf"
)

// DB allows the cluster to restore its state after crash by persisting the
// state to a Cassandra cluster.
// Each DB object holds a reference to one cluster only. If you want multiple
// clusters, one should create multiple instances of DB
//
// before use this struct, user must initialize cassandra keyspace and tables by
// running following CQL commands in Cassandra cqlsh tool
//   CREATE KEYSPACE IF NOT EXISTS partition WITH replication = {
//     'class': 'SimpleStrategy',
//     'replication_factor': 3
//   };
//
//   CREATE TABLE partitioner.conf(cluster TEXT,conf BLOB,PRIMARY KEY(cluster));
type DB struct {
	// hold connection to the database
	session *gocql.Session
}

// connect establish a new session to a Cassandra cluster, this function will
// retry on error automatically, and blocking until a session successfully
// established
func connect(seeds []string, keyspace string) *gocql.Session {
	cluster := gocql.NewCluster(seeds...)
	cluster.Timeout = 10 * time.Second
	cluster.Keyspace = keyspace
	for {
		session, err := cluster.CreateSession()
		if err == nil {
			return session
		}
		fmt.Println("cassandra", err, ". Retring after 5sec...")
		time.Sleep(5 * time.Second)
	}
}

// NewDB creates new DB object for a cluster
// parameters:
//   seeds: contains list of cassandra "host:port"s, used to initially
//     connect to the cluster then the rest of the ring will be automatically
//     discovered.
//   cluster: used to identify the cluster
func NewDB(seeds []string) *DB {
	me := &DB{}
	me.session = connect(seeds, keyspace)
	return me
}

// Store persists configuration to the database
func (me *DB) Store(conf *pb.Configuration) error {
	b, _ := json.Marshal(conf)
	println("STORING", string(b))
	confb, err := proto.Marshal(conf)
	if err != nil {
		return errors.Wrap(err, 500, errors.E_proto_marshal_error)
	}

	err = me.session.Query("INSERT INTO "+tblConf+"(cluster, conf) VALUES(?,?)",
		conf.GetCluster(), confb).Exec()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error)
	}
	return nil
}

// Load lookups configuration for cluster in the database
// if not found, it returns empty configuration with no error
func (me *DB) Load(cluster string) (*pb.Configuration, error) {
	confb := make([]byte, 0)
	err := me.session.Query("SELECT conf FROM "+tblConf+" WHERE cluster=?",
		cluster).Scan(&confb)
	if err != nil && err.Error() == gocql.ErrNotFound.Error() {
		return &pb.Configuration{}, nil
	}

	if err != nil {
		return nil, errors.Wrap(err, 500, errors.E_database_error)
	}

	conf := &pb.Configuration{}
	if err := proto.Unmarshal(confb, conf); err != nil {
		return nil, errors.Wrap(err, 500, errors.E_proto_marshal_error)
	}
	conf.Cluster = cluster
	return conf, nil
}
