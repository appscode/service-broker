package server

import (
	"context"
	"flag"
	"fmt"
	"github.com/spf13/pflag"
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

type BrokerServerOptions struct {
	ExtraOptions *broker.ExtraOptions

	Port     int
	TLSCert  string
	TLSKey   string
	Insecure bool
}

func NewBrokerServerOptions() *BrokerServerOptions {
	return &BrokerServerOptions{
		ExtraOptions: broker.NewExtraOptions(),

		Port:     8080,
		Insecure: false,
	}
}

func (o *BrokerServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&o.Port, "port", o.Port, "use '--port' option to specify the port for broker to listen on.")
	fs.BoolVar(&o.Insecure, "insecure", o.Insecure,
		"use --insecure to use HTTP vs HTTPS.")
	fs.StringVar(&o.TLSCert, "tlsCert", o.TLSCert,
		"base-64 encoded PEM block to use as the certificate for TLS. If '--tlsCert' is used, then '--tlsKey' must also be used. If '--tlsCert' is not used, then TLS will not be used.")
	fs.StringVar(&o.TLSKey, "tlsKey", o.TLSKey,
		"base-64 encoded PEM block to use as the private key matching the TLS certificate. If '--tlsKey' is used, then '--tlsCert' must also be used.")

	o.ExtraOptions.AddFlags(fs)
}

func (o BrokerServerOptions) Validate(args []string) error {
	return nil
}

func (o *BrokerServerOptions) Complete() error {
	return nil
}

func (o BrokerServerOptions) Run() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go cancelOnInterrupt(ctx, cancelFunc)

	if err := o.runWithContext(ctx); err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		return err
	}

	return nil
}

func (o BrokerServerOptions) runWithContext(ctx context.Context) error {
	if flag.Arg(0) == "version" {
		fmt.Printf("%s/%s\n", path.Base(os.Args[0]), "0.1.0")
		return nil
	}
	if (o.TLSCert != "" || o.TLSKey != "") &&
		(o.TLSCert == "" || o.TLSKey == "") {
		fmt.Println("To use TLS, both --tlsCert and --tlsKey must be used")
		return nil
	}

	addr := ":" + strconv.Itoa(o.Port)

	glog.Infoln("broker client creating...")
	b, err := broker.NewBroker(o.ExtraOptions)
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

	if o.TLSCert == "" && o.TLSKey == "" {
		err = s.Run(ctx, addr)
	} else {
		err = s.RunTLS(ctx, addr, o.TLSCert, o.TLSKey)
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
