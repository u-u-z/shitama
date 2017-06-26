package client

import (
	"bytes"
	"net"
	"time"

	"encoding/binary"

	"github.com/evshiron/shitama/common"
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
	d.delay = 0
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

	go d.updateDelay()
	go d.handleConnection()

}

func (d *UDPLinkDummy) Stop() {

	d.pc.Close()

	d.parent.parent.logger.WithFields(logrus.Fields{
		"scope": "udpLink/dummy/Stop",
		"key":   d.peerAddr,
	}).Info("dummy stopped")

}

func (d *UDPLinkDummy) updateDelay() {

	go (func() {

		buf := make([]byte, 37)

		for {

			d.delayBase = time.Now()

			buf[0] = 0x1
			copy(buf[1:17], common.UDPAddrToSockAddr(d.peerAddr))
			copy(buf[17:33], common.UDPAddrToSockAddr(d.peerAddr))

			_, err := d.parent.pc.WriteTo(common.PackData(d.peerAddr, buf), d.parent.hostAddr)

			if err != nil {
				d.parent.parent.logger.WithFields(logrus.Fields{
					"scope": "udpLink/dummy/updateDelay/sender",
				}).Warn(err)
				break
			}

			time.Sleep(1 * time.Second)

		}

	})()

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

		// Duplicate the data.
		data := make([]byte, n)
		copy(data, buf[:n])

		go d.handleReceivedPacket(data)

	}

}

func (d *UDPLinkDummy) handleReceivedPacket(data []byte) {

	/*
		if data[0] == 0x3 {
			// FIXME: This delay is from Shitama to Hisouten.
			d.delay = time.Now().Sub(d.delayBase).Nanoseconds()
		}
	*/

	if data[0] == 0x8 {

		len := int(binary.LittleEndian.Uint32(data[1:5]))

		for i := 0; i < len; i++ {

			pre := common.SockAddrToUDPAddr(data[5+i*16:])
			dummy := d.parent.findDummyByAddr(pre)

			if dummy != nil {

				copy(data[5+i*16:], common.UDPAddrToSockAddr(dummy.peerAddr))

				/*
					post := common.SockAddrToUDPAddr(data[5+i*16:])

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

	if data[0] == 0x2 {

		pre := common.SockAddrToUDPAddr(data[1:])
		dummy := d.parent.findDummyByAddr(pre)

		if dummy != nil {

			copy(data[1:], common.UDPAddrToSockAddr(dummy.peerAddr))

			/*
				post := common.SockAddrToUDPAddr(buf[1:])

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

	d.parent.pc.WriteTo(common.PackData(d.peerAddr, data), d.parent.hostAddr)

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
	l.delay = 0
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

		buf := make([]byte, 8)

		for {

			binary.BigEndian.PutUint64(buf, uint64(time.Now().UnixNano()))

			_, err := pc.WriteTo(buf, l.shardAddr)

			if err != nil {
				l.parent.logger.WithFields(logrus.Fields{
					"scope": "udpLink/updateDelay/sender",
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
					"scope": "udpLink/updateDelay/receiver",
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

		peerAddr, data := common.UnpackData(buf[:n])

		key := peerAddr.String()

		if _, ok := l.dummies[key]; !ok {
			l.newDummy(peerAddr)
		}

		dummy := l.dummies[key]

		// data is duplicated.
		go l.handleReceivedPacket(dummy, data)

	}

}

func (l *UDPLink) handleReceivedPacket(dummy *UDPLinkDummy, data []byte) {

	if data[0] == 0x3 {
		// FIXME: This delay is from Shitama to Hisouten.
		dummy.delay = time.Now().Sub(dummy.delayBase).Nanoseconds()
	}

	if data[0] == 0x1 {

		if bytes.Compare(data[1:17], data[17:33]) != 0 {

			pre := common.SockAddrToUDPAddr(data[17:25])
			dest := l.dummies[pre.String()]

			if dest != nil {

				localAddr := dest.pc.LocalAddr().(*net.UDPAddr)

				copy(data[17:], common.UDPAddrToSockAddr(localAddr))

				/*
					post := common.SockAddrToUDPAddr(data[17:25])

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
			/*
				dummy.delayBase = time.Now()
			*/
		}

	}

	if data[0] == 0x5 {

		len := data[26]

		if len < 0x10 {
			dummy.profile = string(data[27 : 27+len])
		}

	}

	//log.Print(dummy.pc)

	dummy.pc.WriteTo(data, l.GameAddr)

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
