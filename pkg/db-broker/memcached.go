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
	"k8s.io/client-go/rest"
)

type MemcachedProvider struct {
	extClient cs.KubedbV1alpha1Interface
}

func NewMemcachedProvider(config *rest.Config) Provider {
	return &MemcachedProvider{
		extClient: cs.NewForConfigOrDie(config),
	}
}

func NewMemcachedObj(name, namespace string) *api.Memcached {
	return &api.Memcached{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: api.MemcachedSpec{
			Version:    jsonTypes.StrYo("1.5.4"),
			DoNotPause: true,
			Replicas:   types.Int32P(1),
		},
	}
}

func (p MemcachedProvider) Create(name, namespace string) error {
	glog.Infof("Creating memcached obj %q in namespace %q...", name, namespace)
	mc := NewMemcachedObj(name, namespace)

	if _, err := p.extClient.Memcacheds(mc.Namespace).Create(mc); err != nil {
		return err
	}

	return nil
	//return waitForMemcachedBeReady(p.extClient, name, namespace)
}

func (p MemcachedProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting memcached obj %q from namespace %q...", name, namespace)

	mc, err := p.extClient.Memcacheds(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if mc.Spec.DoNotPause {
		if err := patchMemcached(p.extClient, mc); err != nil {
			return err
		}
	}

	if err := p.extClient.Memcacheds(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	glog.Infof("Deleting dormant database obj %q from namespace %q...", name, namespace)
	if err := patchDormantDatabase(p.extClient, name, namespace); err != nil {
		return err
	}

	return p.extClient.DormantDatabases(namespace).Delete(name, deleteInBackground())
}

func (p MemcachedProvider) Bind(
	service corev1.Service,
	params map[string]interface{},
	data map[string]interface{}) (*Credentials, error) {

	if len(service.Spec.Ports) == 0 {
		return nil, errors.Errorf("no ports found")
	}
	svcPort := service.Spec.Ports[0]

	host := buildHostFromService(service)

	database := ""
	if dbVal, ok := params["mcDatabase"]; ok {
		database = dbVal.(string)
	}

	//var user, password string
	//userVal, ok := params["mcUser"]
	//if ok {
	//	user = userVal.(string)
	//
	//	passwordVal, ok := data["mcPassword"]
	//	if !ok {
	//		return nil, errors.Errorf("memcached-password not found in secret keys")
	//	}
	//	password = passwordVal.(string)
	//} else {
	//	user = "root"
	//
	//	rootPassword, ok := data["password"]
	//	if !ok {
	//		return nil, errors.Errorf("memcached-root-password not found in secret keys")
	//	}
	//	password = rootPassword.(string)
	//}

	creds := Credentials{
		Protocol: svcPort.Name,
		Port:     svcPort.Port,
		Host:     host,
		//Username: user,
		//Password: password,
		Database: database,
	}
	creds.URI = buildURI(creds)

	return &creds, nil
}
