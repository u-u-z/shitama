package shard

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
)

type Echo struct {
	parent *Shard
	pc     net.PacketConn
}

func NewEcho(parent *Shard) *Echo {

	e := new(Echo)
	e.parent = parent

	return e

}

func (e *Echo) Start() {

	echoPort := e.parent.Config["echoPort"].(int)

	pc, err := net.ListenPacket("udp4", fmt.Sprintf("0.0.0.0:%d", echoPort))

	if err != nil {
		e.parent.logger.WithFields(logrus.Fields{
			"scope": "echo/Start",
		}).Fatal(err)
	}

	e.pc = pc

	go e.handle()

}

func (e *Echo) handle() {

	buf := make([]byte, 256)

	for {

		n, addr, err := e.pc.ReadFrom(buf)

		if err != nil {
			e.parent.logger.WithFields(logrus.Fields{
				"scope": "echo/handle",
			}).Warn(err)
		}

		//log.Printf("echo server receives %d bytes from %s", n, addr.String())

		e.pc.WriteTo(buf[0:n], addr)

	}

}
