package db_broker

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	namespace  string
	kubeClient kubernetes.Interface

	catalogProviders map[string]map[string]Provider
}

func NewClient(kubeConfigPath, storageClassName string) *Client {
	config := getConfig(kubeConfigPath)
	return &Client{
		kubeClient: loadInClusterClient(config),
		namespace:  loadNamespace(kubeConfigPath),
		catalogProviders: map[string]map[string]Provider{
			CatelogKeyKubeDB: {
				ProviderNameMySQL:         NewMySQLProvider(config, storageClassName),
				ProviderNamePostgreSQL:    NewPostgreSQLProvider(config, storageClassName),
				ProviderNameElasticsearch: NewElasticsearchProvider(config, storageClassName),
				ProviderNameMongoDB:       NewMongoDbProvider(config, storageClassName),
				ProviderNameRedis:         NewRedisProvider(config, storageClassName),
				ProviderNameMemcached:     NewMemcachedProvider(config),
			},
		},
	}
}

func getConfig(kubeConfigPath string) *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		panic(err)
	}
	config.Burst = 100
	config.QPS = 100

	return config
}

func loadInClusterClient(config *rest.Config) kubernetes.Interface {
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return kubeClient
}

func loadNamespace(kubeConfigPath string) string {
	if kubeConfigPath != "" {
		return "default"
	} else {
		if data, err := ioutil.ReadFile(NamespaceFilePath); err == nil {
			if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
				glog.Infof("namespace: %s", ns)
				return ns
			}
		}
	}

	panic("could not detect current namespace")
}

func (c *Client) GetCatalog(catalogPath string, catalogNames ...string) ([]osb.Service, error) {
	glog.Infoln("Listing services for catalog...")

	services := []osb.Service{}
	for _, catalog := range catalogNames {
		if providers, ok := c.catalogProviders[catalog]; ok {
			for providerName := range providers {
				out, err := ioutil.ReadFile(filepath.Join(catalogPath, catalog, fmt.Sprintf("%s.yaml", providerName)))
				if err != nil {
					return nil, err
				}

				service := osb.Service{}
				if err = yaml.Unmarshal(out, &service); err != nil {
					return nil, err
				}
				services = append(services, service)
			}
		}
	}

	glog.Infoln("Service list has been completed for catalog")
	return services, nil
}

func (c *Client) Provision(
	catalogNames []string, provisionInfo ProvisionInfo) error {
	glog.Infof("getting provider %q", provisionInfo.ServiceID)
	var (
		provider Provider
		exists   bool
	)
	for _, catalog := range catalogNames {
		if providers, ok := c.catalogProviders[catalog]; ok {
			if provider, exists = providers[provisionInfo.ServiceID]; exists {
				break
			}
		}
	}
	if !exists {
		return errors.Errorf("No %q provider found", provisionInfo.ServiceID)
	}

	if err := provider.Create(provisionInfo, c.namespace); err != nil {
		return errors.Wrapf(err, "failed to create %s obj %q in namespace %s", provisionInfo.ServiceID, provisionInfo.InstanceName, c.namespace)
	}

	return nil
}

func (c *Client) GetProvisionInfo(catalogNames []string, instanceID, serviceID string) (*ProvisionInfo, error) {
	var (
		provider Provider
		exists   bool
	)
	for _, catalog := range catalogNames {
		if providers, ok := c.catalogProviders[catalog]; ok {
			if provider, exists = providers[serviceID]; exists {
				break
			}
		}
	}
	if !exists {
		return nil, errors.Errorf("No %q provider found", serviceID)
	}

	provisionInfo, err := provider.GetProvisionInfo(instanceID, c.namespace)
	if err != nil {
		return nil, err
	}

	return provisionInfo, nil
}

func (c *Client) Bind(
	catalogNames []string,
	serviceID, planID string, bindParams map[string]interface{},
	provisionInfo ProvisionInfo) (map[string]interface{}, error) {

	params := make(map[string]interface{}, len(bindParams)+len(provisionInfo.Params))
	for k, v := range provisionInfo.Params {
		params[k] = v
	}
	for k, v := range bindParams {
		params[k] = v
	}

	service, err := c.kubeClient.CoreV1().Services(c.namespace).Get(provisionInfo.InstanceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	secrets, err := c.kubeClient.CoreV1().Secrets(c.namespace).List(metav1.ListOptions{
		LabelSelector: labels.Set{
			api.LabelDatabaseName: provisionInfo.InstanceName,
		}.String(),
	})
	data := make(map[string]interface{})
	for _, secret := range secrets.Items {
		for key, value := range secret.Data {
			data[key] = string(value)
		}
	}

	// Apply additional provisioning logic for Service Catalog Enabled services
	var (
		provider Provider
		exists   bool
	)
	for _, catalog := range catalogNames {
		if providers, ok := c.catalogProviders[catalog]; ok {
			if provider, exists = providers[serviceID]; exists {
				break
			}
		}
	}
	if !exists {
		return nil, errors.Errorf("No %q provider found", serviceID)
	}

	creds, err := provider.Bind(*service, params, data)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to bind instance for %q/%q", serviceID, planID)
	}

	return creds.ToMap(), nil
}

func (c *Client) Deprovision(catalogNames []string, serviceID, instanceName string) error {
	glog.Infof("getting provider for %q", serviceID)
	var (
		provider Provider
		exists   bool
	)
	for _, catalog := range catalogNames {
		if providers, ok := c.catalogProviders[catalog]; ok {
			if provider, exists = providers[serviceID]; exists {
				break
			}
		}
	}
	if !exists {
		return errors.Errorf("No %q provider found", serviceID)
	}

	if err := provider.Delete(instanceName, c.namespace); err != nil {
		return errors.Wrapf(err, "failed to delete %s obj %q from namespace %q", serviceID, instanceName, c.namespace)
	}

	return nil
}
