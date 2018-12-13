package db_broker

import (
	"encoding/json"

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

func NewPostgres(name, namespace string, labels, annotations map[string]string, spec api.PostgresSpec) *api.Postgres {
	return &api.Postgres{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: spec,
		//Spec: api.PostgresSpec{
		//	Version:  jsonTypes.StrYo("10.2-v1"),
		//	Replicas: types.Int32P(1),
		//	Storage: &corev1.PersistentVolumeClaimSpec{
		//		Resources: corev1.ResourceRequirements{
		//			Requests: corev1.ResourceList{
		//				corev1.ResourceStorage: resource.MustParse("50Mi"),
		//			},
		//		},
		//		StorageClassName: types.StringP(storageClassName),
		//	},
		//	TerminationPolicy: api.TerminationPolicyWipeOut,
		//},
	}
}

func (p PostgreSQLProvider) Create(provisionInfo ProvisionInfo, namespace string) error {
	glog.Infof("Creating postgres obj %q in namespace %q...", provisionInfo.InstanceName, namespace)

	var (
		pgSpec            api.PostgresSpec
		pgSpecJson        []byte
		provisionInfoJson []byte
		err               error
	)

	if pgSpecJson, err = json.Marshal(provisionInfo.Params["pgSpec"]); err != nil {
		return errors.Wrapf(err, "could not marshall value of pgSpec in provisioning params %v", provisionInfo.Params["pgSpec"])
	}
	if err = json.Unmarshal(pgSpecJson, &pgSpec); err != nil {
		return errors.Errorf("could not unmarshal pgSpec in provisioning params %v", provisionInfo.Params["pgSpec"])
	}

	if provisionInfoJson, err = json.Marshal(provisionInfo); err != nil {
		return errors.Wrapf(err, "could not marshall provisioning info %v", provisionInfo)
	}

	annotations := map[string]string{
		"provision-info": string(provisionInfoJson),
	}
	labels := map[string]string{
		InstanceKey: provisionInfo.InstanceID,
	}

	completePosgresSpec(&pgSpec, provisionInfo.PlanID)
	pg := NewPostgres(provisionInfo.InstanceName, namespace, labels, annotations, pgSpec)
	if _, err := p.extClient.Postgreses(pg.Namespace).Create(pg); err != nil {
		return err
	}

	return nil
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
	//host := service.Spec.ExternalIPs[0]

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
		return instanceFromObjectMeta(postgreses.Items[0].ObjectMeta)
	}

	return nil, nil
}

func completePosgresSpec(pgSpec *api.PostgresSpec, planID string) {
	if pgSpec.Version == "" {
		pgSpec.Version = jsonTypes.StrYo("10.2-v1")
	}

	if pgSpec.Replicas == nil {
		if planID == "ha-postgresql-10-2" {
			pgSpec.Replicas = types.Int32P(3)
		} else {
			pgSpec.Replicas = types.Int32P(1)
		}
	}

	if pgSpec.Storage == nil {
		pgSpec.StorageType = api.StorageTypeEphemeral
	} else {
		pgSpec.StorageType = api.StorageTypeDurable
	}

	if pgSpec.TerminationPolicy == "" {
		pgSpec.TerminationPolicy = api.TerminationPolicyWipeOut
	}
}
