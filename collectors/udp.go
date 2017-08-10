package collectors

import (
	"log"
	"net"

	"github.com/davecgh/go-spew/spew"
	"github.com/devlang2/collectserver/event"
	"github.com/nanobox-io/golang-syslogparser/rfc5424"
)

const (
	msgBufSize = 256
)

type UDPCollector struct {
	format string
	addr   *net.UDPAddr
}

func (s *UDPCollector) Start(c chan<- *event.Event) error {
	spew.Dump()
	conn, err := net.ListenUDP("udp", s.addr)
	if err != nil {
		return err
	}

	go func() {
		buf := make([]byte, msgBufSize)
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Read error: " + err.Error())
				continue
			}

			p := rfc5424.NewParser(buf[:n])
			err = p.Parse()
			if err != nil {
				log.Printf("Parse error: " + err.Error())
				continue
			}

			event := event.NewEvent()
			event.Data = p.Dump()
			event.Addr = addr
			event.Origin = string(buf[:n]) // Original message

			c <- event
		}
	}()
	return nil
}

func (s *UDPCollector) Addr() net.Addr {
	return s.addr
}
