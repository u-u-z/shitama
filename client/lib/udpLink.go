package client

import (
	"net"

	"github.com/sirupsen/logrus"
)

type UDPLinkPeer struct {
	parent   *UDPLink
	peerAddr net.Addr
	pc       net.PacketConn
}

func NewUDPLinkPeer(parent *UDPLink, peerAddr net.Addr) *UDPLinkPeer {

	p := new(UDPLinkPeer)
	p.parent = parent
	p.peerAddr = peerAddr

	return p

}

func (p *UDPLinkPeer) Start() {

	pc, err := net.ListenPacket("udp4", "0.0.0.0:0")

	if err != nil {
		p.parent.parent.logger.WithFields(logrus.Fields{
			"scope": "udpLink/peer/Start",
		}).Fatal(err)
	}

	p.pc = pc

	go p.handleConnection()

}

func (p *UDPLinkPeer) Stop() {
	p.pc.Close()
}

func (p *UDPLinkPeer) handleConnection() {

	buf := make([]byte, 1536)

	for {

		n, _, err := p.pc.ReadFrom(buf)

		if err != nil {
			p.parent.parent.logger.WithFields(logrus.Fields{
				"scope": "udpLink/peer/Start",
			}).Warn(err)
			break
		}

		//log.Printf("peer server receives %d bytes from %s", n, addr.String())

		p.parent.pc.WriteTo(buf[0:n], p.peerAddr)

	}

}

type UDPLink struct {
	GameAddr net.Addr
	parent   *Client
	pc       net.PacketConn
	peers    map[string]*UDPLinkPeer
	hostAddr net.Addr
}

func NewUDPLink(parent *Client, hostAddr net.Addr) *UDPLink {

	l := new(UDPLink)
	l.GameAddr, _ = net.ResolveUDPAddr("udp4", "127.0.0.1:10800")
	l.parent = parent
	l.hostAddr = hostAddr
	l.peers = make(map[string]*UDPLinkPeer)

	return l

}

func (l *UDPLink) Start() {

	pc, err := net.ListenPacket("udp4", "0.0.0.0:0")

	if err != nil {
		l.parent.logger.WithFields(logrus.Fields{
			"scope": "udpLink/Start",
		}).Fatal(err)
	}

	l.pc = pc

	pc.WriteTo([]byte("PHAMTOM"), l.hostAddr)

	go l.handleConnection()

}

func (l *UDPLink) Stop() {

	for addr, peer := range l.peers {
		peer.Stop()
		delete(l.peers, addr)
	}

	l.pc.Close()

}

func (l *UDPLink) handleConnection() {

	buf := make([]byte, 1536)

	for {

		n, addr, err := l.pc.ReadFrom(buf)

		if err != nil {
			l.parent.logger.WithFields(logrus.Fields{
				"scope": "udpLink/handleConnection",
			}).Warn(err)
			break
		}

		//log.Printf("udp server receives %d bytes from %s", n, addr)

		if _, ok := l.peers[addr.String()]; !ok {
			l.newPeer(addr)
		}

		l.peers[addr.String()].pc.WriteTo(buf[0:n], l.GameAddr)

	}

}

func (l *UDPLink) newPeer(peerAddr net.Addr) *UDPLinkPeer {
	p := NewUDPLinkPeer(l, peerAddr)
	l.peers[peerAddr.String()] = p
	p.Start()
	return p
}
