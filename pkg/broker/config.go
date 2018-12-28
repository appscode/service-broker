package broker

import (
	dbsvc "github.com/appscode/service-broker/pkg/kubedb"
	svcat_cs "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"k8s.io/client-go/rest"
)

type config struct {
	CatalogPath  string
	CatalogNames []string
	Async        bool
}

type Config struct {
	config

	ClientConfig *rest.Config
	DBClient     *dbsvc.Client
	SvcCatClient svcat_cs.ServicecatalogV1beta1Interface
}

func NewConfig(clientConfig *rest.Config) *Config {
	return &Config{
		ClientConfig: clientConfig,
	}
}

func (c *Config) New() (*Broker, error) {
	return &Broker{
		dbClient:     c.DBClient,
		svccatClient: c.SvcCatClient,
		async:        c.Async,
		catalogPath:  c.CatalogPath,
		catalogNames: c.CatalogNames,
	}, nil
}
