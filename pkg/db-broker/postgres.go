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
)

type PostgreSQLProvider struct {
	extClient cs.KubedbV1alpha1Interface
}

func NewPostgreSQLProvider(config *rest.Config) Provider {
	return &PostgreSQLProvider{
		extClient: cs.NewForConfigOrDie(config),
	}
}

func NewPostgresObj(name, namespace string) *api.Postgres {
	return &api.Postgres{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: api.PostgresSpec{
			Version: jsonTypes.StrYo("9.6"),
			//DoNotPause: true,
			Replicas: types.Int32P(1),
			Storage: corev1.PersistentVolumeClaimSpec{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
				StorageClassName: types.StringP("standard"),
			},
		},
	}
}

func (p PostgreSQLProvider) Create(name, namespace string) error {
	glog.Infof("Creating postgres obj %q in namespace %q...", name, namespace)
	pgObj := NewPostgresObj(name, namespace)

	if _, err := p.extClient.Postgreses(pgObj.Namespace).Create(pgObj); err != nil {
		return err
	}

	return waitForPostgreSQLBeReady(p.extClient, name, namespace)
}

func (p PostgreSQLProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting postgres obj %q from namespace %q...", name, namespace)

	pgsql, err := p.extClient.Postgreses(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if pgsql.Spec.DoNotPause {
		if err := patchPostgreSQL(p.extClient, pgsql); err != nil {
			return err
		}
	}

	if err := p.extClient.Postgreses(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	glog.Infof("Deleting dormant database obj %q from namespace %q...", name, namespace)
	if err := patchDormantDatabase(p.extClient, name, namespace); err != nil {
		return err
	}

	return p.extClient.DormantDatabases(namespace).Delete(name, deleteInBackground())
}

func (p PostgreSQLProvider) Bind(
	service corev1.Service,
	params map[string]interface{},
	data map[string]interface{}) (*Credentials, error) {

	if len(service.Spec.Ports) == 0 {
		return nil, errors.Errorf("no ports found")
	}
	svcPort := service.Spec.Ports[0]

	host := buildHostFromService(service)

	database := "postgress"
	if dbVal, ok := params["pgsqlDatabase"]; ok {
		database = dbVal.(string)
	}

	var user, password string
	userVal, ok := params["pgsqlUser"]
	if ok {
		user = userVal.(string)
	} else {
		user = "postgres"

		rootPassword, ok := data["POSTGRES_PASSWORD"]
		if !ok {
			return nil, errors.Errorf("pgsql-password not found in secret keys")
		}
		password = rootPassword.(string)
	}

	creds := Credentials{
		Protocol: svcPort.Name,
		Port:     svcPort.Port,
		Host:     host,
		Username: user,
		Password: password,
		Database: database,
	}
	creds.URI = buildURI(creds)

	return &creds, nil
}
