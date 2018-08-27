package db_broker

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	InstanceLabel = "dbbroker.instance"
)

type Client struct {
	namespace  string
	kubeClient kubernetes.Interface
	providers  map[string]Provider
}

func NewClient(kubeConfigPath, storageClassName string) *Client {
	config := getConfig(kubeConfigPath)
	return &Client{
		kubeClient: loadInClusterClient(config),
		namespace:  loadNamespace(kubeConfigPath),
		providers: map[string]Provider{
			"mysql":         NewMySQLProvider(config, storageClassName),
			"postgresql":    NewPostgreSQLProvider(config, storageClassName),
			"elasticsearch": NewElasticsearchProvider(config, storageClassName),
			"mongodb":       NewMongoDbProvider(config, storageClassName),
			"redis":         NewRedisProvider(config, storageClassName),
			"memcached":     NewMemcachedProvider(config),
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
		if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
			if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
				glog.Infof("namespace: %s", ns)
				return ns
			}
		}
	}

	panic("could not detect current namespace")
}

func (c *Client) Provision(serviceID, dbObjName, namespace string, provisionParams map[string]interface{}) error {
	glog.Infof("getting provider %q", serviceID)
	provider, ok := c.providers[serviceID]
	if !ok {
		return errors.Errorf("No %q provider found", serviceID)
	}

	glog.Infof("creating %s obj %q in namespace %q", serviceID, dbObjName, namespace)
	if err := provider.Create(dbObjName, c.namespace); err != nil {
		return errors.Wrapf(err, "failed to create %s obj %q in namespace", serviceID, dbObjName, namespace)
	}
	glog.Infoln("creation complete")

	return nil
}

func (c *Client) Bind(
	dbObjName, serviceID, planID string,
	bindParams, provisionParams map[string]interface{}) (map[string]interface{}, error) {

	params := make(map[string]interface{}, len(bindParams)+len(provisionParams))
	for k, v := range provisionParams {
		params[k] = v
	}
	for k, v := range bindParams {
		params[k] = v
	}

	service, err := c.kubeClient.CoreV1().Services(c.namespace).Get(dbObjName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	secret, err := c.kubeClient.CoreV1().Secrets(c.namespace).Get(dbObjName+"-auth", metav1.GetOptions{})
	if err == nil {
		for key, value := range secret.Data {
			data[key] = string(value)
		}
	} else {
		if !apierrs.IsNotFound(err) {
			return nil, err
		}
	}

	// Apply additional provisioning logic for Service Catalog Enabled services
	provider, ok := c.providers[serviceID]
	if ok {
		creds, err := provider.Bind(*service, params, data)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to bind instance for %q/%q", serviceID, planID)
		}
		for k, v := range creds.ToMap() {
			data[k] = v
		}
	}

	return data, nil
}

func (c *Client) Deprovision(serviceID, dbObjName string) error {
	glog.Infof("getting provider for %q", serviceID)
	provider, ok := c.providers[serviceID]
	if !ok {
		return errors.Errorf("No %q provider found", serviceID)
	}

	fmt.Printf("deleting %s obj %q from namespace %q...", serviceID, dbObjName, c.namespace)
	if err := provider.Delete(dbObjName, c.namespace); err != nil {
		return errors.Wrapf(err, "failed to delete %s obj %q from namespace %q", serviceID, dbObjName, c.namespace)
	}

	glog.Infoln("deletion complete")

	return nil
}
