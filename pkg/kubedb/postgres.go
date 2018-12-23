package kubedb

import (
	"fmt"
	"strings"

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
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

type PostgreSQLProvider struct {
	extClient        cs.KubedbV1alpha1Interface
	storageClassName string
}

func NewPostgreSQLProvider(config *rest.Config) Provider {
	return &PostgreSQLProvider{
		extClient: cs.NewForConfigOrDie(config),
	}
}

func demoPostgresSpec() api.PostgresSpec {
	return api.PostgresSpec{
		Version:           jsonTypes.StrYo(demoPostgresVersion),
		Replicas:          types.Int32P(1),
		StorageType:       api.StorageTypeEphemeral,
		TerminationPolicy: api.TerminationPolicyWipeOut,
	}
}

func demoHAPostgresSpec() api.PostgresSpec {
	pgSpec := demoPostgresSpec()
	pgSpec.Replicas = types.Int32P(3)

	return pgSpec
}

func (p PostgreSQLProvider) Metadata() (string, string) {
	return "kubedb", "postgresql"
}

func (p PostgreSQLProvider) Create(provisionInfo ProvisionInfo) error {
	var pg api.Postgres

	// set metadata from provision info
	if err := provisionInfo.applyToMetadata(&pg.ObjectMeta); err != nil {
		return err
	}

	// set postgres spec
	switch provisionInfo.PlanID {
	case planPostgresDemo:
		pg.Spec = demoPostgresSpec()
	case planPostgresHADemo:
		pg.Spec = demoHAPostgresSpec()
	case planPostgres:
		if err := provisionInfo.applyToSpec(&pg.Spec); err != nil {
			return err
		}
	}

	glog.Infof("Creating postgres obj %q in namespace %q...", pg.Name, pg.Namespace)
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
	app *appcat.AppBinding,
	params map[string]interface{},
	data map[string]interface{}) (*Credentials, error) {

	host, err := app.Hostname()
	if err != nil {
		return nil, errors.Wrapf(err, `failed to retrieve "host" from secret for %s %s/%s`, app.Spec.Type, app.Namespace, app.Name)
	}

	port, err := app.Port()
	if err != nil {
		return nil, errors.Wrapf(err, `failed to retrieve "port" from secret for %s %s/%s`, app.Spec.Type, app.Namespace, app.Name)
	}

	uri, err := app.URL()
	if err != nil {
		return nil, errors.Wrapf(err, `failed to retrieve "uri" from secret for %s %s/%s`, app.Spec.Type, app.Namespace, app.Name)
	}

	username, ok := data["username"]
	if !ok {
		return nil, errors.Errorf(`"username" not found in secret keys for %s %s/%s`, app.Spec.Type, app.Namespace, app.Name)
	}

	password, ok := data["password"]
	if !ok {
		return nil, errors.Errorf(`"password" not found in secret keys for %s %s/%s`, app.Spec.Type, app.Namespace, app.Name)
	}

	return &Credentials{
		Protocol: app.Spec.ClientConfig.Service.Scheme,
		Host:     host,
		Port:     port,
		URI:      uri,
		Username: username,
		Password: password,
	}, nil
}

func (p PostgreSQLProvider) GetProvisionInfo(instanceID string) (*ProvisionInfo, error) {
	postgreses, err := p.extClient.Postgreses(corev1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: k8sLabels.Set{
			InstanceKey: instanceID,
		}.String(),
	})
	if err != nil || len(postgreses.Items) == 0 {
		return nil, err
	}

	if len(postgreses.Items) > 1 {
		var instances []string
		for _, postgres := range postgreses.Items {
			instances = append(instances, fmt.Sprintf("%s/%s", postgres.Namespace, postgres.Name))
		}

		return nil, errors.Errorf("%d Postgreses with instance id %s found: %s",
			len(postgreses.Items), instanceID, strings.Join(instances, ", "))
	}
	return provisionInfoFromObjectMeta(postgreses.Items[0].ObjectMeta)
}
