package server

import (
	"github.com/appscode/service-broker/pkg/broker"
	dbsvc "github.com/appscode/service-broker/pkg/kubedb"
	svcat_cs "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/spf13/pflag"
)

type ExtraOptions struct {
	CatalogPath  string
	CatalogNames []string
	Async        bool

	QPS   float64
	Burst int
}

func NewExtraOptions() *ExtraOptions {
	return &ExtraOptions{
		CatalogPath: "/etc/config/catalogs",
		Async:       false,
		QPS:         100,
		Burst:       100,
	}
}

func (s *ExtraOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.CatalogPath, "catalog-path", s.CatalogPath, "The path to the catalog.")
	fs.StringSliceVar(&s.CatalogNames, "catalog-names", s.CatalogNames,
		"List of catalogs those can be run by this service-broker, comma separated.")
	fs.BoolVar(&s.Async, "async", s.Async, "Indicates whether the broker is handling the requests asynchronously.")

	fs.Float64Var(&s.QPS, "qps", s.QPS, "The maximum QPS to the master from this client")
	fs.IntVar(&s.Burst, "burst", s.Burst, "The maximum burst for throttle")
}

func (s *ExtraOptions) ApplyTo(cfg *broker.Config) error {
	var err error

	cfg.ClientConfig.QPS = float32(s.QPS)
	cfg.ClientConfig.Burst = s.Burst

	cfg.CatalogPath = s.CatalogPath
	cfg.CatalogNames = s.CatalogNames
	cfg.Async = s.Async

	cfg.DBClient = dbsvc.NewClient(cfg.ClientConfig)
	if cfg.SvcCatClient, err = svcat_cs.NewForConfig(cfg.ClientConfig); err != nil {
		return err
	}
	return nil
}
