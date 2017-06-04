package holder

import (
	"github.com/evshiron/shitama/common"
	"github.com/hashicorp/yamux"
	kcp "github.com/xtaci/kcp-go"
)

type ClientInfo struct {
	common.ClientInfo
	conn    *kcp.UDPSession
	session *yamux.Session
}

func NewClientInfo(addr string, conn *kcp.UDPSession, session *yamux.Session) ClientInfo {

	c := ClientInfo{common.ClientInfo{Addr: addr}, nil, nil}
	c.conn = conn
	c.session = session

	return c

}
