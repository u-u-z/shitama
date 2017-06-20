package shard

import (
	"encoding/gob"
	"net"

	"github.com/hashicorp/yamux"
	"github.com/sirupsen/logrus"
	kcp "github.com/xtaci/kcp-go"
)

type Shard struct {
	Config     map[string]interface{}
	Echo       *Echo
	Tunnel     *Tunnel
	Portal     *Portal
	PublicAddr net.Addr
	logger     *logrus.Logger
}

func NewShard() *Shard {

	s := new(Shard)

	s.logger = logrus.New()

	s.Config = make(map[string]interface{})
	s.Config["holderKcpAddr"] = "shitama.tldr.run:31337"
	s.Config["holderKcpAddrAlt"] = "115.159.87.170:31337"
	s.Config["echoPort"] = 0

	s.Echo = NewEcho(s)
	s.Tunnel = NewTunnel(s)
	s.Portal = NewPortal(s)

	s.Tunnel.OnConnected.SubscribeAsync("", func(publicAddr net.Addr) {
		s.logger.WithFields(logrus.Fields{
			"scope":      "shard/handleConnected",
			"publicAddr": publicAddr,
		}).Info("connected")
		s.PublicAddr = publicAddr
	}, true)

	s.Tunnel.OnDisconnected.SubscribeAsync("", func() {
		s.logger.WithFields(logrus.Fields{
			"scope": "shard/handleDisconnected",
		}).Info("disconnected")
	}, true)

	s.Tunnel.OnRequestReceived.SubscribeAsync("/api/shards/relay", func(conn *kcp.UDPSession, session *yamux.Session, stream *yamux.Stream, decoder *gob.Decoder, encoder *gob.Encoder) {

		s.logger.WithFields(logrus.Fields{
			"scope": "shard/handleRequestReceived",
		}).Info("handle /api/shards/relay")

		var clientAddr string
		var transport string
		decoder.Decode(&clientAddr)
		decoder.Decode(&transport)

		link := s.Portal.NewLink(clientAddr, transport)

		if link == nil {
			encoder.Encode("ERROR_LINK_INVALID")
			encoder.Encode("ERROR_LINK_INVALID")
			return
		}

		encoder.Encode(link.HostAddr())
		encoder.Encode(link.GuestAddr())

	}, true)

	return s

}

func (s *Shard) Start() {

	s.Echo.Start()
	s.Tunnel.Start()
	s.Portal.Start()

}
