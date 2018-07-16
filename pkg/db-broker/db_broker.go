package db_broker

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	InstanceLabel = "dbbroker.instance"
)

type Client struct {
	namespace  string
	coreClient kubernetes.Interface
	providers  map[string]Provider
}

func NewClient(KubeConfig string) *Client {
	ns := loadNamespace(KubeConfig)
	fmt.Println("broker namespace =", ns)
	return &Client{
		coreClient: loadInClusterClient(KubeConfig),
		namespace:  ns,
		providers: map[string]Provider{
			"mysql": NewMySQLProvider(KubeConfig),
			//"mariadb":    MariadbProvider{},
			"postgresql": NewPostgreSQLProvider(KubeConfig),
			//"mongodb":    MongodbProvider{},
		},
	}
}

func loadInClusterClient(kubeConfig string) kubernetes.Interface {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset
}

func loadNamespace(kubeConfig string) string {
	if kubeConfig != "" {
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

func (c *Client) Provision(serviceID, planID, namespace string, provisionParams map[string]interface{}) error {
	//glog.Infoln("getting provider for 'mysqldb'")
	glog.Infoln("getting provider %q", serviceID)
	//provider, ok := c.providers["mysqldb"]
	provider, ok := c.providers[serviceID]
	if !ok {
		return errors.Errorf("No %q provider found", serviceID)
	}

	glog.Infof("creating %q obj", planID)
	if err := provider.Create(planID, c.namespace); err != nil {
		return err
		errors.Wrapf(err, "failed to create %q", planID)
	}

	return nil
}

func (c *Client) Bind(
	serviceID, planID string,
	bindParams, provisionParams map[string]interface{}) (map[string]interface{}, error) {

	params := make(map[string]interface{}, len(bindParams)+len(provisionParams))
	for k, v := range provisionParams {
		params[k] = v
	}
	for k, v := range bindParams {
		params[k] = v
	}

	service, err := c.coreClient.CoreV1().Services(c.namespace).Get(planID, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	secret, err := c.coreClient.CoreV1().Secrets(c.namespace).Get(planID+"-auth", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	for key, value := range secret.Data {
		data[key] = string(value)
	}

	// Apply additional provisioning logic for Service Catalog Enabled services
	//provider, ok := c.providers["mysqldb"]
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

func (c *Client) Deprovision(serviceID, planID string) error {
	//fmt.Print("getting provider for 'mysqldb'\n")
	glog.Infof("getting provider for %q", serviceID)
	//provider, ok := c.providers["mysqldb"]
	provider, ok := c.providers[serviceID]
	if !ok {
		return errors.Errorf("No %q provider found", serviceID)
	}

	fmt.Printf("deleting %q obj", planID)
	if err := provider.Delete(planID, c.namespace); err != nil {
		return errors.Wrapf(err, "failed to delete %q", planID)
	}

	return nil
}
