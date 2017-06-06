package shard

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"

	"bytes"

	"github.com/sirupsen/logrus"
)

type UDPLink struct {
	parent     *Portal
	pcHost     net.PacketConn
	pcGuest    net.PacketConn
	clientAddr net.Addr
	hostAddr   net.Addr
	active     time.Time
}

func NewUDPLink(parent *Portal, clientAddr net.Addr) *UDPLink {

	l := new(UDPLink)
	l.parent = parent
	l.clientAddr = clientAddr

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

	l.pcHost.Close()
	l.pcGuest.Close()

}

func (l *UDPLink) handleHostConnection() {

	buf := make([]byte, 256)

	for {

		n, addr, err := l.pcHost.ReadFrom(buf)

		if err != nil {
			l.parent.parent.logger.WithFields(logrus.Fields{
				"scope": "udpLink/handleHostConnection",
			}).Warn(err)
			break
		}

		if bytes.Equal(buf[:n], []byte("PHANTOM")) {
			if l.hostAddr == nil {
				l.hostAddr = addr
				l.parent.parent.logger.WithFields(logrus.Fields{
					"scope": "udpLink/handleHostConnection",
				}).Info("host bound")
			}
		} else {
			guestAddr, data := l.unpackData(buf[:n])
			l.pcGuest.WriteTo(data, guestAddr)
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

		udpAddr, ok := addr.(*net.UDPAddr)

		if !ok {
			l.parent.parent.logger.WithFields(logrus.Fields{
				"scope": "udpLink/handleGuestConnection",
			}).Warn("ERROR_ADDR_INVALID")
			break
		}

		l.pcHost.WriteTo(l.packData(udpAddr, buf[:n]), l.hostAddr)

	}

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
