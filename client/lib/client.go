package client

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"os"

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

	switch os.Getenv("LOGLEVEL") {
	case "debug":
		c.logger.Level = logrus.DebugLevel
		break
	default:
		c.logger.Level = logrus.InfoLevel
		break
	}

	c.Config = make(map[string]interface{})
	c.Config["holderKcpAddr"] = "shitama.tldr.run:31337"
	c.Config["holderKcpAddrAlt"] = "115.159.87.170:31337"

	c.Tunnel = NewTunnel(c)
	c.API = NewAPI(c)

	c.shards = make([]ShardInfo, 0)
	c.link = nil

	c.Tunnel.OnConnected.SubscribeAsync("", func() {

		c.logger.WithFields(logrus.Fields{
			"scope": "client/handleConnected",
		}).Info("connected")

		c.connected = true

	}, false)

	c.Tunnel.OnDisconnected.SubscribeAsync("", func() {

		c.logger.WithFields(logrus.Fields{
			"scope": "client/handleDisconnected",
		}).Info("disconnected")

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

	shard := c.findShardByAddr(shardAddr)

	if shard == nil {
		return "ERROR_SHARD_NOT_FOUND", "ERROR_SHARD_NOT_FOUND"
	}

	hostAddr, guestAddr = c.Tunnel.zShardRelay(shard.Addr, transport)

	if !strings.Contains(hostAddr, "ERROR") && !strings.Contains(guestAddr, "ERROR") {
		c.newLink(shard, hostAddr, transport)
	}

	return hostAddr, guestAddr

}

func (c *Client) GetConnectionStatus() map[string]interface{} {

	status := make(map[string]interface{})

	if c.link == nil {

		status["linkEstablished"] = false
		status["linkAddr"] = ""
		status["linkDelay"] = 0
		status["linkDelayDelta"] = 0
		status["peers"] = make([]interface{}, 0)

	} else {

		status["linkEstablished"] = true
		status["linkAddr"] = c.link.hostAddr.String()
		status["linkDelay"] = c.link.delay
		status["linkDelayDelta"] = c.link.delayDelta
		status["peers"] = make([]map[string]interface{}, 0)

		for _, dummy := range c.link.dummies {

			peer := make(map[string]interface{})
			peer["remoteAddr"] = dummy.peerAddr.String()
			peer["localAddr"] = dummy.pc.LocalAddr().String()
			peer["delay"] = dummy.delay
			peer["profile"] = dummy.profile
			peer["active"] = dummy.active.UnixNano()

			status["peers"] = append(status["peers"].([]map[string]interface{}), peer)

		}

	}

	return status

}

func (c *Client) findShardByAddr(addr string) *ShardInfo {

	for _, shard := range c.shards {
		if shard.Addr == addr {
			return &shard
		}
	}

	return nil

}

func (c *Client) tcpPing(ip string, ports []uint16) time.Duration {

	for _, port := range ports {

		addr := fmt.Sprintf("%s:%d", ip, port)

		start := time.Now()

		conn, err := net.Dial("tcp4", addr)
		defer conn.Close()

		if err != nil {
			continue
		}

		return time.Since(start)

	}

	return 9999 * time.Second

}

func (c *Client) updateRTTs(shards []ShardInfo) {

	for idx := range shards {
		rtt := c.tcpPing(shards[idx].IP, []uint16{22, 3389})
		shards[idx].RTT = float32(rtt) / 1e6
	}

}

func (c *Client) newLink(shard *ShardInfo, hostAddr string, transport string) {

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

		shardAddr, err := net.ResolveUDPAddr("udp4", shard.Addr)
		if err != nil {
			c.logger.WithFields(logrus.Fields{
				"scope": "client/newLink",
			}).Warn(err)
		}

		c.link = NewUDPLink(c, shardAddr, addr)
		c.link.Start()

		break

	}

}
