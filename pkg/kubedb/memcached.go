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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
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

func demoMemcachedSpec() api.MemcachedSpec {
	return api.MemcachedSpec{
		Version:  jsonTypes.StrYo(demoMemcachedVersion),
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
	}
}

func (p MemcachedProvider) Metadata() (string, string) {
	return "kubedb", "memcached"
}

func (p MemcachedProvider) Create(provisionInfo ProvisionInfo) error {
	var mc api.Memcached

	// set metadata from provision info
	if err := provisionInfo.applyToMetadata(&mc.ObjectMeta); err != nil {
		return err
	}

	// set postgres spec
	switch provisionInfo.PlanID {
	case PlanMemcachedDemo:
		mc.Spec = demoMemcachedSpec()
	case PlanMemcached:
		if err := provisionInfo.applyToSpec(&mc.Spec); err != nil {
			return err
		}
	}

	glog.Infof("Creating memcached obj %q in namespace %q...", mc.Name, mc.Namespace)
	_, err := p.extClient.Memcacheds(mc.Namespace).Create(&mc)

	return err
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

	return &Credentials{
		Protocol: app.Spec.ClientConfig.Service.Scheme,
		Host:     host,
		Port:     port,
		URI:      uri,
	}, nil
}

func (p MemcachedProvider) GetProvisionInfo(instanceID string) (*ProvisionInfo, error) {
	memcacheds, err := p.extClient.Memcacheds(corev1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labels.Set{
			InstanceKey: instanceID,
		}.String(),
	})
	if err != nil || len(memcacheds.Items) == 0 {
		return nil, err
	}

	if len(memcacheds.Items) > 1 {
		var instances []string
		for _, memcached := range memcacheds.Items {
			instances = append(instances, fmt.Sprintf("%s/%s", memcached.Namespace, memcached.Name))
		}

		return nil, errors.Errorf("%d Memcacheds with instance id %s found: %s",
			len(memcacheds.Items), instanceID, strings.Join(instances, ", "))
	}
	return provisionInfoFromObjectMeta(memcacheds.Items[0].ObjectMeta)
}
