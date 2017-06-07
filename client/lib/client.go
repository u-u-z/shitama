package client

import (
	"encoding/binary"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type ShardInfoSlice []ShardInfo

func (s ShardInfoSlice) Len() int {
	return len(s)
}

func (s ShardInfoSlice) Less(i, j int) bool {
	return s[i].RTT < s[j].RTT
}

func (s ShardInfoSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type Client struct {
	logger    *logrus.Logger
	Config    map[string]interface{}
	Tunnel    *Tunnel
	API       *API
	connected bool
	shards    []ShardInfo
	link      *UDPLink
}

func NewClient() *Client {

	c := new(Client)

	c.logger = logrus.New()

	c.Config = make(map[string]interface{})
	c.Config["holderKcpAddr"] = "shitama.tldr.run:31337"

	c.Tunnel = NewTunnel(c)
	c.API = NewAPI(c)

	c.Tunnel.OnConnected.SubscribeAsync("", func() {

		c.logger.WithFields(logrus.Fields{
			"scope": "client/handleConnected",
		}).Println("connected")

		c.connected = true

	}, false)

	c.Tunnel.OnDisconnected.SubscribeAsync("", func() {

		c.logger.WithFields(logrus.Fields{
			"scope": "client/handleDisconnected",
		}).Println("disconnected")

		c.connected = false

	}, false)

	return c

}

func (c *Client) Start() {

	c.Tunnel.Start()
	c.API.Start()

}

func (c *Client) GetStatus() map[string]interface{} {

	status := make(map[string]interface{})
	status["connected"] = c.connected

	return status

}

func (c *Client) UpdateShards() []ShardInfo {

	if !c.connected {
		return make([]ShardInfo, 0)
	}

	shards := c.Tunnel.zGetShards()

	if shards == nil {
		shards = make([]ShardInfo, 0)
	}

	c.updateRTTs(shards)

	sort.Sort(ShardInfoSlice(shards))

	c.shards = shards

	return shards

}

func (c *Client) RequestRelay(shardAddr string, transport string) (hostAddr string, guestAddr string) {

	if !c.connected {
		return "ERROR_UNCONNECTED", "ERROR_UNCONNECTED"
	}

	hostAddr, guestAddr = c.Tunnel.zShardRelay(shardAddr, transport)

	if !strings.Contains(hostAddr, "ERROR") && !strings.Contains(guestAddr, "ERROR") {
		c.newLink(hostAddr, transport)
	}

	return hostAddr, guestAddr

}

func (c *Client) updateRTTs(shards []ShardInfo) {

	type Pair struct {
		key   string
		value uint64
	}

	pc, err := net.ListenPacket("udp4", "0.0.0.0:0")

	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"scope": "client/updateRTTs",
		}).Fatal(err)
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(time.Now().UnixNano()))

	for _, shard := range shards {
		addr, err := net.ResolveUDPAddr("udp4", shard.Addr)
		if err != nil {
			c.logger.WithFields(logrus.Fields{
				"scope": "client/updateRTTs",
			}).Warn(err)
			continue
		}
		for i := 0; i < 16; i++ {
			pc.WriteTo(buf, addr)
		}
	}

	shardRTTs := make(map[string][]uint64)

	pairs := make(chan Pair)

	go (func() {

		for {

			_, addr, err := pc.ReadFrom(buf)

			if err != nil {
				c.logger.WithFields(logrus.Fields{
					"scope": "client/updateRTTs",
				}).Warn(err)
				break
			}

			key := addr.String()
			now := uint64(time.Now().UnixNano())
			then := binary.BigEndian.Uint64(buf)

			pairs <- Pair{key: key, value: now - then}

		}

	})()

WaitLoop:
	for {
		select {
		case v := <-pairs:
			if _, ok := shardRTTs[v.key]; !ok {
				shardRTTs[v.key] = make([]uint64, 0)
			}
			shardRTTs[v.key] = append(shardRTTs[v.key], v.value)
			break
		case <-time.After(1 * time.Second):
			break WaitLoop
		}
	}

	pc.Close()

	for idx := range shards {
		shard := &shards[idx]
		if rtts, ok := shardRTTs[shard.Addr]; ok {
			var sum uint64
			for _, v := range rtts {
				sum += v
			}
			shard.RTT = float32(sum) / 1e6 / float32(len(rtts))
		}
	}

}

func (c *Client) newLink(hostAddr string, transport string) {

	if c.link != nil {
		c.link.Stop()
	}

	addr, err := net.ResolveUDPAddr("udp4", hostAddr)

	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"scope": "client/newLink",
		}).Fatal(err)
	}

	switch transport {
	case "udp":
		c.link = NewUDPLink(c, addr)
		c.link.Start()
		break
	}

}
