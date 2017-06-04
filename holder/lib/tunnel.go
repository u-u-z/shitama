package holder

import (
	"encoding/gob"

	"github.com/asaskevich/EventBus"
	"github.com/hashicorp/yamux"
	"github.com/sirupsen/logrus"
	kcp "github.com/xtaci/kcp-go"
)

type Tunnel struct {
	OnConnected        EventBus.Bus
	OnDisconnected     EventBus.Bus
	OnRequestReceived  EventBus.Bus
	OnPublishReceived  EventBus.Bus
	OnPeerConnected    EventBus.Bus
	OnPeerDisconnected EventBus.Bus
	parent             *Holder
}

func NewTunnel(parent *Holder) *Tunnel {

	t := new(Tunnel)
	t.parent = parent

	t.OnConnected = EventBus.New()
	t.OnDisconnected = EventBus.New()
	t.OnRequestReceived = EventBus.New()
	t.OnPublishReceived = EventBus.New()
	t.OnPeerConnected = EventBus.New()
	t.OnPeerDisconnected = EventBus.New()

	return t

}

func (t *Tunnel) Start() {

	server, err := kcp.ListenWithOptions("0.0.0.0:31337", nil, 10, 3)

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/Start",
		}).Fatal(err)
	}

	t.parent.logger.WithFields(logrus.Fields{
		"scope": "tunnel/Start",
	}).Info("listen at 0.0.0.0:31337")

	go t.startKCPAccepting(server)

}

func (t *Tunnel) startKCPAccepting(server *kcp.Listener) {

	for {

		conn, err := server.AcceptKCP()

		if err != nil {
			t.parent.logger.WithFields(logrus.Fields{
				"scope": "tunnel/startKCPAccepting",
			}).Fatal(err)
		}

		go t.handleConnection(conn)

	}

}

func (t *Tunnel) handleConnection(conn *kcp.UDPSession) {

	//logger.Println("new peer is connected")

	config := yamux.DefaultConfig()
	config.LogOutput = t.parent.logger.WithFields(logrus.Fields{
		"scope": "tunnel/yamux",
	}).Writer()

	session, err := yamux.Server(conn, config)

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/handleConnection",
		}).Fatal(err)
	}

	go t.OnPeerConnected.Publish("", conn, session)

	for {

		stream, err := session.AcceptStream()

		if err != nil {
			t.parent.logger.WithFields(logrus.Fields{
				"scope": "tunnel/handleConnection",
			}).Warn(err)
			break
		}

		var id int

		decoder := gob.NewDecoder(stream)
		decoder.Decode(&id)

		encoder := gob.NewEncoder(stream)

		switch id {
		case 22:

			var name string
			decoder.Decode(&name)

			t.parent.logger.WithFields(logrus.Fields{
				"scope": "tunnel/handleConnection",
				"name":  name,
			}).Info("receive request")

			go t.OnRequestReceived.Publish(name, conn, session, stream, decoder, encoder)

			break

		}

	}

	go t.OnPeerDisconnected.Publish("", conn, session)

	session.Close()
	conn.Close()

	//logger.Println("peer is disconnected")

}
