package client

import (
	"time"

	"encoding/gob"

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
	parent            *Client
	conn              *kcp.UDPSession
	session           *yamux.Session
}

func NewTunnel(parent *Client) *Tunnel {

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

	t.zInitClient()

	t.OnConnected.Publish("")

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
			t.parent.logger.WithFields(logrus.Fields{
				"scope": "tunnel/keepAlive",
			}).Warn(err)
			t.stopConnection()
			break
		}

		//log.Printf("rtt is %dms", duration/time.Millisecond)

		time.Sleep(1 * time.Second)

	}

}

func (t *Tunnel) zInitClient() {

	stream, err := t.session.OpenStream()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/zInitClient",
		}).Warn(err)
		return
	}

	encoder := gob.NewEncoder(stream)
	encoder.Encode(22)
	encoder.Encode("/api/clients/init")

	var addr string

	decoder := gob.NewDecoder(stream)
	decoder.Decode(&addr)

	stream.Close()

}

func (t *Tunnel) zGetShards() []ShardInfo {

	stream, err := t.session.OpenStream()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/zGetShards",
		}).Warn(err)
		return nil
	}

	encoder := gob.NewEncoder(stream)
	encoder.Encode(22)
	encoder.Encode("/api/shards")

	data := make([]ShardInfo, 0)

	decoder := gob.NewDecoder(stream)
	decoder.Decode(&data)

	stream.Close()

	return data

}

func (t *Tunnel) zShardRelay(shardAddr string, transport string) (hostAddr string, guestAddr string) {

	stream, err := t.session.OpenStream()

	if err != nil {
		t.parent.logger.WithFields(logrus.Fields{
			"scope": "tunnel/zShardRelay",
		}).Warn(err)
		return err.Error(), err.Error()
	}

	encoder := gob.NewEncoder(stream)
	encoder.Encode(22)
	encoder.Encode("/api/shards/relay")
	encoder.Encode(shardAddr)
	encoder.Encode(transport)

	decoder := gob.NewDecoder(stream)
	decoder.Decode(&hostAddr)
	decoder.Decode(&guestAddr)

	stream.Close()

	return hostAddr, guestAddr

}
