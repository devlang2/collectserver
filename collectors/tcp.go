package collectors

import (
	"crypto/tls"
	"encoding/gob"
	"expvar"
	"io"
	"net"

	"github.com/devplayg/tcpserver/event"
	log "github.com/sirupsen/logrus"
)

var stats = expvar.NewMap("tcp")

type TCPCollector struct {
	addrStr   string
	addr      net.Addr
	tlsConfig *tls.Config
}

type Result struct {
	Code    int
	Message string
}

func (this *TCPCollector) Start(c chan<- *event.Event) error {
	var ln net.Listener
	var err error
	if this.tlsConfig == nil {
		ln, err = net.Listen("tcp", this.addrStr)
	} else {
		ln, err = tls.Listen("tcp", this.addrStr, this.tlsConfig)
	}
	if err != nil {
		return err
	}
	this.addr = ln.Addr()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			go this.handleConnection(conn, c)
		}
	}()
	return nil
}

func (this *TCPCollector) Addr() net.Addr {
	return this.addr
}

func (this *TCPCollector) handleConnection(conn net.Conn, c chan<- *event.Event) {
	stats.Add("tcpConnections", 1)
	log.Info("Connected from ", conn.RemoteAddr().String())
	defer func() {
		stats.Add("tcpConnections", -1)
		conn.Close()
	}()

	host, port, _ := net.SplitHostPort(conn.RemoteAddr().String())
	ip := net.ParseIP(host)
	for {
		decoder := gob.NewDecoder(conn)
		events := make([]event.Event, 0, 3)
		err := decoder.Decode(&events)

		if err != nil {
			stats.Add("tcpDecodeError", 1)
			if err != io.EOF {
				log.Error(err.Error())
			}

			return
		}
		for i, _ := range events {
			events[i].SrcIP = ip
			events[i].SrcPort = port

			c <- &events[i]
		}
		stats.Add("tcpEventsRx", int64(len(events)))

	}
}
