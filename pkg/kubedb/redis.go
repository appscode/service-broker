package kubedb

import (
	"fmt"
	"strings"

	jsonTypes "github.com/appscode/go/encoding/json/types"
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

type RedisProvider struct {
	extClient cs.KubedbV1alpha1Interface
}

func NewRedisProvider(config *rest.Config) Provider {
	return &RedisProvider{
		extClient: cs.NewForConfigOrDie(config),
	}
}

func demoRedisSpec() api.RedisSpec {
	return api.RedisSpec{
		Version:           jsonTypes.StrYo(demoRedisVersion),
		StorageType:       api.StorageTypeEphemeral,
		TerminationPolicy: api.TerminationPolicyWipeOut,
	}
}

func (p RedisProvider) Metadata() (string, string) {
	return "kubedb", "redis"
}

func (p RedisProvider) Create(provisionInfo ProvisionInfo) error {
	var rd api.Redis

	// set metadata from provision info
	if err := provisionInfo.applyToMetadata(&rd.ObjectMeta); err != nil {
		return err
	}

	// set postgres spec
	switch provisionInfo.PlanID {
	case planRedisDemo:
		rd.Spec = demoRedisSpec()
	case planRedis:
		if err := provisionInfo.applyToSpec(&rd.Spec); err != nil {
			return err
		}
	}

	glog.Infof("Creating redis obj %q in namespace %q...", rd.Name, rd.Namespace)
	_, err := p.extClient.Redises(rd.Namespace).Create(&rd)

	return err
}

func (p RedisProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting redis obj %q from namespace %q...", name, namespace)

	rd, err := p.extClient.Redises(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if rd.Spec.TerminationPolicy != api.TerminationPolicyWipeOut {
		if err := patchRedis(p.extClient, rd, func(in *api.Redis) *api.Redis {
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}); err != nil {
			return err
		}
	}

	if err := p.extClient.Redises(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (p RedisProvider) Bind(
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

func (p RedisProvider) GetProvisionInfo(instanceID string) (*ProvisionInfo, error) {
	redises, err := p.extClient.Redises(corev1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labels.Set{
			InstanceKey: instanceID,
		}.String(),
	})
	if err != nil || len(redises.Items) == 0 {
		return nil, err
	}

	if len(redises.Items) > 1 {
		var instances []string
		for _, redis := range redises.Items {
			instances = append(instances, fmt.Sprintf("%s/%s", redis.Namespace, redis.Name))
		}

		return nil, errors.Errorf("%d Redises with instance id %s found: %s",
			len(redises.Items), instanceID, strings.Join(instances, ", "))
	}
	return provisionInfoFromObjectMeta(redises.Items[0].ObjectMeta)
}
