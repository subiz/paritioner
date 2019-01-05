package main

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/urfave/cli"
)

type Config struct {
	CassandraSeeds []string `required:"true"`
	Port           string   `required:"true"`
	Services       []string `required:"true"`
}

var c Config

func main() {
	envconfig.MustProcess("coor", &c)
	app := cli.NewApp()
	app.Name = "coordinator"

	app.Action = func(c *cli.Context) error {
		fmt.Println("hello from partition coordinator")
		return nil
	}

	app.Commands = []cli.Command{
		{Name: "daemon", Usage: "run server", Action: daemon},
	}
	app.RunAndExitOnError()
}
