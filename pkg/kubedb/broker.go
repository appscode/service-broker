package kubedb

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
	appcat_util "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1/util"
)

type Client struct {
	kubeClient kubernetes.Interface
	appClient  appcat_cs.AppcatalogV1alpha1Interface

	serviceProviders map[string]Provider
}

func NewClient(kubeConfigPath, storageClassName string) *Client {
	config := getConfig(kubeConfigPath)
	return &Client{
		kubeClient: kubernetes.NewForConfigOrDie(config),
		appClient:  appcat_cs.NewForConfigOrDie(config),
		serviceProviders: map[string]Provider{
			KubeDBServiceMySQL:         NewMySQLProvider(config, storageClassName),
			KubeDBServicePostgreSQL:    NewPostgreSQLProvider(config, storageClassName),
			KubeDBServiceElasticsearch: NewElasticsearchProvider(config, storageClassName),
			KubeDBServiceMongoDB:       NewMongoDbProvider(config, storageClassName),
			KubeDBServiceRedis:         NewRedisProvider(config, storageClassName),
			KubeDBServiceMemcached:     NewMemcachedProvider(config),
		},
	}
}

// TODO: create once
func getConfig(kubeConfigPath string) *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		panic(err)
	}
	config.Burst = 100
	config.QPS = 100

	return config
}

func (c *Client) GetCatalog(catalogPath string, catalogNames ...string) ([]osb.Service, error) {
	glog.Infoln("Listing services for catalog...")

	catalogs := sets.NewString(catalogNames...)

	var services []osb.Service
	for _, provider := range c.serviceProviders {
		catalog, serviceName := provider.Metadata()
		if catalogs.Has(catalog) {
			out, err := ioutil.ReadFile(filepath.Join(catalogPath, catalog, fmt.Sprintf("%s.yaml", serviceName)))
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

	glog.Infoln("Service list has been completed for catalog")
	return services, nil
}

func (c *Client) Provision(provisionInfo ProvisionInfo) error {
	glog.Infof("getting provider %q", provisionInfo.ServiceID)

	provider, exists := c.serviceProviders[provisionInfo.ServiceID]
	if !exists {
		return errors.Errorf("No %q provider found", provisionInfo.ServiceID)
	}

	if err := provider.Create(provisionInfo); err != nil {
		return errors.Wrapf(err, "failed to create %s obj %q in namespace %s",
			provisionInfo.ServiceID, provisionInfo.InstanceName, provisionInfo.Namespace)
	}

	return nil
}

func (c *Client) GetProvisionInfo(instanceID, serviceID string) (*ProvisionInfo, error) {
	provider, exists := c.serviceProviders[serviceID]
	if !exists {
		return nil, errors.Errorf("No %q provider found", serviceID)
	}

	return provider.GetProvisionInfo(instanceID)
}

func (c *Client) Bind(
	serviceID, planID string, bindParams map[string]interface{},
	provisionInfo ProvisionInfo) (map[string]interface{}, error) {

	params := make(map[string]interface{}, len(bindParams)+len(provisionInfo.Params))
	for k, v := range provisionInfo.Params {
		params[k] = v
	}
	for k, v := range provisionInfo.ExtraParams {
		params[k] = v
	}
	for k, v := range bindParams {
		params[k] = v
	}

	app, err := c.appClient.AppBindings(provisionInfo.Namespace).Get(provisionInfo.InstanceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	if app.Spec.Secret != nil {
		secret, err := c.kubeClient.CoreV1().Secrets(provisionInfo.Namespace).Get(app.Spec.Secret.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		for key, value := range secret.Data {
			data[key] = string(value)
		}
		err = appcat_util.TransformCredentials(c.kubeClient, app.Spec.SecretTransforms, data)
		if err != nil {
			return nil, err
		}
	}
	if len(app.Spec.ClientConfig.CABundle) > 0 {
		data["root.pem"] = app.Spec.ClientConfig.CABundle
	}

	// Apply additional provisioning logic for Service Catalog Enabled services
	provider, exists := c.serviceProviders[serviceID]
	if !exists {
		return nil, errors.Errorf("No %q provider found", serviceID)
	}

	creds, err := provider.Bind(app, params, data)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to bind instance for %q/%q", serviceID, planID)
	}

	return creds.ToMap()
}

func (c *Client) Deprovision(serviceID, instanceName, namespace string) error {
	glog.Infof("getting provider for %q", serviceID)

	provider, exists := c.serviceProviders[serviceID]
	if !exists {
		return errors.Errorf("No %q provider found", serviceID)
	}

	if err := provider.Delete(instanceName, namespace); err != nil {
		return errors.Wrapf(err, "failed to delete %s obj %q from namespace %q", serviceID, instanceName, namespace)
	}

	return nil
}
