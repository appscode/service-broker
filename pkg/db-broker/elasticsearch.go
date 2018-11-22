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

func NewElasticsearch(name, namespace, storageClassName string) *api.Elasticsearch {
	return &api.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: api.ElasticsearchSpec{
			Version:   jsonTypes.StrYo("6.3-v1"),
			Replicas:  types.Int32P(1),
			EnableSSL: true,
			Storage: &corev1.PersistentVolumeClaimSpec{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("1Gi"),
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

func NewElasticsearchCluster(name, namespace, storageClassName string) *api.Elasticsearch {
	return &api.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: api.ElasticsearchSpec{
			Version:           jsonTypes.StrYo("6.3-v1"),
			EnableSSL:         true,
			StorageType:       api.StorageTypeDurable,
			TerminationPolicy: api.TerminationPolicyWipeOut,
			ServiceTemplate: ofst.ServiceTemplateSpec{
				Spec: ofst.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			Topology: &api.ElasticsearchClusterTopology{
				Master: api.ElasticsearchNode{
					Prefix:   "master",
					Replicas: types.Int32P(1),
					Storage: &corev1.PersistentVolumeClaimSpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: types.StringP(storageClassName),
					},
				},
				Data: api.ElasticsearchNode{
					Prefix:   "data",
					Replicas: types.Int32P(2),
					Storage: &corev1.PersistentVolumeClaimSpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						StorageClassName: types.StringP(storageClassName),
					},
				},
				Client: api.ElasticsearchNode{
					Prefix:   "client",
					Replicas: types.Int32P(1),
					Storage: &corev1.PersistentVolumeClaimSpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("50Mi"),
							},
						},
						StorageClassName: types.StringP(storageClassName),
					},
				},
			},
		},
	}
}

func (p ElasticsearchProvider) Create(planID, name, namespace string) error {
	glog.Infof("Creating elasticsearch obj %q in namespace %q...", name, namespace)

	var es *api.Elasticsearch

	switch planID {
	case "elasticsearch-6-3":
		es = NewElasticsearch(name, namespace, p.storageClassName)
	case "elasticsearch-cluster-6-3":
		es = NewElasticsearchCluster(name, namespace, p.storageClassName)
	}

	if _, err := p.extClient.Elasticsearches(es.Namespace).Create(es); err != nil {
		return err
	}

	return nil
}

func (p ElasticsearchProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting elasticsearch obj %q from namespace %q...", name, namespace)

	es, err := p.extClient.Elasticsearches(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if es.Spec.TerminationPolicy != api.TerminationPolicyWipeOut {
		if err := patchElasticsearch(p.extClient, es); err != nil {
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

	// todo: connScheme should be set depending on es.spec.EnableSSL, once we implement parametes passing
	connScheme = "https"
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
	//host := service.Spec.ExternalIPs[0]

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

	cert, ok := data["root.pem"]
	if !ok {
		return nil, errors.Errorf("root certificate not found in secret keys")
	}
	rootCert = cert.(string)

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
