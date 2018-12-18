package db_broker

import (
	"fmt"
	"strings"

	jsonTypes "github.com/appscode/go/encoding/json/types"
	"github.com/golang/glog"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
)

type MySQLProvider struct {
	extClient        cs.KubedbV1alpha1Interface
	storageClassName string
}

func NewMySQLProvider(config *rest.Config, storageClassName string) Provider {
	return &MySQLProvider{
		extClient:        cs.NewForConfigOrDie(config),
		storageClassName: storageClassName,
	}
}

func demoMySQLSpec() api.MySQLSpec {
	return api.MySQLSpec{
		Version:           jsonTypes.StrYo(demoMySQLVersion),
		StorageType:       api.StorageTypeEphemeral,
		TerminationPolicy: api.TerminationPolicyWipeOut,
	}
}

func (p MySQLProvider) Create(provisionInfo ProvisionInfo) error {
	var my api.MySQL

	// set metadata from provision info
	if err := provisionInfo.applyToMetadata(&my.ObjectMeta); err != nil {
		return err
	}

	// set postgres spec
	switch provisionInfo.PlanID {
	case planMySQLDemo:
		my.Spec = demoMySQLSpec()
	case planMySQL:
		if err := provisionInfo.applyToSpec(&my.Spec); err != nil {
			return err
		}
	}

	glog.Infof("Creating mysql obj %q in namespace %q...", my.Name, my.Namespace)
	_, err := p.extClient.MySQLs(my.Namespace).Create(&my)

	return err
}

func (p MySQLProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting mysql obj %q from namespace %q...", name, namespace)

	my, err := p.extClient.MySQLs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if my.Spec.TerminationPolicy != api.TerminationPolicyWipeOut {
		if err := patchMySQL(p.extClient, my, func(in *api.MySQL) *api.MySQL {
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}); err != nil {
			return err
		}
	}

	if err := p.extClient.MySQLs(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (p MySQLProvider) Bind(
	service corev1.Service,
	params map[string]interface{},
	data map[string]interface{}) (*Credentials, error) {

	var (
		user, password   string
		connScheme, host string
		port             int32
	)

	connScheme = "mysql"
	if len(service.Spec.Ports) == 0 {
		return nil, errors.Errorf("no ports found")
	}
	for _, p := range service.Spec.Ports {
		if p.Name == "db" {
			port = p.Port
			break
		}
	}

	host = buildHostFromService(service)

	mysqlUser, ok := data["username"]
	if !ok {
		return nil, errors.Errorf("username not found in secret keys")
	}
	user = mysqlUser.(string)

	mysqlPassword, ok := data["password"]
	if !ok {
		return nil, errors.Errorf("password not found in secret keys")
	}
	password = mysqlPassword.(string)

	creds := Credentials{
		Protocol: connScheme,
		Port:     port,
		Host:     host,
		Username: user,
		Password: password,
	}
	creds.URI = buildURI(creds)

	return &creds, nil
}

func (p MySQLProvider) GetProvisionInfo(instanceID, namespace string) (*ProvisionInfo, error) {
	mysqls, err := p.extClient.MySQLs(corev1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labels.Set{
			InstanceKey: instanceID,
		}.String(),
	})
	if err != nil {
		return nil, err
	}

	if len(mysqls.Items) > 1 {
		var instances []string
		for _, mysql := range mysqls.Items {
			instances = append(instances, fmt.Sprintf("%s/%s", mysql.Namespace, mysql.Namespace))
		}

		return nil, errors.Errorf("%d MySQLs with instance id %s found: %s",
			len(mysqls.Items), instanceID, strings.Join(instances, ", "))
	} else if len(mysqls.Items) == 1 {
		return provisionInfoFromObjectMeta(mysqls.Items[0].ObjectMeta)
	}

	return nil, nil
}
