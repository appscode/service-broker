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

type ElasticsearchProvider struct {
	extClient cs.KubedbV1alpha1Interface
}

func NewElasticsearchProvider(config *rest.Config) Provider {
	return &ElasticsearchProvider{
		extClient: cs.NewForConfigOrDie(config),
	}
}

func NewElasticsearchObj(name, namespace string) *api.Elasticsearch {
	return &api.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: api.ElasticsearchSpec{
			Version:    jsonTypes.StrYo("5.6"),
			DoNotPause: true,
			Replicas:   types.Int32P(1),
			Storage: &corev1.PersistentVolumeClaimSpec{
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

func (p ElasticsearchProvider) Create(name, namespace string) error {
	glog.Infof("Creating elasticsearch obj %q in namespace %q...", name, namespace)
	es := NewElasticsearchObj(name, namespace)

	if _, err := p.extClient.Elasticsearches(es.Namespace).Create(es); err != nil {
		return err
	}

	return nil
	//return waitForElasticsearchBeReady(p.extClient, name, namespace)
}

func (p ElasticsearchProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting elasticsearch obj %q from namespace %q...", name, namespace)

	es, err := p.extClient.Elasticsearches(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if es.Spec.DoNotPause {
		if err := patchElasticsearch(p.extClient, es); err != nil {
			return err
		}
	}

	if err := p.extClient.Elasticsearches(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	glog.Infof("Deleting dormant database obj %q from namespace %q...", name, namespace)
	if err := patchDormantDatabase(p.extClient, name, namespace); err != nil {
		return err
	}

	return p.extClient.DormantDatabases(namespace).Delete(name, deleteInBackground())
}

func (p ElasticsearchProvider) Bind(
	service corev1.Service,
	params map[string]interface{},
	data map[string]interface{}) (*Credentials, error) {

	if len(service.Spec.Ports) == 0 {
		return nil, errors.Errorf("no ports found")
	}
	svcPort := service.Spec.Ports[0]

	host := buildHostFromService(service)

	database := ""
	if dbVal, ok := params["esDatabase"]; ok {
		database = dbVal.(string)
	}

	var user, password string

	user = "admin"
	rootPassword, ok := data["ADMIN_PASSWORD"]
	if !ok {
		return nil, errors.Errorf("ADMIN_PASSWORD not found in secret keys")
	}
	password = rootPassword.(string)

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
