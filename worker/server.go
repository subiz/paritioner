package coordinator

import (
	"time"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/subiz/errors"
	shttp "github.com/subiz/goutils/http"
	pb "github.com/subiz/header/partitioner"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	coor        *Coor
	kubeservice string
}

func RunServe(cluster, kubeservice string, cassandraSeeds []string, port string) error {
	me := &Server{}
	me.coor = &Coor{}
	me.kubeservice = kubeservice
	me.coor.Run(NewDB(cassandraSeeds, cluster))

	http.HandleFunc("/request", me.handleRequest)
	http.HandleFunc("/commit", me.handleCommit)
	http.HandleFunc("/abort", me.handleAbort)
	http.HandleFunc("/config", me.handleGetConfig)
	return http.ListenAndServe(":"+port, nil)
}

// queryParam lookups parameter value in http url query by its name
// if not found, return default string
// if the query is an array of values, it only returns the first element
func queryParam(r *http.Request, name, defaultValue string) string {
	values, ok := r.URL.Query()[name]
	if !ok || len(values[0]) < 1 {
		return defaultValue
	}
	return values[0]
}

func containsType(header string, typ string) bool {
	formats := strings.Split(header, ",")
	typ = strings.TrimSpace(typ)
	for _, f := range formats {
		// ignore ;q= traling, see
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept
		f := strings.TrimSpace(strings.Split(f, ";")[0])

		if f == typ {
			return true
		}
	}
	return false
}

type Message interface {
	Reset()
	String() string
	ProtoMessage()
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

func reply(w http.ResponseWriter, r *http.Request, code int, msg Message) {
	accept := r.Header.Get("Accept")
	w.WriteHeader(200)

	if containsType(accept, "application/protobuf") {
		w.Header().Set("Content-Type", "application/protobuf")
		b, _ := proto.Marshal(msg)
		w.Write(b)
		return
	}

	// default
	w.Header().Set("Content-Type", "application/json")
	b, _ := msg.MarshalJSON()
	w.Write(b)
}

func (me *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	reply(w, r, 200, me.coor.Config)
}

func (me *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	join := &pb.JoinClusterRequest{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(join); err != nil {
		reply(w, r, 400, &pb.Event{
			Version:      me.coor.Version,
			Term:         me.coor.Term,
			Cluster:      me.coor.Cluster,
			Type:         pb.Event_ERROR.String(),
			ErrorCode:    errors.E_json_marshal_error.String(),
			ErrorMessage: "request body should be a valid json",
		})
		return
	}

	if err := me.coor.Join(join); err != nil {
		e := errors.Wrap(err, 500, errors.E_unknown).(*errors.Error)
		reply(w, r, 400, &pb.Event{
			Version:      me.coor.Version,
			Term:         me.coor.Term,
			Cluster:      me.coor.Cluster,
			Type:         pb.Event_ERROR.String(),
			ErrorCode:    e.Code,
			ErrorMessage: e.Description,
		})
		return
	}

	reply(w, r, 200, nil)
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

		for _, wk := range me.coor.Config.GetWorkers() {
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

func (me *Server) Request(old, new *pb.WorkerConfiguration) bool {
	b, _ := new.MarshalJSON()
	body, err := shttp.Request("POST", old.GetHost()+"/block", b, &shttp.Config{
		Header:  map[string]string{"Content-Type": "application/json"},
		Timeout: 10 * time.Second,
	})
	return err != nil
}

func (me *Server) Abort(w *pb.WorkerConfiguration) {
	shttp.Request("POST", w.GetHost()+"/abort", nil, &shttp.Config{
		Header:  map[string]string{"Content-Type": "application/json"},
		Timeout: 10 * time.Second,
	})
}

func (me *Server) Apply(w *pb.WorkerConfiguration) {
	shttp.Request("POST", w.GetHost()+"/apply", nil, &shttp.Config{
		Header:  map[string]string{"Content-Type": "application/json"},
		Timeout: 10 * time.Second,
	})
}
