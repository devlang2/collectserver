package collectors

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/devlang2/collectserver/event"
)

type Collector interface {
	Start(chan<- *event.Event) error
	Addr() net.Addr
}

func NewCollector(proto, addr string, tlsConfig *tls.Config) (Collector, error) {
	if strings.ToLower(proto) == "tcp" {

	} else if strings.ToLower(proto) == "udp" {
		addr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return nil, err
		}

		return &UDPCollector{addr: addr}, nil
	}
	return nil, fmt.Errorf("unsupport collector protocol")
}
