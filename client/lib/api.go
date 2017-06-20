package client

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

type API struct {
	parent *Client
}

func NewAPI(parent *Client) *API {

	a := new(API)

	a.parent = parent

	return a

}

func (a *API) Start() {

	server := http.NewServeMux()

	server.HandleFunc("/api/status", func(res http.ResponseWriter, req *http.Request) {
		data := a.parent.GetStatus()
		a.json(res, req, data)
	})

	server.HandleFunc("/api/shards", func(res http.ResponseWriter, req *http.Request) {
		shards := a.parent.UpdateShards()
		a.json(res, req, shards)
	})

	server.HandleFunc("/api/shards/relay", func(res http.ResponseWriter, req *http.Request) {

		shardAddr := req.FormValue("shardAddr")
		transport := req.FormValue("transport")

		hostAddr, guestAddr := a.parent.RequestRelay(shardAddr, transport)

		data := make(map[string]interface{})
		data["hostAddr"] = hostAddr
		data["guestAddr"] = guestAddr

		a.json(res, req, data)

	})

	server.HandleFunc("/api/connectionStatus", func(res http.ResponseWriter, req *http.Request) {
		data := a.parent.GetConnectionStatus()
		a.json(res, req, data)
	})

	go (func() {

		err := http.ListenAndServe("127.0.0.1:61337", server)

		if err != nil {
			a.parent.logger.WithFields(logrus.Fields{
				"scope": "api/Start",
			}).Fatal(err)
		}

	})()

}

func (a *API) json(res http.ResponseWriter, req *http.Request, data interface{}) {

	buf, err := json.Marshal(data)

	if err != nil {
		a.parent.logger.WithFields(logrus.Fields{
			"scope": "api/json",
		}).Fatal(err)
	}

	res.Write(buf)

}
