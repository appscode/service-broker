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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

type ElasticsearchProvider struct {
	extClient        cs.KubedbV1alpha1Interface
	storageClassName string
}

func NewElasticsearchProvider(config *rest.Config, storageClassName string) Provider {
	return &ElasticsearchProvider{
		extClient:        cs.NewForConfigOrDie(config),
		storageClassName: storageClassName,
	}
}

func demoElasticsearchSpec() api.ElasticsearchSpec {
	return api.ElasticsearchSpec{
		Version:           jsonTypes.StrYo(demoElasticSearchVersion),
		Replicas:          types.Int32P(1),
		EnableSSL:         true,
		StorageType:       api.StorageTypeEphemeral,
		TerminationPolicy: api.TerminationPolicyWipeOut,
	}
}

func demoElasticsearchClusterSpec() api.ElasticsearchSpec {
	return api.ElasticsearchSpec{
		Version:           jsonTypes.StrYo(demoElasticSearchVersion),
		EnableSSL:         true,
		StorageType:       api.StorageTypeEphemeral,
		TerminationPolicy: api.TerminationPolicyWipeOut,
		Topology: &api.ElasticsearchClusterTopology{
			Master: api.ElasticsearchNode{
				Prefix:   "master",
				Replicas: types.Int32P(1),
			},
			Data: api.ElasticsearchNode{
				Prefix:   "data",
				Replicas: types.Int32P(2),
			},
			Client: api.ElasticsearchNode{
				Prefix:   "client",
				Replicas: types.Int32P(1),
			},
		},
	}
}

func (p ElasticsearchProvider) Metadata() (string, string) {
	return "kubedb", "elasticsearch"
}

func (p ElasticsearchProvider) Create(provisionInfo ProvisionInfo) error {
	var es api.Elasticsearch

	// set metadata from provision info
	if err := provisionInfo.applyToMetadata(&es.ObjectMeta); err != nil {
		return err
	}

	// set postgres spec
	switch provisionInfo.PlanID {
	case planElasticSearchDemo:
		es.Spec = demoElasticsearchSpec()
	case planElasticSearchClusterDemo:
		es.Spec = demoElasticsearchClusterSpec()
	case planElasticSearch:
		if err := provisionInfo.applyToSpec(&es.Spec); err != nil {
			return err
		}
	}

	glog.Infof("Creating elasticsearch obj %q in namespace %q...", es.Name, es.Namespace)
	_, err := p.extClient.Elasticsearches(es.Namespace).Create(&es)

	return err
}

func (p ElasticsearchProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting elasticsearch obj %q from namespace %q...", name, namespace)

	es, err := p.extClient.Elasticsearches(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if es.Spec.TerminationPolicy != api.TerminationPolicyWipeOut {
		if err := patchElasticsearch(p.extClient, es, func(in *api.Elasticsearch) *api.Elasticsearch {
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}); err != nil {
			return err
		}
	}

	if err := p.extClient.Elasticsearches(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (p ElasticsearchProvider) Bind(
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

	rootCert, ok := data["root.pem"]
	if !ok {
		return nil, errors.Errorf(`root certificate not found in secret keys for %s %s/%s`, app.Spec.Type, app.Namespace, app.Name)
	}

	return &Credentials{
		Protocol: app.Spec.ClientConfig.Service.Scheme,
		Host:     host,
		Port:     port,
		URI:      uri,
		Username: username,
		Password: password,
		RootCert: rootCert,
	}, nil
}

func (p ElasticsearchProvider) GetProvisionInfo(instanceID string) (*ProvisionInfo, error) {
	elasticsearches, err := p.extClient.Elasticsearches(corev1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labels.Set{
			InstanceKey: instanceID,
		}.String(),
	})
	if err != nil || len(elasticsearches.Items) == 0 {
		return nil, err
	}

	if len(elasticsearches.Items) > 1 {
		var instances []string
		for _, elasticsearch := range elasticsearches.Items {
			instances = append(instances, fmt.Sprintf("%s/%s", elasticsearch.Namespace, elasticsearch.Name))
		}

		return nil, errors.Errorf("%d Elasticsearch clusters with instance id %s found: %s",
			len(elasticsearches.Items), instanceID, strings.Join(instances, ", "))
	}

	return provisionInfoFromObjectMeta(elasticsearches.Items[0].ObjectMeta)
}
