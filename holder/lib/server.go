package holder

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Server struct {
	parent *Holder
}

func NewServer(parent *Holder) *Server {

	s := new(Server)
	s.parent = parent

	return s

}

func (s *Server) Start() {

	server := http.NewServeMux()

	server.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {

		fmt.Fprintln(res, "<html><body>Hello world!</body></html>")

	})

	server.HandleFunc("/api/shards", func(res http.ResponseWriter, req *http.Request) {

		data, err := json.MarshalIndent(s.parent.shards, "", "  ")

		if err != nil {
			s.parent.logger.WithFields(logrus.Fields{
				"scope": "server/handler",
			}).Warn(err)
			res.Write([]byte("500 Internal Server Error"))
			return
		}

		res.Write(data)

	})

	server.HandleFunc("/api/clients", func(res http.ResponseWriter, req *http.Request) {

		data, err := json.MarshalIndent(s.parent.clients, "", "  ")

		if err != nil {
			s.parent.logger.WithFields(logrus.Fields{
				"scope": "server/handler",
			}).Warn(err)
			res.Write([]byte("500 Internal Server Error"))
			return
		}

		res.Write(data)

	})

	go http.ListenAndServe("0.0.0.0:41337", server)

}
