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

type RedisProvider struct {
	extClient        cs.KubedbV1alpha1Interface
	storageClassName string
}

func NewRedisProvider(config *rest.Config, storageClassName string) Provider {
	return &RedisProvider{
		extClient:        cs.NewForConfigOrDie(config),
		storageClassName: storageClassName,
	}
}

func NewRedis(name, namespace, storageClassName string) *api.Redis {
	return &api.Redis{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: api.RedisSpec{
			Version: jsonTypes.StrYo("4.0-v1"),
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

func (p RedisProvider) Create(planID, name, namespace string) error {
	glog.Infof("Creating redis obj %q in namespace %q...", name, namespace)
	rd := NewRedis(name, namespace, p.storageClassName)

	if _, err := p.extClient.Redises(rd.Namespace).Create(rd); err != nil {
		return err
	}

	return nil
}

func (p RedisProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting redis obj %q from namespace %q...", name, namespace)

	rd, err := p.extClient.Redises(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if rd.Spec.TerminationPolicy != api.TerminationPolicyWipeOut {
		if err := patchRedis(p.extClient, rd); err != nil {
			return err
		}
	}

	if err := p.extClient.Redises(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (p RedisProvider) Bind(
	service corev1.Service,
	params map[string]interface{},
	data map[string]interface{}) (*Credentials, error) {

	var (
		user, password   string
		connScheme, host string
		port             int32
	)

	connScheme = "redis"
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
	if dbVal, ok := params["rdDatabase"]; ok {
		database = dbVal.(string)
	}

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
