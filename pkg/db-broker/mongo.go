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

type MongoDbProvider struct {
	extClient cs.KubedbV1alpha1Interface
}

func NewMongoDbProvider(config *rest.Config) Provider {
	return &MongoDbProvider{
		extClient: cs.NewForConfigOrDie(config),
	}
}

func NewMongoDbObj(name, namespace string) *api.MongoDB {
	return &api.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: api.MongoDBSpec{
			Version:    jsonTypes.StrYo("3.4"),
			DoNotPause: true,
			Replicas:   types.Int32P(1),
			Storage: corev1.PersistentVolumeClaimSpec{
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

func (p MongoDbProvider) Create(name, namespace string) error {
	glog.Infof("Creating mongodb obj %q in namespace %q...", name, namespace)
	mg := NewMongoDbObj(name, namespace)

	if _, err := p.extClient.MongoDBs(mg.Namespace).Create(mg); err != nil {
		return err
	}

	return nil
	//return waitForMongoDbBeReady(p.extClient, name, namespace)
}

func (p MongoDbProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting mongodb obj %q from namespace %q...", name, namespace)

	mg, err := p.extClient.MongoDBs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if mg.Spec.DoNotPause {
		if err := patchMongoDb(p.extClient, mg); err != nil {
			return err
		}
	}

	if err := p.extClient.MongoDBs(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	glog.Infof("Deleting dormant database obj %q from namespace %q...", name, namespace)
	if err := patchDormantDatabase(p.extClient, name, namespace); err != nil {
		return err
	}

	return p.extClient.DormantDatabases(namespace).Delete(name, deleteInBackground())
}

func (p MongoDbProvider) Bind(
	service corev1.Service,
	params map[string]interface{},
	data map[string]interface{}) (*Credentials, error) {

	if len(service.Spec.Ports) == 0 {
		return nil, errors.Errorf("no ports found")
	}
	svcPort := service.Spec.Ports[0]

	host := buildHostFromService(service)

	database := ""
	if dbVal, ok := params["mgDatabase"]; ok {
		database = dbVal.(string)
	}

	var user, password string
	userVal, ok := params["mgUser"]
	if ok {
		user = userVal.(string)

		passwordVal, ok := data["mgPassword"]
		if !ok {
			return nil, errors.Errorf("mongodb-password not found in secret keys")
		}
		password = passwordVal.(string)
	} else {
		user = "root"

		rootPassword, ok := data["password"]
		if !ok {
			return nil, errors.Errorf("mongodb-root-password not found in secret keys")
		}
		password = rootPassword.(string)
	}

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
