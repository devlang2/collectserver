package collectors

import (
	//	"bufio"
	//	"bytes"
	"crypto/tls"
	"expvar"
	//	"io"
	"encoding/gob"
	"net"
	"time"
	//	"github.com/davecgh/go-spew/spew"
	"github.com/devlang2/collectserver/event"
	log "github.com/sirupsen/logrus"
)

var stats = expvar.NewMap("tcp")

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
	log.Println("Conn open")
	stats.Add("tcpConnections", 1)
	defer func() {
		stats.Add("tcpConnections", -1)
		conn.Close()
	}()
	for {
		decoder := gob.NewDecoder(conn)
		agents := make([]event.Agent, 0, 3)
		err := decoder.Decode(&agents)
		if err != nil {
			log.Error(err.Error())
			return
		}
		stats.Add("tcpEventsRx", int64(len(agents)))
		go func() {
			log.Printf("Inserting..")
			time.Sleep(10 * time.Second)
			log.Printf("Insert complete")
		}()

		//        encoder := gob.NewEncoder(conn)
		//		spew.Dump(agents)
	}
}

//func (s *TCPCollector) handleConnection1(conn net.Conn, c chan<- *event.Event) {
//	stats.Add("tcpConnections", 1)
//	defer func() {
//		stats.Add("tcpConnections", -1)
//		conn.Close()
//	}()

//	decoder := gob.NewDecoder(conn)
//	var agents [2]event.Agent
//	err := decoder.Decode(&agents)
//	if err != nil {
//		log.Error(err.Error())
//	}
//	spew.Dump(agents)
//}

//func (s *TCPCollector) handleConnection(conn net.Conn, c chan<- *event.Event) {
//	stats.Add("tcpConnections", 1)
//	defer func() {
//		stats.Add("tcpConnections", -1)
//		conn.Close()
//	}()
//	reader := bufio.NewReader(conn)
//	data := event.NewAgentData(msgBufSize)

//	for {
//		conn.SetReadDeadline(time.Now().Add(tcpTimeout))
//		b, err := reader.ReadByte()
//		if err != nil {
//			stats.Add("tcpConnReadError", 1)
//			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
//				stats.Add("tcpConnReadTimeout", 1)
//			} else if err == io.EOF {
//				stats.Add("tcpConnReadEOF", 1)
//			} else {
//				stats.Add("tcpConnUnrecoverError", 1)
//				return
//			}

//			// Parse event
//			//
//			// agent, err := data.decode()
//			Decode()

//		} else {
//			stats.Add("tcpBytesRead", 1)
//			data.Push(b)

//			//            data.
//			// Parse Event
//			//			//			spew.Dump(b)
//			//			//			spew.Dump(string(b))
//			//			log, match = delimiter.Push(b)
//		}
//		//		//		// Log line available?
//		//		//		spew.Dump(match)
//		//		if match {
//		//			spew.Println("****** " + err.Error())
//		//			spew.Dump(bytes.NewBufferString(log))
//		//			//			stats.Add("tcpEventsRx", 1)
//		//			//			if parser.Parse(bytes.NewBufferString(log).Bytes()) {
//		//			//				c <- &Event{
//		//			//					Text:          string(parser.Raw),
//		//			//					Parsed:        parser.Result,
//		//			//					ReceptionTime: time.Now().UTC(),
//		//			//					Sequence:      atomic.AddInt64(&sequenceNumber, 1),
//		//			//					SourceIP:      conn.RemoteAddr().String(),
//		//			//				}
//		//			//			}
//		//		}
//		// Was the connection closed?
//		if err == io.EOF {
//			return
//		}
//	}

//	//	parser, err := NewParser(s.format)
//	//	if err != nil {
//	//		panic(fmt.Sprintf("failed to create TCP connection parser:%s", err.Error()))
//	//	}
//	//	delimiter := NewSyslogDelimiter(msgBufSize)
//	//	reader := bufio.NewReader(conn)
//	//	var log string
//	//	var match bool
//	//	for {
//	//		conn.SetReadDeadline(time.Now().Add(newlineTimeout))
//	//		b, err := reader.ReadByte()
//	//		if err != nil {
//	//			stats.Add("tcpConnReadError", 1)
//	//			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
//	//				stats.Add("tcpConnReadTimeout", 1)
//	//			} else if err == io.EOF {
//	//				stats.Add("tcpConnReadEOF", 1)
//	//			} else {
//	//				stats.Add("tcpConnUnrecoverError", 1)
//	//				//				return
//	//			}
//	//			log, match = delimiter.Vestige()
//	//		} else {
//	//			stats.Add("tcpBytesRead", 1)
//	//			//			spew.Dump(b)
//	//			//			spew.Dump(string(b))
//	//			log, match = delimiter.Push(b)
//	//		}
//	//		//		// Log line available?
//	//		//		spew.Dump(match)
//	//		if match {
//	//			spew.Println("****** " + err.Error())
//	//			spew.Dump(bytes.NewBufferString(log))
//	//			//			stats.Add("tcpEventsRx", 1)
//	//			//			if parser.Parse(bytes.NewBufferString(log).Bytes()) {
//	//			//				c <- &Event{
//	//			//					Text:          string(parser.Raw),
//	//			//					Parsed:        parser.Result,
//	//			//					ReceptionTime: time.Now().UTC(),
//	//			//					Sequence:      atomic.AddInt64(&sequenceNumber, 1),
//	//			//					SourceIP:      conn.RemoteAddr().String(),
//	//			//				}
//	//			//			}
//	//		}
//	//		// Was the connection closed?
//	//		if err == io.EOF {
//	//			return
//	//		}
//	//	}
//}
