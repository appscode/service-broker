package db_broker

import (
	"io/ioutil"
	"strings"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
	shell "github.com/codeskyblue/go-sh"
	"fmt"
)

const (
	InstanceLabel       = "dbbroker.instance"
)

type Client struct {
	namespace                 string
	coreClient                kubernetes.Interface
	providers                 map[string]Provider
}

func NewClient(KubeConfig string) *Client {
	ns := loadNamespace(KubeConfig)
	fmt.Println("broker namespace =", ns)
	return &Client{
		coreClient:                loadInClusterClient(KubeConfig),
		namespace:       ns,
		providers: map[string]Provider{
			"mysqldb":      NewMySQLProvider(KubeConfig),
			//"mariadb":    MariadbProvider{},
			//"postgresql": PostgresProvider{},
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

func (c *Client) Provision(instanceID, serviceID, planID, namespace string, provisionParams map[string]interface{}) error {
	sh := shell.NewSession()
	args := []interface{}{fmt.Sprintf("--namespace=%s", c.namespace)}
	SetupServer := filepath.Join("hack", "dev", "kubedb.sh")
	cmd := sh.Command(SetupServer, args...)
	err := cmd.Run()
	fmt.Print("'kubedb.sh' scripts run finish\n")
	if err != nil {
		return errors.Wrap(err, "failed to run 'kubedb' operator")
	}

	fmt.Print("getting provider for 'mysqldb'\n")
	provider, ok := c.providers["mysqldb"]
	if !ok {
		return errors.New("No 'mysqldb' provider found")
	}

	fmt.Printf("creating mysql obj %q\n", instanceID)
	if err :=  provider.Create(instanceID, c.namespace); err != nil {
		return err
	}

	return nil
}

func (c *Client) Bind(
	instanceID, serviceID string,
	bindParams, provisionParams map[string]interface{}) (map[string]interface{}, error) {

	params := make(map[string]interface{}, len(bindParams)+len(provisionParams))
	for k, v := range provisionParams {
		params[k] = v
	}
	for k, v := range bindParams {
		params[k] = v
	}

	service, err := c.coreClient.CoreV1().Services(c.namespace).Get(instanceID, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	secret, err := c.coreClient.CoreV1().Secrets(c.namespace).Get(instanceID+"-auth", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	for key, value := range secret.Data {
		data[key] = string(value)
	}

	// Apply additional provisioning logic for Service Catalog Enabled services
	//provider, ok := c.providers[serviceID]
	provider, ok := c.providers["mysqldb"]
	if ok {
		creds, err := provider.Bind(*service, params, data)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to bind instance %q", instanceID)
		}
		for k, v := range creds.ToMap() {
			data[k] = v
		}
	}

	return data, nil
}

func (c *Client) Deprovision(instanceID string) error {
	fmt.Print("getting provider for 'mysqldb'\n")
	provider, ok := c.providers["mysqldb"]
	if !ok {
		return errors.New("No 'mysqldb' provider found")
	}

	fmt.Printf("deleting mysql obj %q\n", instanceID)
	if err :=  provider.Delete(instanceID); err != nil {
		return errors.Wrapf(err, "failed to delete mysql instance %q", instanceID)
	}

	fmt.Print("'kubedb.sh' scripts run finish\n")
	sh := shell.NewSession()
	args := []interface{}{fmt.Sprintf("--namespace=%s", c.namespace), "--uninstall", "--purge"}
	SetupServer := filepath.Join("hack", "dev", "kubedb.sh")
	cmd := sh.Command(SetupServer, args...)
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "failed to delete 'kubedb' operator")
	}

	return nil
}