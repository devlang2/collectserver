package event

import (
	"net"
	"time"

	"github.com/nanobox-io/golang-syslogparser"
)

type Event struct {
	Data   syslogparser.LogParts
	Origin string
	Rdate  time.Time
	Addr   *net.UDPAddr
}

func NewEvent() *Event {
	return &Event{Rdate: time.Now()}
}
