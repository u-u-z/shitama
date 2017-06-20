package client

import (
	"bytes"
	"net"
	"time"

	"encoding/binary"

	"github.com/sirupsen/logrus"
)

type UDPLinkDummy struct {
	parent    *UDPLink
	peerAddr  *net.UDPAddr
	pc        net.PacketConn
	delay     int64
	delayBase time.Time
	profile   string
	active    time.Time
}

func NewUDPLinkDummy(parent *UDPLink, peerAddr *net.UDPAddr) *UDPLinkDummy {

	d := new(UDPLinkDummy)
	d.parent = parent
	d.peerAddr = peerAddr
	d.delay = 65535
	d.delayBase = time.Now()
	d.profile = "key"
	d.active = time.Now()

	return d

}

func (d *UDPLinkDummy) Start() {

	pc, err := net.ListenPacket("udp4", "127.0.0.1:0")

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

	d.parent.parent.logger.WithFields(logrus.Fields{
		"scope": "udpLink/dummy/Stop",
		"key":   d.peerAddr,
	}).Info("dummy stopped")

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

		d.active = time.Now()

		if buf[0] == 0x3 {
			// FIXME: This delay is from Shitama to Hisouten.
			d.delay = time.Now().Sub(d.delayBase).Nanoseconds()
		}

		if buf[0] == 0x8 {

			len := int(binary.LittleEndian.Uint32(buf[1:5]))

			for i := 0; i < len; i++ {

				pre := d.parent.sockAddrToUDPAddr(buf[5+i*16:])
				dummy := d.parent.findDummyByAddr(pre)

				if dummy != nil {

					copy(buf[5+i*16:], d.parent.udpAddrToSockAddr(dummy.peerAddr))

					/*
						post := d.parent.sockAddrToUDPAddr(buf[5+i*16:])

						d.parent.parent.logger.WithFields(logrus.Fields{
							"scope": "udpLink/dummy/handleConnection",
							"pre":   pre,
							"post":  post,
						}).Info("rewrite 0x8 packet")
					*/

				} else {

					d.parent.parent.logger.WithFields(logrus.Fields{
						"scope": "udpLink/dummy/handleConnection",
						"pre":   pre,
					}).Warn("rewrite 0x8 packet mismatch")

				}

			}

		}

		if buf[0] == 0x2 {

			pre := d.parent.sockAddrToUDPAddr(buf[1:])
			dummy := d.parent.findDummyByAddr(pre)

			if dummy != nil {

				copy(buf[1:], d.parent.udpAddrToSockAddr(dummy.peerAddr))

				/*
					post := d.parent.sockAddrToUDPAddr(buf[1:])

					d.parent.parent.logger.WithFields(logrus.Fields{
						"scope": "udpLink/dummy/handleConnection",
						"pre":   pre,
						"post":  post,
					}).Info("rewrite 0x2 packet")
				*/

			} else {

				d.parent.parent.logger.WithFields(logrus.Fields{
					"scope": "udpLink/dummy/handleConnection",
					"pre":   pre,
				}).Warn("rewrite 0x2 packet mismatch")

			}

		}

		d.parent.pc.WriteTo(d.parent.packData(d.peerAddr, buf[:n]), d.parent.hostAddr)

	}

}

type UDPLink struct {
	GameAddr   net.Addr
	parent     *Client
	pc         net.PacketConn
	shardAddr  net.Addr
	hostAddr   net.Addr
	dummies    map[string]*UDPLinkDummy
	delay      int64
	delayDelta int64
	active     time.Time
}

func NewUDPLink(parent *Client, shardAddr net.Addr, hostAddr net.Addr) *UDPLink {

	l := new(UDPLink)
	l.GameAddr, _ = net.ResolveUDPAddr("udp4", "127.0.0.1:10800")
	l.parent = parent
	l.shardAddr = shardAddr
	l.hostAddr = hostAddr
	l.dummies = make(map[string]*UDPLinkDummy)
	l.delay = 65535
	l.delayDelta = 0
	l.active = time.Now()

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

	go l.keepAlive()
	go l.updateDelay()
	go l.handleConnection()

}

func (l *UDPLink) Stop() {

	for addr, peer := range l.dummies {
		peer.Stop()
		delete(l.dummies, addr)
	}

	l.pc.Close()

}

func (l *UDPLink) keepAlive() {

	for {

		_, err := l.pc.WriteTo([]byte("PHANTOM"), l.hostAddr)

		if err != nil {
			l.parent.logger.WithFields(logrus.Fields{
				"scope": "udpLink/keepAlive",
			}).Warn(err)
			break
		}

		time.Sleep(10 * time.Second)

	}

}

func (l *UDPLink) updateDelay() {

	pc, err := net.ListenPacket("udp4", "0.0.0.0:0")

	if err != nil {
		l.parent.logger.WithFields(logrus.Fields{
			"scope": "udpLink/updateDelay",
		}).Fatal(err)
	}

	go (func() {

		for {

			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, uint64(time.Now().UnixNano()))

			_, err := pc.WriteTo(buf, l.shardAddr)

			if err != nil {
				l.parent.logger.WithFields(logrus.Fields{
					"scope": "udpLink/updateRelay/sender",
				}).Warn(err)
				break
			}

			time.Sleep(1 * time.Second)

		}

	})()

	go (func() {

		buf := make([]byte, 1536)

		for {

			_, _, err := pc.ReadFrom(buf)

			if err != nil {
				l.parent.logger.WithFields(logrus.Fields{
					"scope": "udpLink/updateRelay/receiver",
				}).Warn(err)
				break
			}

			now := uint64(time.Now().UnixNano())
			then := binary.BigEndian.Uint64(buf)
			delay := int64(now - then)

			l.delayDelta = delay - l.delay
			l.delay = delay

		}

	})()

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

		l.active = time.Now()

		peerAddr, data := l.unpackData(buf[:n])
		key := peerAddr.String()

		if _, ok := l.dummies[key]; !ok {
			l.newDummy(peerAddr)
		}

		dummy := l.dummies[key]

		if data[0] == 0x1 {

			if bytes.Compare(data[1:17], data[17:33]) != 0 {

				pre := l.sockAddrToUDPAddr(data[17:25])
				dest := l.dummies[pre.String()]

				if dest != nil {

					localAddr := dest.pc.LocalAddr().(*net.UDPAddr)

					copy(data[17:], l.udpAddrToSockAddr(localAddr))

					/*
						post := l.sockAddrToUDPAddr(data[17:25])

						l.parent.logger.WithFields(logrus.Fields{
							"scope": "udpLink/handleConnection",
							"pre":   pre,
							"post":  post,
						}).Info("rewrite 0x1 packet")
					*/

				} else {

					l.parent.logger.WithFields(logrus.Fields{
						"scope": "udpLink/handleConnection",
						"pre":   pre,
					}).Warn("rewrite 0x1 packet mismatch")

				}

			} else {

				dummy.delayBase = time.Now()

			}

		}

		if data[0] == 0x5 {

			len := data[26]

			if len < 0x10 {
				dummy.profile = string(data[27 : 27+len])
			}

		}

		dummy.pc.WriteTo(data, l.GameAddr)

	}

}

func (l *UDPLink) newDummy(peerAddr *net.UDPAddr) *UDPLinkDummy {

	p := NewUDPLinkDummy(l, peerAddr)

	key := peerAddr.String()
	l.dummies[key] = p

	p.Start()

	l.parent.logger.WithFields(logrus.Fields{
		"scope": "udpLink/newDummy",
		"key":   key,
	}).Info("new dummy started")

	return p

}

// Need optimization.
func (l *UDPLink) findDummyByAddr(addr net.Addr) *UDPLinkDummy {

	for _, dummy := range l.dummies {
		if dummy.pc.LocalAddr().String() == addr.String() {
			return dummy
		}
	}

	return nil

}

func (l *UDPLink) udpAddrToSockAddr(addr *net.UDPAddr) []byte {

	buf := make([]byte, 8)

	binary.BigEndian.PutUint16(buf[:2], 0x200)
	binary.BigEndian.PutUint16(buf[2:4], uint16(addr.Port))
	copy(buf[4:8], addr.IP[len(addr.IP)-4:])

	return buf

}

func (l *UDPLink) sockAddrToUDPAddr(buf []byte) *net.UDPAddr {

	addr := new(net.UDPAddr)
	addr.IP = make([]byte, 16)
	addr.IP[10] = 255
	addr.IP[11] = 255
	copy(addr.IP[len(addr.IP)-4:], buf[4:8])
	addr.Port = int(binary.BigEndian.Uint16(buf[2:4]))

	return addr

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
