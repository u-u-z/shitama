package holder

import (
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/hashicorp/yamux"
	"github.com/sirupsen/logrus"
	kcp "github.com/xtaci/kcp-go"
)

type Holder struct {
	logger  *logrus.Logger
	Tunnel  *Tunnel
	Server  *Server
	shards  []ShardInfo
	clients []ClientInfo
}

func NewHolder() *Holder {

	h := new(Holder)

	h.logger = logrus.New()

	h.Tunnel = NewTunnel(h)
	h.Server = NewServer(h)

	h.shards = make([]ShardInfo, 0)
	h.clients = make([]ClientInfo, 0)

	h.Tunnel.OnRequestReceived.SubscribeAsync("/api/shards/init", func(conn *kcp.UDPSession, session *yamux.Session, stream *yamux.Stream, decoder *gob.Decoder, encoder *gob.Encoder) {

		h.logger.WithFields(logrus.Fields{
			"scope": "holder/handleRequestReceived",
		}).Info("handle /api/shards/init")

		var echoPort int
		decoder.Decode(&echoPort)

		ip := strings.Split(conn.RemoteAddr().String(), ":")[0]

		shard := NewShardInfo(fmt.Sprintf("%s:%d", ip, echoPort), ip, echoPort, conn, session)

		h.shards = append(h.shards, shard)

		encoder.Encode(shard.Addr)

	}, false)

	h.Tunnel.OnRequestReceived.SubscribeAsync("/api/clients/init", func(conn *kcp.UDPSession, session *yamux.Session, stream *yamux.Stream, decoder *gob.Decoder, encoder *gob.Encoder) {

		h.logger.WithFields(logrus.Fields{
			"scope": "holder/handleRequestReceived",
		}).Info("handle /api/clients/init")

		addr := conn.RemoteAddr().String()

		h.clients = append(h.clients, NewClientInfo(addr, conn, session))

		encoder.Encode(addr)

	}, false)

	h.Tunnel.OnRequestReceived.SubscribeAsync("/api/shards", func(conn *kcp.UDPSession, session *yamux.Session, stream *yamux.Stream, decoder *gob.Decoder, encoder *gob.Encoder) {

		h.logger.WithFields(logrus.Fields{
			"scope": "holder/handleRequestReceived",
		}).Info("handle /api/shards")

		encoder.Encode(h.shards)

	}, false)

	h.Tunnel.OnRequestReceived.SubscribeAsync("/api/shards/relay", func(conn *kcp.UDPSession, session *yamux.Session, stream *yamux.Stream, decoder *gob.Decoder, encoder *gob.Encoder) {

		h.logger.WithFields(logrus.Fields{
			"scope": "holder/handleRequestReceived",
		}).Info("handle /api/shards/relay")

		clientAddr := conn.RemoteAddr().String()

		var shardAddr string
		var transport string
		decoder.Decode(&shardAddr)
		decoder.Decode(&transport)

		shard := h.findShardByAddr(shardAddr)

		if shard == nil {
			encoder.Encode("ERROR_SHARD_NOT_FOUND")
			return
		}

		hostAddr, guestAddr := h.zShardRelay(shard.session, clientAddr, transport)

		encoder.Encode(hostAddr)
		encoder.Encode(guestAddr)

	}, false)

	h.Tunnel.OnPeerConnected.SubscribeAsync("", func(conn *kcp.UDPSession, session *yamux.Session) {

		h.logger.WithFields(logrus.Fields{
			"scope": "holder/handlePeerConnected",
		}).Info("handle peer connected")

	}, false)

	h.Tunnel.OnPeerDisconnected.SubscribeAsync("", func(conn *kcp.UDPSession, session *yamux.Session) {

		h.logger.WithFields(logrus.Fields{
			"scope": "holder/handlePeerDisconnected",
		}).Info("handle peer disconnected")

		h.findShardAndRemove(conn, session)

	}, false)

	return h

}

func (h *Holder) Start() {

	h.Tunnel.Start()
	h.Server.Start()

}

func (h *Holder) findShardByAddr(addr string) *ShardInfo {

	for _, v := range h.shards {
		if v.Addr == addr {
			return &v
		}
	}

	return nil

}

func (h *Holder) findShardAndRemove(conn *kcp.UDPSession, session *yamux.Session) int {

	idx := -1

	for i, shard := range h.shards {
		if shard.conn == conn && shard.session == session {
			idx = i
		}
	}

	if idx >= 0 {
		h.shards = append(h.shards[0:idx], h.shards[idx+1:]...)
	}

	return idx

}

func (h *Holder) zShardRelay(session *yamux.Session, clientAddr string, transport string) (hostAddr string, guestAddr string) {

	stream, err := session.OpenStream()

	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"scope": "holder/zShardRelay",
		}).Warn(err)
		return err.Error(), err.Error()
	}

	encoder := gob.NewEncoder(stream)
	encoder.Encode(22)
	encoder.Encode("/api/shards/relay")
	encoder.Encode(clientAddr)
	encoder.Encode(transport)

	decoder := gob.NewDecoder(stream)
	decoder.Decode(&hostAddr)
	decoder.Decode(&guestAddr)

	stream.Close()

	return hostAddr, guestAddr

}
