package engine

import (
	"database/sql"
	"expvar"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/devlang2/collectserver/event"
	"github.com/devlang2/golibs/network"
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
	//	initDatabase("root:sniper123!@#@tcp(127.0.0.1:3306)/awserver?charset=utf8&allowAllFiles=true")
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

			result, err := db.Exec(fmt.Sprintf("LOAD DATA LOCAL INFILE %q INTO TABLE syslog", file.Name()))
			if err != nil {
				errChan <- err
				return
			}

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
			stats.Add("eventsCollected", int64(len(queue)))
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
				log.Println("timeout")
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
	tmpfile, err := ioutil.TempFile(datadir, "syslog_"+time.Now().Format("20060102_150405_"))
	defer tmpfile.Close()
	if err != nil {
		return tmpfile, err
	}
	var str string
	for _, r := range queue {
		str += fmt.Sprintf("%s\t%s\t%d\t%d\t%s\t%s\t%d\t%d\t%s\t%d\t%d\t%s\t%s\t%s\t%s\n",
			r.Data["timestamp"].(time.Time).Format(YYYYMMDDHH24MISS), // timestamp
			r.Rdate.Format(YYYYMMDDHH24MISS),                         // rdate
			network.IpToInt32(r.Addr.IP),                             // src_ip
			r.Addr.Port,                                              // port
			r.Data["hostname"].(string),                              // hostname
			r.Data["proc_id"].(string),                               // proc_id
			r.Data["facility"].(int),                                 // facility
			r.Data["severity"].(int),                                 // severity
			r.Data["app_name"].(string),                              // app_name
			r.Data["priority"].(int),                                 // priority
			r.Data["version"].(int),                                  // version
			r.Data["msg_id"].(string),                                // msg_id
			r.Data["message"].(string),                               // message
			r.Data["structured_data"].(string),                       // structured_data
			r.Origin, // origin
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
