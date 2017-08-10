package main

import (
	"crypto/x509"
	"expvar"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
	//	"log"
	//	//	"net"
	//	"net/http"
	"os"
	//	"os/signal"
	"runtime"
	//	"syscall"
	"crypto/rand"
	"time"
	//	"io"
	"crypto/tls"
	"io/ioutil"

	"github.com/devlang2/collectserver/collectors"
	"github.com/devlang2/collectserver/engine"
	//	//	"github.com/iwondory/udpserver/event"
	//	"github.com/davecgh/go-spew/spew"
)

const (
	DefaultBatchSize       = 4
	DefaultBatchDuration   = 3000 // ms
	DefaultBatchMaxPending = 3
	DefaultDataDir         = "./temp"
	DefaultTCPAddr         = "localhost:9001"
	DefaultUDPAddr         = "localhost:9002"
)

var (
	stats = expvar.NewMap("server")
	fs    *flag.FlagSet
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Set logger
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		DisableColors: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	log.Info("Starting server..")

	// Set flags
	fs = flag.NewFlagSet("", flag.ExitOnError)
	var (
		batchSize       = fs.Int("batchsize", DefaultBatchSize, "Batch size")
		batchDuration   = fs.Int("duration", DefaultBatchDuration, "Batch timeout, in milliseconds")
		batchMaxPending = fs.Int("maxpending", DefaultBatchMaxPending, "Maximum pending events")
		datadir         = fs.String("datadir", DefaultUDPAddr, "Set data directory")
		udpAddr         = fs.String("udpaddr", DefaultUDPAddr, "Syslog server UDP bind address in the form host:port")
		tcpAddr         = fs.String("tcpaddr", DefaultTCPAddr, "Syslog server TCP bind address in the form host:port")
		caPemPath       = fs.String("tlspem", "", "path to CA PEM file for TLS-enabled TCP server. If not set, TLS not activated")
		caKeyPath       = fs.String("tlskey", "", "path to CA key file for TLS-enabled TCP server. If not set, TLS not activated")
		isDebug         = fs.Bool("debug", false, "Is debug mode?")
	)
	fs.Usage = printHelp
	fs.Parse(os.Args[1:])
	if *isDebug {
		//		log.SetOutput(log.DebugLevel)
	}

	// Start engine
	duration := time.Duration(*batchDuration) * time.Millisecond
	batcher := engine.NewBatcher(duration, *batchSize, *batchMaxPending, *datadir)
	errChan := make(chan error)
	if err := batcher.Start(errChan); err != nil {
		log.Fatalf("failed to start batcher: %s", err.Error())
	}

	go logDrain("error", errChan)

	// Start UDP collector
	if err := startUDPCollector(*udpAddr, batcher); err != nil {

		//		log.Fatalf("failed to start UDP collector: %s", err.Error())
	}

	// Start TCP collector
	var tlsConfig *tls.Config
	var err error
	if *caPemPath != "" && *caKeyPath != "" {
		tlsConfig, err = newTLSConfig(*caPemPath, *caKeyPath)
		if err != nil {
			log.Fatalf("failed to configure TLS: %s", err.Error())
		}
		log.Printf("TLS successfully configured")
	}

	if err := startTCPCollector(*tcpAddr, tlsConfig, batcher); err != nil {
	}
	//	log.Printf("UDP collector listening to %s", *udpIface)

	//	// Start monitoring
	//	startStatusMonitoring(monitorIface)

	//	// Stop
	//	waitForSignals()
}

//func startStatusMonitoring(monitorIface *string) error {
//	http.HandleFunc("/e", func(w http.ResponseWriter, r *http.Request) {
//	})

//	go http.ListenAndServe(*monitorIface, nil)
//	return nil
//}

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
	//	collector, err := input.NewCollector("tcp", iface, format, tls)
	//	if err != nil {
	//		return fmt.Errorf(("failed to create TCP collector: %s"), err.Error())
	//	}
	//	if err := collector.Start(batcher.C()); err != nil {
	//		return fmt.Errorf("failed to start TCP collector: %s", err.Error())
	//	}

	return nil
}

//func startUDPCollector(iface, format string, batcher *ekanite.Batcher) error {
//	collector, err := input.NewCollector("udp", iface, format, nil)
//	if err != nil {
//		return fmt.Errorf("failed to create UDP collector: %s", err.Error())
//	}
//	if err := collector.Start(batcher.C()); err != nil {
//		return fmt.Errorf("failed to start UDP collector: %s", err.Error())
//	}

//	return nil
//}

//func waitForSignals() {
//	signalCh := make(chan os.Signal, 1)
//	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
//	select {
//	case <-signalCh:
//		log.Println("signal received, shutting down...")
//	}
//}

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
