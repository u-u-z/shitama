package shard

import (
	"fmt"
	"net"
	"strings"
	"time"

	"bytes"

	"github.com/evshiron/shitama/common"
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

		l.active = time.Now()

		if bytes.Equal(buf[:n], []byte("PHANTOM")) {
			if l.hostAddr == nil {
				l.hostAddr = addr
				l.parent.parent.logger.WithFields(logrus.Fields{
					"scope": "udpLink/handleHostConnection",
				}).Info("host bound")
			}
		} else {
			guestAddr, data := common.UnpackData(buf[:n])
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

		l.pcHost.WriteTo(common.PackData(udpAddr, buf[:n]), l.hostAddr)

	}

}
