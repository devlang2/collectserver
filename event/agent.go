package event

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	Time               time.Time
	Guid               uuid.UUID // AD2BDBE0-BB14-4CBA-A1A4-F9CFD096774F
	IP                 net.IP    // IP
	Mac                string    // MAC
	ComputerName       string    // WSAHN-PC
	OsVersionNumber    float32   // 10.0
	OsIsServer         int       // 0
	OsBit              int       // 64
	FullPolicyVersion  string    // 1026
	TodayPolicyVersion string    // 1028
}

//func NewAgent() *Agent {
//	return &Agent{Rdate: time.Now()}
//}
