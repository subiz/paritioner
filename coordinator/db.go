package coordinator

import (
	"github.com/gocql/gocql"
	"github.com/golang/protobuf/proto"
	"github.com/subiz/cassandra"
	"github.com/subiz/errors"
	pb "github.com/subiz/header/partitioner"
)

const (
	keyspace = "partitioner"
	tblConf  = "conf"
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
type DB struct {
	// hold connection to the database
	session *gocql.Session

	// ID of cluster
	cluster string
}

// NewDB creates new DB object for a cluster
// parameters:
//   seeds: contains list of cassandra "host:port"s, used to initially
//     connect to the cluster then the rest of the ring will be automatically
//     discovered.
//   cluster: used to identify the cluster
func NewDB(seeds []string, cluster string) *DB {
	cql := &cassandra.Query{}
	if err := cql.Connect(seeds, keyspace); err != nil {
		panic(err)
	}

	me := &DB{}
	me.session = cql.Session
	return me
}

// Store persists configuration to the database
func (me *DB) Store(conf *pb.Configuration) error {
	confb, err := proto.Marshal(conf)
	if err != nil {
		return errors.Wrap(err, 500, errors.E_proto_marshal_error)
	}

	err = me.session.Query("INSERT INTO "+tblConf+"(cluster, conf) VALUES(?,?)",
		me.cluster, confb).Exec()
	if err != nil {
		return errors.Wrap(err, 500, errors.E_database_error)
	}
	return nil
}

// Load lookups configuration for cluster in the database
// if not found, it returns empty configuration with no error
func (me *DB) Load() (*pb.Configuration, error) {
	confb := make([]byte, 0)
	err := me.session.Query("SELECT conf FROM "+tblConf+" WHERE cluster=?",
		me.cluster).Scan(&confb)
	if err != nil && err.Error() == gocql.ErrNotFound.Error() {
		return &pb.Configuration{}, nil
	}

	conf := &pb.Configuration{}
	if err := proto.Unmarshal(confb, conf); err != nil {
		return nil, errors.Wrap(err, 500, errors.E_proto_marshal_error)
	}
	return conf, nil
}
