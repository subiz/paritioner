package coordinator

import (
	"fmt"
	pb "github.com/subiz/header/partitioner"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	coor        *Coor
	kubeservice string
}

func RunServe(cluster, kubeservice string, cassandraSeeds []string, port string) error {
	me := &Server{}
	me.kubeservice = kubeservice
	me.coor = NewCoordinator(cluster, NewDB(cassandraSeeds, cluster))

	grpcServer := grpc.NewServer()
	pb.RegisterCoordinatorServer(grpcServer, me.coor)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	return grpcServer.Serve(lis)
}



func (me *Server) lookupDNS() {
	for {
		changed := false
		nodes := make([]string, 0)

		ips, err := net.LookupIP(me.kubeservice)
		if err != nil {
			fmt.Printf("Could not get IPs: %v\n", err)
			goto end
		}

		for _, wk := range me.coor.config.GetWorkers() {
			wssplits := strings.Split(wk.GetId(), "-")
			if len(wssplits) < 2 {
				break
			}

			n, err := strconv.Atoi(wssplits[len(wssplits)-1])
			if err != nil {
				break
			}
			if n < len(ips) {
				changed = true
				nodes = append(nodes, wk.GetId())
			}
		}

		if changed {
			me.coor.Change(nodes)
		}

	end:
		time.Sleep(2 * time.Second)
	}
}

func (me *Server) Request(n *pb.WorkerConfiguration) bool {
	return true
}
