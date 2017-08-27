package main

// If failed to conn. to db, handle error well
import (
	"crypto/x509"
	"expvar"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
	//	"log"
	//	//	"net"
	"crypto/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
	//	"io"
	"crypto/tls"
	"io/ioutil"

	"github.com/devplayg/tcpserver/collectors"
	"github.com/devplayg/tcpserver/engine"
	//	//	"github.com/iwondory/udpserver/event"
	//	"github.com/davecgh/go-spew/spew"
)

const (
	DefaultBatchSize       = 1000
	DefaultBatchDuration   = 5000 // ms
	DefaultBatchMaxPending = 1000
	DefaultCPUCount        = 2
	DefaultDataDir         = "./temp"
	DefaultTCPAddr         = "localhost:8808"
	DefaultMonitorAddr     = "localhost:8080"
)

var (
	stats = expvar.NewMap("server")
	fs    *flag.FlagSet
)

func init() {

	// Set logger
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		DisableColors: true,
	})
	log.SetLevel(log.InfoLevel)
	//	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	//	if err == nil {
	//		log.SetOutput(file)
	//	} else {
	//		log.Error("Failed to log to file, using default stderr")
	//		log.SetOutput(os.Stdout)
	//	}
}

func main() {
	// Set flags
	fs = flag.NewFlagSet("", flag.ExitOnError)
	var (
		batchSize       = fs.Int("batchsize", DefaultBatchSize, "Batch size")
		cpuCount        = fs.Int("cpu", DefaultCPUCount, "CPU count")
		batchDuration   = fs.Int("duration", DefaultBatchDuration, "Batch timeout, in milliseconds")
		batchMaxPending = fs.Int("maxpending", DefaultBatchMaxPending, "Maximum pending events")
		datadir         = fs.String("datadir", DefaultDataDir, "Set data directory")
		tcpAddr         = fs.String("tcpaddr", DefaultTCPAddr, "Syslog server TCP bind address in the form host:port")
		monAddr         = fs.String("monaddr", DefaultMonitorAddr, "Monitor bind address in the form host:port")
		caPemPath       = fs.String("tlspem", "", "path to CA PEM file for TLS-enabled TCP server. If not set, TLS not activated")
		caKeyPath       = fs.String("tlskey", "", "path to CA key file for TLS-enabled TCP server. If not set, TLS not activated")
		isDebug         = fs.Bool("debug", false, "Is debug mode?")
	)
	fs.Usage = printHelp
	fs.Parse(os.Args[1:])

	if *cpuCount > runtime.NumCPU() {
		*cpuCount = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(*cpuCount)

	log.Infof("CPU: %d/%d, Batch size: %d, Batch timeout: %d(ms), Batch max pending: %d", *cpuCount, runtime.NumCPU(), *batchSize, *batchDuration, *batchMaxPending)
	if *isDebug {
		log.SetLevel(log.DebugLevel)
		log.Info("Starting server..[Debug mode]")
	} else {
		log.Info("Starting server..")
	}

	// Start engine
	duration := time.Duration(*batchDuration) * time.Millisecond
	batcher := engine.NewBatcher(duration, *batchSize, *batchMaxPending, *datadir)
	errChan := make(chan error)
	if err := batcher.Start(errChan); err != nil {
		log.Fatalf("Failed to start batcher: %s", err.Error())
	}
	log.Info("Batcher is started")

	// Start log drain
	go logDrain("error", errChan)

	// Start TCP collector
	var tlsConfig *tls.Config
	var err error
	if *caPemPath != "" && *caKeyPath != "" {
		tlsConfig, err = newTLSConfig(*caPemPath, *caKeyPath)
		if err != nil {
			log.Fatalf("Failed to configure TLS: %s", err.Error())
		}
		log.Printf("TLS successfully configured")
	}

	if err := startTCPCollector(*tcpAddr, tlsConfig, batcher); err != nil {
		log.Fatalf("failed to start TCP collector: %s", err.Error())
	}
	log.Info("TCP collector is started")
	//	log.Printf("UDP collector listening to %s", *udpIface)

	// Start monitoring
	startStatusMonitoring(monAddr)

	// Stop
	waitForSignals()
}

func startStatusMonitoring(monitorIface *string) error {
	http.HandleFunc("/e", func(w http.ResponseWriter, r *http.Request) {
	})

	go http.ListenAndServe(*monitorIface, nil)
	return nil
}

func logDrain(msg string, errChan <-chan error) {
	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.Printf("%s: %s", msg, err.Error())
			}
		}
	}
}

func startTCPCollector(addr string, tls *tls.Config, batcher *engine.Batcher) error {
	collector, err := collectors.NewCollector("tcp", addr, tls)
	if err != nil {
		return fmt.Errorf(("failed to create TCP collector: %s"), err.Error())
	}
	if err := collector.Start(batcher.C()); err != nil {
		return fmt.Errorf("failed to start TCP collector: %s", err.Error())
	}

	return nil
}

func startUDPCollector(addr string, batcher *engine.Batcher) error {
	collector, err := collectors.NewCollector("udp", addr, nil)
	if err != nil {
		return fmt.Errorf("failed to create UDP collector: %s", err.Error())
	}
	if err := collector.Start(batcher.C()); err != nil {
		return fmt.Errorf("failed to start UDP collector: %s", err.Error())
	}

	return nil
}

func printHelp() {
	fmt.Println("dataserver [options]")
	fs.PrintDefaults()
}

func newTLSConfig(caPemPath, caKeyPath string) (*tls.Config, error) {
	var config *tls.Config

	caPem, err := ioutil.ReadFile(caPemPath)
	if err != nil {
		return nil, err
	}
	ca, err := x509.ParseCertificate(caPem)
	if err != nil {
		return nil, err
	}

	caKey, err := ioutil.ReadFile(caKeyPath)
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKCS1PrivateKey(caKey)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AddCert(ca)

	cert := tls.Certificate{
		Certificate: [][]byte{caPem},
		PrivateKey:  key,
	}

	config = &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
	}

	config.Rand = rand.Reader

	return config, nil
}

func waitForSignals() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalCh:
		log.Println("Signal received, shutting down...")
	}
}

//drop table log_agent;
//create table log_agent (

//	Time               datetime not null,
//	Guid               char(36) not null,
//	IP                 int unsigned not null,
//	Mac                char(17) not null,
//	ComputerName       varchar(64) not null,
//	OsVersionNumber    float(5,1),
//	OsIsServer         tinyint unsigned not null,
//	OsBit              tinyint unsigned not null,
//	FullPolicyVersion  int unsigned not null,
//	TodayPolicyVersion int unsigned not null,
//	Sequence           int unsigned not null,

//	SrcIP int unsigned not null,
//	SrcPort smallint unsigned not null,
//	RegDate datetime not null  DEFAULT CURRENT_TIMESTAMP,

//	key ix_time(time)

//)
