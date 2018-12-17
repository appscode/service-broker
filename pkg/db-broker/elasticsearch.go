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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
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
		Version:           jsonTypes.StrYo("6.3-v1"),
		Replicas:          types.Int32P(1),
		EnableSSL:         true,
		StorageType:       api.StorageTypeEphemeral,
		TerminationPolicy: api.TerminationPolicyWipeOut,
	}
}

func demoElasticsearchClusterSpec() api.ElasticsearchSpec {
	return api.ElasticsearchSpec{
		Version:           jsonTypes.StrYo("6.3-v1"),
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

func (p ElasticsearchProvider) Create(provisionInfo ProvisionInfo, namespace string) error {
	glog.Infof("Creating elasticsearch obj %q in namespace %q...", provisionInfo.InstanceName, namespace)

	var es api.Elasticsearch

	// set metadata from provision info
	if err := provisionInfo.applyToMetadata(&es.ObjectMeta, namespace); err != nil {
		return err
	}

	// set postgres spec
	switch provisionInfo.PlanID {
	case "demo-elasticsearch":
		es.Spec = demoElasticsearchSpec()
	case "demo-elasticsearch-cluster":
		es.Spec = demoElasticsearchClusterSpec()
	case "elasticsearch":
		if err := provisionInfo.applyToSpec(&es.Spec); err != nil {
			return err
		}
	}

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
	service corev1.Service,
	params map[string]interface{},
	data map[string]interface{}) (*Credentials, error) {

	var (
		user, password   string
		connScheme, host string
		port             int32
		rootCert         string
	)

	connScheme = "http"
	if enableSSL, ok := params["enableSSL"]; ok && enableSSL.(bool) {
		connScheme = "https"
		cert, ok := data["root.pem"]
		if !ok {
			return nil, errors.Errorf("root certificate not found in secret keys")
		}
		rootCert = cert.(string)
	}

	if len(service.Spec.Ports) == 0 {
		return nil, errors.Errorf("no ports found")
	}
	for _, p := range service.Spec.Ports {
		if p.Name == api.ElasticsearchRestPortName {
			port = p.Port
			break
		}
	}

	host = buildHostFromService(service)

	database := ""
	if dbVal, ok := params["esDatabase"]; ok {
		database = dbVal.(string)
	}
	userVal, ok := params["esUser"]

	if ok {
		user = userVal.(string)
	} else {
		adminUser, ok := data["ADMIN_USERNAME"]
		if !ok {
			return nil, errors.Errorf("ADMIN_USERNAME not found in secret keys")
		}
		user = adminUser.(string)
	}

	adminPassword, ok := data["ADMIN_PASSWORD"]
	if !ok {
		return nil, errors.Errorf("ADMIN_PASSWORD not found in secret keys")
	}
	password = adminPassword.(string)

	creds := Credentials{
		Protocol: connScheme,
		Port:     port,
		Host:     host,
		Username: user,
		Password: password,
		Database: database,
		RootCert: rootCert,
	}
	creds.URI = buildURI(creds)

	return &creds, nil
}

func (p ElasticsearchProvider) GetProvisionInfo(instanceID, namespace string) (*ProvisionInfo, error) {
	elasticsearches, err := p.extClient.Elasticsearches(corev1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labels.Set{
			InstanceKey: instanceID,
		}.String(),
	})
	if err != nil {
		return nil, err
	}

	var provisionInfo *ProvisionInfo
	if len(elasticsearches.Items) > 1 {
		return nil, errors.New("number of instances with same instance id should not be more than one")
	} else if len(elasticsearches.Items) == 1 {
		if provisionInfo, err = provisionInfoFromObjectMeta(elasticsearches.Items[0].ObjectMeta); err != nil {
			return nil, err
		}
		if provisionInfo.ExtraParams == nil {
			provisionInfo.ExtraParams = make(map[string]interface{})
		}
		provisionInfo.ExtraParams["enableSSL"] = elasticsearches.Items[0].Spec.EnableSSL
	}

	return provisionInfo, nil
}
