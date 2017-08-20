package engine

import (
	"database/sql"
	"expvar"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/devlang2/golibs/network"
	"github.com/devlang2/tcpserver/event"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

var (
	stats = expvar.NewMap("engine")
	db    *sql.DB
)

const (
	YYYYMMDDHH24MISS = "2006-01-02 15:04:05"
)

type Batcher struct {
	duration time.Duration
	size     int
	datadir  string

	c chan *event.Event
}

func init() {
	initDatabase("root:sniper123!@#@tcp(aptxa:3306)/awserver?charset=utf8&allowAllFiles=true")
}

func NewBatcher(duration time.Duration, size, maxpending int, datadir string) *Batcher {
	return &Batcher{
		duration: duration,
		size:     size,
		datadir:  datadir,
		c:        make(chan *event.Event, maxpending),
	}
}

func (this *Batcher) Start(errChan chan<- error) error {
	// Create data directory
	if _, err := os.Stat(this.datadir); os.IsNotExist(err) {
		err := os.Mkdir(this.datadir, 0755)
		if err != nil {
			return err
		}
	}

	go func() {
		timer := time.NewTimer(this.duration)
		timer.Stop()

		queue := make([]*event.Event, 0, this.size)
		save := func() {
			file, err := saveAsFile(this.datadir, queue)
			if err != nil {
				errChan <- err
				return
			}

			result, err := db.Exec(fmt.Sprintf("LOAD DATA LOCAL INFILE %q INTO TABLE log_agent(Time,Guid,IP,Mac,ComputerName,OsVersionNumber,OsIsServer,OsBit,FullPolicyVersion,TodayPolicyVersion,Sequence,SrcIP,SrcPort)", file.Name()))
			if err != nil {
				errChan <- err
				return
			}
			//			time.Sleep(1 * time.Second)

			affectedRows, _ := result.RowsAffected()
			if int64(len(queue)) == affectedRows {
				err := os.Remove(file.Name())
				if err != nil {
					errChan <- err
				}
			} else {
				errChan <- err
			}

			// Load data
			stats.Add("eventsInserted", int64(len(queue)))
			queue = make([]*event.Event, 0, this.size)
		}

		for {
			select {
			case event := <-this.c:
				queue = append(queue, event)
				if len(queue) == 1 {
					timer.Reset(this.duration)
				}

				if len(queue) == this.size {
					timer.Stop()
					save()
				}

			case <-timer.C:
				save()
			}
		}

	}()
	return nil
}

func (b *Batcher) C() chan<- *event.Event {
	return b.c
}

func saveAsFile(datadir string, queue []*event.Event) (*os.File, error) {
	// Write the data in the queue to a file
	tmpfile, err := ioutil.TempFile(datadir, "log_"+time.Now().Format("20060102_150405_"))
	defer tmpfile.Close()
	if err != nil {
		return tmpfile, err
	}
	var str string
	for _, r := range queue {
		str += fmt.Sprintf("%s\t%s\t%d\t%s\t%s\t%4.1f\t%d\t%d\t%s\t%s\t%d\t%d\t%s\n",
			r.Time.Format(YYYYMMDDHH24MISS),
			r.Guid.String(),
			network.IpToInt32(r.IP),
			r.Mac,                      //                Mac
			r.ComputerName,             //                ComputerName
			r.OsVersionNumber,          //                OsVersionNumber
			r.OsIsServer,               //                OsIsServer
			r.OsBit,                    //                OsBit
			r.FullPolicyVersion,        //                FullPolicyVersion
			r.TodayPolicyVersion,       //                TodayPolicyVersion
			r.Sequence,                 //                Sequence
			network.IpToInt32(r.SrcIP), //                SrcIP
			r.SrcPort,                  //                SrcPort
		)
	}

	if _, err := tmpfile.WriteString(str); err != nil {
		return tmpfile, err
	}

	// Write data in file to database
	return tmpfile, nil
}

func initDatabase(dataSourceName string) {
	var err error
	db, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Panic(err)
	}

	if err = db.Ping(); err != nil {
		log.Panic(err)
	}

	db.SetMaxIdleConns(3)
	db.SetMaxOpenConns(3)
}
