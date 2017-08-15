package event

import (
	"net"
	"time"

	//	"github.com/davecgh/go-spew/spew"
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

//func (this *AgentData) Decode() (*Agent, error) {
//    var agents []*Agent
//     dec := gob.NewDecoder(conn)
//    p := &P{}
//    dec.Decode(p)
//    fmt.Printf("Received : %+v", p);
////    deco := gob.NewDecoder(&network)
//	//	spew.Dump(this.buffer)
//    //var q Q
//	//err = dec.Decode(&q)
//	//if err != nil {
//	//    log.Fatal("decode error:", err)
//	//}
//	return nil, nil
//}

//func NewAgent() *Agent {
//	return &Agent{Rdate: time.Now()}
//}
