package shard

import (
	"strconv"
	"strings"
	"time"

	"encoding/gob"

	"net"

	"github.com/asaskevich/EventBus"
	"github.com/hashicorp/yamux"
	"github.com/sirupsen/logrus"
	kcp "github.com/xtaci/kcp-go"
)

type Tunnel struct {
	OnConnected       EventBus.Bus
	OnDisconnected    EventBus.Bus
	OnRequestReceived EventBus.Bus
	OnPublishReceived EventBus.Bus
	OnPipeReceived    EventBus.Bus
	parent            *Shard
	conn              *kcp.UDPSession
	session           *yamux.Session
}

func NewTunnel(parent *Shard) *Tunnel {

	t := new(Tunnel)
	t.parent = parent

	t.OnConnected = EventBus.New()
	t.OnDisconnected = EventBus.New()
	t.OnRequestReceived = EventBus.New()
	t.OnPublishReceived = EventBus.New()
	t.OnPipeReceived = EventBus.New()

	t.OnDisconnected.SubscribeAsync("", func() {
		go t.Start()
	}, true)

	return t

}

func (t *Tunnel) Start() {

	holderKcpAddr := t.parent.Config["holderKcpAddr"].(string)
	holderKcpAddrAlt := t.parent.Config["holderKcpAddrAlt"].(string)

	var err error
	var conn *kcp.UDPSession

	conn, err = kcp.DialWithOptions(holderKcpAddr, nil, 10, 3)

	if err != nil {

		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/Start",
		}).Warn(err)

		conn, err = kcp.DialWithOptions(holderKcpAddrAlt, nil, 10, 3)

		if err != nil {
			t.parent.logger.WithFields(logrus.Fields{
				"scope": "tunnel/Start",
			}).Fatal(err)
		}

	}

	t.conn = conn

	go t.handleConnection(conn)

}

func (t *Tunnel) handleConnection(conn *kcp.UDPSession) {

	config := yamux.DefaultConfig()
	config.LogOutput = t.parent.logger.WithFields(logrus.Fields{
		"scope": "tunnel/yamux",
	}).Writer()

	session, err := yamux.Client(conn, config)

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/handleConnection",
		}).Fatal(err)
	}

	t.session = session

	_, err = t.session.Ping()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/handleConnection",
		}).Warn(err)
		t.stopConnection()
		return
	}

	go t.keepAlive()

	t.zInitShard()

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

}

func (t *Tunnel) stopConnection() {

	err := t.session.GoAway()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/stopConnection",
		}).Warn(err)
	}

	err = t.session.Close()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/stopConnection",
		}).Warn(err)
	}

	err = t.conn.Close()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/stopConnection",
		}).Warn(err)
	}

	t.OnDisconnected.Publish("")

}

func (t *Tunnel) keepAlive() {

	for {

		_, err := t.session.Ping()

		if err != nil {
			t.parent.logger.Println(err)
			t.stopConnection()
			break
		}

		//log.Printf("rtt is %dms", duration/time.Millisecond)

		time.Sleep(1 * time.Second)

	}

}

func (t *Tunnel) zInitShard() {

	stream, err := t.session.OpenStream()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/zInitShard",
		}).Warn(err)
		return
	}

	echoPort, err := strconv.Atoi(strings.Split(t.parent.Echo.pc.LocalAddr().String(), ":")[1])

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/zInitShard",
		}).Warn(err)
		return
	}

	encoder := gob.NewEncoder(stream)
	encoder.Encode(22)
	encoder.Encode("/api/shards/init")
	encoder.Encode(echoPort)

	var publicAddr string

	decoder := gob.NewDecoder(stream)
	decoder.Decode(&publicAddr)

	addr, err := net.ResolveUDPAddr("udp4", publicAddr)

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/zInitShard",
		}).Warn(err)
		return
	}

	t.OnConnected.Publish("", addr)

	stream.Close()

}
