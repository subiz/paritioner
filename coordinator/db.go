package main

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/golang/protobuf/proto"
	"github.com/subiz/errors"
	pb "github.com/subiz/header/partitioner"
	"time"
)

const (
	keyspace = "partitioner"
	tblConf  = "conf"
	tblHost  = "host"
)

// DB allows cluster to reload its state after crash by persisting the state to
// Cassandra database.
// each DB object with hold reference to one cluster only. If one want multiple
// clusters, one should create multiple instances of DB
//
// befor use this struct, user must initialize cassandra keyspace and tables by
// running following CQL commands in Cassandra cqlsh tool
//   CREATE KEYSPACE IF NOT EXISTS partition WITH replication = {
//     'class': 'SimpleStrategy',
//     'replication_factor': 3
//   };
//
//   CREATE TABLE partitioner.conf(cluster TEXT,conf BLOB,PRIMARY KEY(cluster));
//   CREATE TABLE partitioner.host(cluster TEXT,id ASCII, host ASCII, PRIMARY KEY(cluster, id));
type DB struct {
	// hold connection to the database
	session *gocql.Session
}

// connect creates new session to cassandra database, this function auto
// retry on error, it blocks until connection is successfully made
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
	return conf, nil
}

// SaveHost persists pair <worker ID and worker host> to the database.
// After this, user can use LoadHosts to lookup host
func (me *DB) SaveHost(cluster, id, host string) error {
	err := me.session.Query("INSERT INTO "+tblHost+"(cluster, id, host) VALUES(?,?,?)",
		cluster, id, host).Exec()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error)
	}
	return nil
}

// RemoveHost deletes worker's host by it's ID
func (me *DB) RemoveHost(cluster, id string) error {
	err := me.session.Query("DELETE FROM "+tblHost+" WHERE cluster=? AND id=?",
		cluster, id).Exec()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error)
	}
	return nil
}

// LoadHosts lookups all worker hosts and IDs by cluster name
func (me *DB) LoadHosts(cluster string) (map[string]string, error) {
	hosts := make(map[string]string)
	iter := me.session.Query("SELECT id,host FROM "+tblHost+" WHERE cluster=?",
		cluster).Iter()
	id, host := "", ""
	for iter.Scan(&id, &host) {
		hosts[id] = host
	}
	if err := iter.Close(); err != nil {
		return nil, errors.Wrap(err, 500, errors.E_database_error, cluster, id)
	}
	return hosts, nil
}
