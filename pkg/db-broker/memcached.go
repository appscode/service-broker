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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	ofst "kmodules.xyz/offshoot-api/api/v1"
)

type MemcachedProvider struct {
	extClient cs.KubedbV1alpha1Interface
}

func NewMemcachedProvider(config *rest.Config) Provider {
	return &MemcachedProvider{
		extClient: cs.NewForConfigOrDie(config),
	}
}

func NewMemcached(name, namespace string, labels, annotations map[string]string) *api.Memcached {
	return &api.Memcached{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: api.MemcachedSpec{
			Version:  jsonTypes.StrYo("1.5.4-v1"),
			Replicas: types.Int32P(3),
			PodTemplate: ofst.PodTemplateSpec{
				Spec: ofst.PodSpec{
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("250m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
					},
				},
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

func (p MemcachedProvider) Create(provisionInfo ProvisionInfo, namespace string) error {
	glog.Infof("Creating memcached obj %q in namespace %q...", provisionInfo.InstanceName, namespace)

	var (
		provisionInfoJson []byte
		err               error
	)

	if provisionInfoJson, err = json.Marshal(provisionInfo); err != nil {
		return errors.Wrapf(err, "could not marshall provisioning info %v", provisionInfo)
	}
	annotations := map[string]string{
		"provision-info": string(provisionInfoJson),
	}
	labels := map[string]string{
		InstanceKey: provisionInfo.InstanceID,
	}

	mc := NewMemcached(provisionInfo.InstanceName, namespace, labels, annotations)

	if _, err := p.extClient.Memcacheds(mc.Namespace).Create(mc); err != nil {
		return err
	}

	return nil
}

func (p MemcachedProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting memcached obj %q from namespace %q...", name, namespace)

	mc, err := p.extClient.Memcacheds(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if mc.Spec.TerminationPolicy != api.TerminationPolicyWipeOut {
		if err := patchMemcached(p.extClient, mc, func(in *api.Memcached) *api.Memcached {
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}); err != nil {
			return err
		}
	}

	if err := p.extClient.Memcacheds(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (p MemcachedProvider) Bind(
	service corev1.Service,
	params map[string]interface{},
	data map[string]interface{}) (*Credentials, error) {

	var (
		user, password   string
		connScheme, host string
		port             int32
	)

	connScheme = "memcache"
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
	if dbVal, ok := params["mcDatabase"]; ok {
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

func (p MemcachedProvider) GetProvisionInfo(instanceID, namespace string) (*ProvisionInfo, error) {
	memcacheds, err := p.extClient.Memcacheds(corev1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labels.Set{
			InstanceKey: instanceID,
		}.String(),
	})
	if err != nil {
		return nil, err
	}

	if len(memcacheds.Items) > 0 {
		return instanceFromObjectMeta(memcacheds.Items[0].ObjectMeta)
	}

	return nil, nil
}
