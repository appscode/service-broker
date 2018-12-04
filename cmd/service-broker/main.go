package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"

	"github.com/appscode/service-broker/pkg/broker"
	"github.com/golang/glog"
	"github.com/pmorie/osb-broker-lib/pkg/metrics"
	"github.com/pmorie/osb-broker-lib/pkg/rest"
	"github.com/pmorie/osb-broker-lib/pkg/server"
	prom "github.com/prometheus/client_golang/prometheus"
)

var options struct {
	broker.Options

	Port     int
	TLSCert  string
	TLSKey   string
	Insecure bool
}

func init() {
	flag.IntVar(&options.Port, "port", 8005,
		"use '--port' option to specify the port for broker to listen on")
	flag.BoolVar(&options.Insecure, "insecure", false,
		"use --insecure to use HTTP vs HTTPS.")
	flag.StringVar(&options.TLSCert, "tlsCert", "",
		"base-64 encoded PEM block to use as the certificate for TLS. If '--tlsCert' is used, then '--tlsKey' must also be used. If '--tlsCert' is not used, then TLS will not be used.")
	flag.StringVar(&options.TLSKey, "tlsKey", "",
		"base-64 encoded PEM block to use as the private key matching the TLS certificate. If '--tlsKey' is used, then '--tlsCert' must also be used")

	broker.AddFlags(&options.Options)
	flag.Parse()
}

func main() {
	fmt.Println(os.LookupEnv("PWD"))
	if err := run(); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		glog.Fatalln(err)
	}
}

func run() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go cancelOnInterrupt(ctx, cancelFunc)

	return runWithContext(ctx)
}

func runWithContext(ctx context.Context) error {
	if flag.Arg(0) == "version" {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), "0.1.0")
		return nil
	}
	if (options.TLSCert != "" || options.TLSKey != "") &&
		(options.TLSCert == "" || options.TLSKey == "") {
		fmt.Println("To use TLS, both --tlsCert and --tlsKey must be used")
		return nil
	}

	addr := ":" + strconv.Itoa(options.Port)

	glog.Infoln("broker client creating...")
	b, err := broker.NewBroker(options.Options)
	glog.Infoln("broker client created")

	if err != nil {
		return err
	}

	// Prometheus metrics
	reg := prom.NewRegistry()
	osbMetrics := metrics.New()
	reg.MustRegister(osbMetrics)

	api, err := rest.NewAPISurface(b, osbMetrics)
	if err != nil {
		return err
	}

	s := server.New(api, reg)

	glog.Infof("Starting broker!")

	if options.TLSCert == "" && options.TLSKey == "" {
		err = s.Run(ctx, addr)
	} else {
		err = s.RunTLS(ctx, addr, options.TLSCert, options.TLSKey)
	}
	return err
}

func cancelOnInterrupt(ctx context.Context, f context.CancelFunc) {
	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-term:
			glog.Infof("Received SIGTERM, exiting gracefully...")
			f()
			os.Exit(0)
		case <-ctx.Done():
			os.Exit(0)
		}
	}
}
