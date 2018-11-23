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

type MongoDbProvider struct {
	extClient        cs.KubedbV1alpha1Interface
	storageClassName string
}

func NewMongoDbProvider(config *rest.Config, storageClassName string) Provider {
	return &MongoDbProvider{
		extClient:        cs.NewForConfigOrDie(config),
		storageClassName: storageClassName,
	}
}

func NewMongoDB(name, namespace, storageClassName string) *api.MongoDB {
	return &api.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: api.MongoDBSpec{
			Version:     jsonTypes.StrYo("3.6-v1"),
			StorageType: api.StorageTypeDurable,
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

func NewMongoDBCluster(name, namespace, storageClassName string) *api.MongoDB {
	mg := NewMongoDB(name, namespace, storageClassName)
	mg.Spec.Replicas = types.Int32P(3)
	mg.Spec.ReplicaSet = &api.MongoDBReplicaSet{
		Name: "rs0",
	}

	return mg
}

func (p MongoDbProvider) Create(planID, name, namespace string) error {
	glog.Infof("Creating mongodb obj %q in namespace %q...", name, namespace)

	var mg *api.MongoDB

	switch planID {
	case "mongodb-3-6":
		mg = NewMongoDB(name, namespace, p.storageClassName)
	case "mongodb-cluster-3-6":
		mg = NewMongoDBCluster(name, namespace, p.storageClassName)
	}

	if _, err := p.extClient.MongoDBs(mg.Namespace).Create(mg); err != nil {
		return err
	}

	return nil
}

func (p MongoDbProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting mongodb obj %q from namespace %q...", name, namespace)

	mg, err := p.extClient.MongoDBs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if mg.Spec.TerminationPolicy != api.TerminationPolicyWipeOut {
		if err := patchMongoDb(p.extClient, mg); err != nil {
			return err
		}
	}

	if err := p.extClient.MongoDBs(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (p MongoDbProvider) Bind(
	service corev1.Service,
	params map[string]interface{},
	data map[string]interface{}) (*Credentials, error) {

	var (
		user, password   string
		connScheme, host string
		port             int32
	)

	connScheme = "mongodb"
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
	if dbVal, ok := params["mgDatabase"]; ok {
		database = dbVal.(string)
	}

	userVal, ok := params["mgUser"]
	if ok {
		user = userVal.(string)
	} else {
		mgUser, ok := data["user"]
		if !ok {
			return nil, errors.Errorf("user not found in secret keys")
		}
		user = mgUser.(string)
	}

	mgPassword, ok := data["password"]
	if !ok {
		return nil, errors.Errorf("password not found in secret keys")
	}
	password = mgPassword.(string)

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
