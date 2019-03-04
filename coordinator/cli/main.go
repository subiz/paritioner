package main

import (
	"github.com/subiz/partitioner/coordinator"
	"github.com/kelseyhightower/envconfig"
	"github.com/urfave/cli"
)

type Config struct {
	CassandraSeeds []string `required:"true"`
	Port           string   `required:"true"`
	Services       []string `required:"true"`
	CassandraUser  string   `required:"true"`
	CassandraPass  string   `required:"true"`
}

var cf Config

func main() {
	envconfig.MustProcess("coor", &cf)
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{Name: "daemon", Usage: "run server", Action: daemon},
	}
	app.RunAndExitOnError()
}

// daemon loads all clusters and start GRPC server
func daemon(ctx *cli.Context) {
	coordinator.RunCoordinator(cf.CassandraSeeds, cf.Services, cf.Port,
		cf.CassandraUser, cf.CassandraPass)
}
