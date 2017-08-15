package collectors

import (
	"bufio"
	//	"bytes"
	"crypto/tls"
	"expvar"
	"io"
	"net"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/devlang2/collectserver/event"
)

var stats = expvar.NewMap("tcp")

type Data struct {
}

type TCPCollector struct {
	addrStr   string
	addr      net.Addr
	tlsConfig *tls.Config
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

func (s *TCPCollector) handleConnection(conn net.Conn, c chan<- *event.Event) {
	stats.Add("tcpConnections", 1)
	defer func() {
		stats.Add("tcpConnections", -1)
		conn.Close()
	}()
	reader := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(tcpTimeout))
		b, err := reader.ReadByte()
		spew.Dump(b)
		if err != nil {
			stats.Add("tcpConnReadError", 1)
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				stats.Add("tcpConnReadTimeout", 1)
			} else if err == io.EOF {
				stats.Add("tcpConnReadEOF", 1)
			} else {
				stats.Add("tcpConnUnrecoverError", 1)
				return
			}

			// Parse event
			//
			//
		} else {
			stats.Add("tcpBytesRead", 1)
			// Parse Event
			//			//			spew.Dump(b)
			//			//			spew.Dump(string(b))
			//			log, match = delimiter.Push(b)
		}
		//		//		// Log line available?
		//		//		spew.Dump(match)
		//		if match {
		//			spew.Println("****** " + err.Error())
		//			spew.Dump(bytes.NewBufferString(log))
		//			//			stats.Add("tcpEventsRx", 1)
		//			//			if parser.Parse(bytes.NewBufferString(log).Bytes()) {
		//			//				c <- &Event{
		//			//					Text:          string(parser.Raw),
		//			//					Parsed:        parser.Result,
		//			//					ReceptionTime: time.Now().UTC(),
		//			//					Sequence:      atomic.AddInt64(&sequenceNumber, 1),
		//			//					SourceIP:      conn.RemoteAddr().String(),
		//			//				}
		//			//			}
		//		}
		// Was the connection closed?
		if err == io.EOF {
			return
		}
	}

	//	parser, err := NewParser(s.format)
	//	if err != nil {
	//		panic(fmt.Sprintf("failed to create TCP connection parser:%s", err.Error()))
	//	}
	//	delimiter := NewSyslogDelimiter(msgBufSize)
	//	reader := bufio.NewReader(conn)
	//	var log string
	//	var match bool
	//	for {
	//		conn.SetReadDeadline(time.Now().Add(newlineTimeout))
	//		b, err := reader.ReadByte()
	//		if err != nil {
	//			stats.Add("tcpConnReadError", 1)
	//			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
	//				stats.Add("tcpConnReadTimeout", 1)
	//			} else if err == io.EOF {
	//				stats.Add("tcpConnReadEOF", 1)
	//			} else {
	//				stats.Add("tcpConnUnrecoverError", 1)
	//				//				return
	//			}
	//			log, match = delimiter.Vestige()
	//		} else {
	//			stats.Add("tcpBytesRead", 1)
	//			//			spew.Dump(b)
	//			//			spew.Dump(string(b))
	//			log, match = delimiter.Push(b)
	//		}
	//		//		// Log line available?
	//		//		spew.Dump(match)
	//		if match {
	//			spew.Println("****** " + err.Error())
	//			spew.Dump(bytes.NewBufferString(log))
	//			//			stats.Add("tcpEventsRx", 1)
	//			//			if parser.Parse(bytes.NewBufferString(log).Bytes()) {
	//			//				c <- &Event{
	//			//					Text:          string(parser.Raw),
	//			//					Parsed:        parser.Result,
	//			//					ReceptionTime: time.Now().UTC(),
	//			//					Sequence:      atomic.AddInt64(&sequenceNumber, 1),
	//			//					SourceIP:      conn.RemoteAddr().String(),
	//			//				}
	//			//			}
	//		}
	//		// Was the connection closed?
	//		if err == io.EOF {
	//			return
	//		}
	//	}
}
