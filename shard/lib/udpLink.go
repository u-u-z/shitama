package shard

import (
	"fmt"
	"net"
	"strings"
	"time"

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
		p.parent.parent.parent.logger.WithFields(logrus.Fields{
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
			p.parent.parent.parent.logger.WithFields(logrus.Fields{
				"scope": "udpLink/peer/handleConnection",
			}).Warn(err)
			break
		}

		//log.Printf("peer server receives %d bytes from %s", n, addr.String())

		p.parent.active = time.Now()

		p.parent.pcGuest.WriteTo(buf[0:n], p.peerAddr)

	}

}

type UDPLink struct {
	parent     *Portal
	pcHost     net.PacketConn
	pcGuest    net.PacketConn
	peers      map[string]*UDPLinkPeer
	clientAddr net.Addr
	hostAddr   net.Addr
	active     time.Time
}

func NewUDPLink(parent *Portal, clientAddr net.Addr) *UDPLink {

	l := new(UDPLink)
	l.parent = parent
	l.clientAddr = clientAddr
	l.peers = make(map[string]*UDPLinkPeer)

	return l

}

func (l *UDPLink) Start() {

	pcHost, err := net.ListenPacket("udp4", "0.0.0.0:0")

	if err != nil {
		l.parent.parent.logger.WithFields(logrus.Fields{
			"scope": "udpLink/Start",
		}).Fatal(err)
	}

	l.pcHost = pcHost

	pcGuest, err := net.ListenPacket("udp4", "0.0.0.0:0")

	if err != nil {
		l.parent.parent.logger.WithFields(logrus.Fields{
			"scope": "udpLink/Start",
		}).Fatal(err)
	}

	l.pcGuest = pcGuest

	l.active = time.Now()

	go l.handleHostConnection()
	go l.handleGuestConnection()

}

func (l *UDPLink) HostAddr() string {
	return fmt.Sprintf("%s:%s",
		strings.Split(l.parent.parent.PublicAddr.String(), ":")[0],
		strings.Split(l.pcHost.LocalAddr().String(), ":")[1],
	)
}

func (l *UDPLink) GuestAddr() string {
	return fmt.Sprintf("%s:%s",
		strings.Split(l.parent.parent.PublicAddr.String(), ":")[0],
		strings.Split(l.pcGuest.LocalAddr().String(), ":")[1],
	)
}

func (l *UDPLink) Transport() string {
	return "udp"
}

func (l *UDPLink) Expired() bool {
	return time.Now().Sub(l.active).Minutes() > 1
}

func (l *UDPLink) Stop() {

	for addr, peer := range l.peers {
		peer.Stop()
		delete(l.peers, addr)
	}

	l.pcHost.Close()
	l.pcGuest.Close()

}

func (l *UDPLink) handleHostConnection() {

	buf := make([]byte, 256)

	for {

		_, addr, err := l.pcHost.ReadFrom(buf)

		if err != nil {
			l.parent.parent.logger.WithFields(logrus.Fields{
				"scope": "udpLink/handleHostConnection",
			}).Warn(err)
			break
		}

		//log.Printf("host server receives %d bytes from %s", n, addr.String())

		if l.hostAddr == nil {
			l.hostAddr = addr
		}

	}

}

func (l *UDPLink) handleGuestConnection() {

	buf := make([]byte, 1536)

	for {

		n, addr, err := l.pcGuest.ReadFrom(buf)

		if err != nil {
			l.parent.parent.logger.WithFields(logrus.Fields{
				"scope": "udpLink/handleGuestConnection",
			}).Warn(err)
			break
		}

		//log.Printf("guest server receives %d bytes from %s", n, addr)

		if _, ok := l.peers[addr.String()]; !ok {
			l.newPeer(addr)
		}

		l.peers[addr.String()].pc.WriteTo(buf[0:n], l.hostAddr)

	}

}

func (l *UDPLink) newPeer(peerAddr net.Addr) *UDPLinkPeer {

	p := NewUDPLinkPeer(l, peerAddr)

	l.peers[peerAddr.String()] = p

	p.Start()

	return p

}
