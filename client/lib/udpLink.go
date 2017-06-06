package client

import (
	"net"

	"encoding/binary"

	"github.com/sirupsen/logrus"
)

type UDPLinkDummy struct {
	parent    *UDPLink
	guestAddr *net.UDPAddr
	pc        net.PacketConn
}

func NewUDPLinkDummy(parent *UDPLink, guestAddr *net.UDPAddr) *UDPLinkDummy {

	d := new(UDPLinkDummy)
	d.parent = parent
	d.guestAddr = guestAddr

	return d

}

func (d *UDPLinkDummy) Start() {

	pc, err := net.ListenPacket("udp4", "0.0.0.0:0")

	if err != nil {
		d.parent.parent.logger.WithFields(logrus.Fields{
			"scope": "udpLink/dummy/Start",
		}).Fatal(err)
	}

	d.pc = pc

	go d.handleConnection()

}

func (d *UDPLinkDummy) Stop() {
	d.pc.Close()
}

func (d *UDPLinkDummy) handleConnection() {

	buf := make([]byte, 1536)

	for {

		n, _, err := d.pc.ReadFrom(buf)

		if err != nil {
			d.parent.parent.logger.WithFields(logrus.Fields{
				"scope": "udpLink/dummy/handleConnection",
			}).Warn(err)
			break
		}

		d.parent.pc.WriteTo(d.parent.packData(d.guestAddr, buf[:n]), d.parent.hostAddr)

	}

}

type UDPLink struct {
	GameAddr net.Addr
	parent   *Client
	pc       net.PacketConn
	dummies  map[string]*UDPLinkDummy
	hostAddr net.Addr
}

func NewUDPLink(parent *Client, hostAddr net.Addr) *UDPLink {

	l := new(UDPLink)
	l.GameAddr, _ = net.ResolveUDPAddr("udp4", "127.0.0.1:10800")
	l.parent = parent
	l.hostAddr = hostAddr
	l.dummies = make(map[string]*UDPLinkDummy)

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

	pc.WriteTo([]byte("PHANTOM"), l.hostAddr)

	go l.handleConnection()

}

func (l *UDPLink) Stop() {

	for addr, peer := range l.dummies {
		peer.Stop()
		delete(l.dummies, addr)
	}

	l.pc.Close()

}

func (l *UDPLink) handleConnection() {

	buf := make([]byte, 1536)

	for {

		n, _, err := l.pc.ReadFrom(buf)

		if err != nil {
			l.parent.logger.WithFields(logrus.Fields{
				"scope": "udpLink/handleConnection",
			}).Warn(err)
			break
		}

		guestAddr, data := l.unpackData(buf[:n])
		guestKey := guestAddr.String()

		if _, ok := l.dummies[guestKey]; !ok {
			l.newDummy(guestAddr)
		}

		l.dummies[guestKey].pc.WriteTo(data, l.GameAddr)

	}

}

func (l *UDPLink) newDummy(peerAddr *net.UDPAddr) *UDPLinkDummy {
	p := NewUDPLinkDummy(l, peerAddr)
	l.dummies[peerAddr.String()] = p
	p.Start()
	return p
}

func (l *UDPLink) packData(addr *net.UDPAddr, data []byte) []byte {

	/*

		buffer := new(bytes.Buffer)

		encoder := gob.NewEncoder(buffer)
		encoder.Encode(addr)
		encoder.Encode(data)

		return buffer.Bytes()

	*/

	buf := make([]byte, len(data)+6)

	copy(buf[:4], addr.IP[len(addr.IP)-4:])
	binary.BigEndian.PutUint16(buf[4:6], uint16(addr.Port))
	copy(buf[6:], data)

	return buf

}

func (l *UDPLink) unpackData(buf []byte) (addr *net.UDPAddr, data []byte) {

	/*

		buffer := bytes.NewBuffer(buf)

		decoder := gob.NewDecoder(buffer)
		decoder.Decode(&addr)
		decoder.Decode(&data)

		return addr, data

	*/

	addr = new(net.UDPAddr)
	addr.IP = make([]byte, 16)
	addr.IP[10] = 255
	addr.IP[11] = 255
	copy(addr.IP[len(addr.IP)-4:], buf[:4])
	addr.Port = int(binary.BigEndian.Uint16(buf[4:6]))

	data = make([]byte, len(buf)-6)
	copy(data, buf[6:])

	return addr, data

}
