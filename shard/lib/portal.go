package shard

import (
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

type Link interface {
	Start()
	HostAddr() string
	GuestAddr() string
	Transport() string
	Expired() bool
	Stop()
}

type Portal struct {
	parent *Shard
	links  []Link
}

func NewPortal(parent *Shard) *Portal {

	p := new(Portal)
	p.parent = parent
	p.links = make([]Link, 0)

	return p

}

func (p *Portal) Start() {

	go p.recycle()

}

func (p *Portal) NewLink(clientAddr string, transport string) Link {

	switch transport {
	case "udp":

		addr, err := net.ResolveUDPAddr("udp4", clientAddr)

		if err != nil {
			p.parent.logger.WithFields(logrus.Fields{
				"scope": "portal/NewLink",
			}).Fatal(err)
		}

		link := NewUDPLink(p, addr)
		p.links = append(p.links, link)
		link.Start()

		return link

	case "kcp":
	case "tcp":
	default:
		return nil
	}

	return nil

}

func (p *Portal) recycle() {

	for {

		//log.Print("recycle")

		for i := len(p.links) - 1; i >= 0; i-- {
			link := p.links[i]
			if link.Expired() {
				p.parent.logger.WithFields(logrus.Fields{
					"scope":    "portal/recycle",
					"linkAddr": link.HostAddr(),
				}).Info("recycle link")
				link.Stop()
				p.links = append(p.links[:i], p.links[i+1:]...)
			}
		}

		time.Sleep(10 * time.Minute)

	}

}
