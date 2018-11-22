package db_broker

import (
	jsonTypes "github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/types"
	"github.com/golang/glog"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ofst "kmodules.xyz/offshoot-api/api/v1"
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

func NewMySQL(name, namespace, storageClassName string) *api.MySQL {
	return &api.MySQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: api.MySQLSpec{
			Version: jsonTypes.StrYo("8.0-v1"),
			Storage: &corev1.PersistentVolumeClaimSpec{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("50Mi"),
					},
				},
				StorageClassName: types.StringP(storageClassName),
			},
			TerminationPolicy: api.TerminationPolicyWipeOut,
			ServiceTemplate: ofst.ServiceTemplateSpec{
				Spec: ofst.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
		},
	}
}

func (p MySQLProvider) Create(planID, name, namespace string) error {
	glog.Infof("Creating mysql obj %q in namespace %q...", name, namespace)
	my := NewMySQL(name, namespace, p.storageClassName)

	if _, err := p.extClient.MySQLs(my.Namespace).Create(my); err != nil {
		return err
	}

	return nil
}

func (p MySQLProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting mysql obj %q from namespace %q...", name, namespace)

	mysql, err := p.extClient.MySQLs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if mysql.Spec.TerminationPolicy != api.TerminationPolicyWipeOut {
		if err := patchMySQL(p.extClient, mysql); err != nil {
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
	//host := service.Spec.ExternalIPs[0]

	database := ""
	if dbVal, ok := params["mysqlDatabase"]; ok {
		database = dbVal.(string)
	}

	userVal, ok := params["mysqlUser"]
	if ok {
		user = userVal.(string)
	} else {
		mysqlUser, ok := data["user"]
		if !ok {
			return nil, errors.Errorf("mysql-user not found in secret keys")
		}
		user = mysqlUser.(string)
	}

	mysqlPassword, ok := data["password"]
	if !ok {
		return nil, errors.Errorf("mysql-password not found in secret keys")
	}
	password = mysqlPassword.(string)

	creds := Credentials{
		Protocol: connScheme,
		Port:     port,
		Host:     host,
		Username: user,
		Password: password,
		Database: database,
	}
	creds.URI = buildURI(creds)

	return &creds, nil
}
