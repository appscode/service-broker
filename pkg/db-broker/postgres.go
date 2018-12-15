package db_broker

import (
	jsonTypes "github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/types"
	"github.com/golang/glog"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
)

type PostgreSQLProvider struct {
	extClient        cs.KubedbV1alpha1Interface
	storageClassName string
}

func NewPostgreSQLProvider(config *rest.Config, storageClassName string) Provider {
	return &PostgreSQLProvider{
		extClient:        cs.NewForConfigOrDie(config),
		storageClassName: storageClassName,
	}
}

func DemoPostgresSpec() api.PostgresSpec {
	return api.PostgresSpec{
		Version:           jsonTypes.StrYo("10.2-v1"),
		Replicas:          types.Int32P(1),
		StorageType:       api.StorageTypeEphemeral,
		TerminationPolicy: api.TerminationPolicyWipeOut,
	}
}

func DemoHAPostgresSpec() api.PostgresSpec {
	pgSpec := DemoPostgresSpec()
	pgSpec.Replicas = types.Int32P(3)

	return pgSpec
}

func (p PostgreSQLProvider) Create(provisionInfo ProvisionInfo, namespace string) error {
	glog.Infof("Creating postgres obj %q in namespace %q...", provisionInfo.InstanceName, namespace)

	var pg api.Postgres

	// set metadata from provision info
	if err := provisionInfo.applyToMetadata(&pg.ObjectMeta, namespace); err != nil {
		return err
	}

	// set postgres spec
	switch provisionInfo.PlanID {
	case "demo-postgresql":
		pg.Spec = DemoPostgresSpec()
	case "demo-ha-postgresql":
		pg.Spec = DemoHAPostgresSpec()
	case "postgresql":
		if err := provisionInfo.applyToSpec(&pg.Spec); err != nil {
			return err
		}
	}

	_, err := p.extClient.Postgreses(pg.Namespace).Create(&pg)

	return err
}

func (p PostgreSQLProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting postgres obj %q from namespace %q...", name, namespace)

	pgsql, err := p.extClient.Postgreses(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if pgsql.Spec.TerminationPolicy != api.TerminationPolicyWipeOut {
		if err := patchPostgreSQL(p.extClient, pgsql, func(in *api.Postgres) *api.Postgres {
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}); err != nil {
			return err
		}
	}

	if err := p.extClient.Postgreses(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (p PostgreSQLProvider) Bind(
	service corev1.Service,
	params map[string]interface{},
	data map[string]interface{}) (*Credentials, error) {

	var (
		user, password   string
		connScheme, host string
		port             int32
	)

	connScheme = "postgresql"
	if len(service.Spec.Ports) == 0 {
		return nil, errors.Errorf("no ports found")
	}
	for _, p := range service.Spec.Ports {
		if p.Name == "api" {
			port = p.Port
			break
		}
	}

	host = buildHostFromService(service)

	database := "postgres"
	if dbVal, ok := params["pgsqlDatabase"]; ok {
		database = dbVal.(string)
	}

	userVal, ok := params["pgsqlUser"]
	if ok {
		user = userVal.(string)
	} else {
		pgUser, ok := data["POSTGRES_USER"]
		if !ok {
			return nil, errors.Errorf("POSTGRES_USER not found in secret keys")
		}
		user = pgUser.(string)
	}

	pgPassword, ok := data["POSTGRES_PASSWORD"]
	if !ok {
		return nil, errors.Errorf("POSTGRES_PASSWORD not found in secret keys")
	}
	password = pgPassword.(string)

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

func (p PostgreSQLProvider) GetProvisionInfo(instanceID, namespace string) (*ProvisionInfo, error) {
	postgreses, err := p.extClient.Postgreses(corev1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: k8sLabels.Set{
			InstanceKey: instanceID,
		}.String(),
	})
	if err != nil {
		return nil, err
	}

	if len(postgreses.Items) > 0 {
		return provisionInfoFromObjectMeta(postgreses.Items[0].ObjectMeta)
	}

	return nil, nil
}
