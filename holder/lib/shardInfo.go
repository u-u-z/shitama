package holder

import (
	"github.com/evshiron/shitama/common"
	"github.com/hashicorp/yamux"
	kcp "github.com/xtaci/kcp-go"
)

type ShardInfo struct {
	common.ShardInfo
	conn    *kcp.UDPSession
	session *yamux.Session
}

func NewShardInfo(addr string, ip string, echoPort int, conn *kcp.UDPSession, session *yamux.Session) ShardInfo {

	s := ShardInfo{common.ShardInfo{Addr: addr, IP: ip, EchoPort: echoPort}, nil, nil}
	s.conn = conn
	s.session = session

	return s

}
