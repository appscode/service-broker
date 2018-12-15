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

func DemoMongoDBSpec() api.MongoDBSpec {
	return api.MongoDBSpec{
		Version:           jsonTypes.StrYo("3.6-v1"),
		StorageType:       api.StorageTypeEphemeral,
		TerminationPolicy: api.TerminationPolicyWipeOut,
	}
}

func DemoMongoDBClusterSpec() api.MongoDBSpec {
	mgSpec := DemoMongoDBSpec()
	mgSpec.Replicas = types.Int32P(3)
	mgSpec.ReplicaSet = &api.MongoDBReplicaSet{
		Name: "rs0",
	}

	return mgSpec
}

func (p MongoDbProvider) Create(provisionInfo ProvisionInfo, namespace string) error {
	glog.Infof("Creating mongodb obj %q in namespace %q...", provisionInfo.InstanceName, namespace)

	var mg api.MongoDB

	// set metadata from provision info
	if err := provisionInfo.applyToMetadata(&mg.ObjectMeta, namespace); err != nil {
		return err
	}

	// set postgres spec
	switch provisionInfo.PlanID {
	case "demo-mongodb":
		mg.Spec = DemoMongoDBSpec()
	case "demo-mongodb-cluster":
		mg.Spec = DemoMongoDBClusterSpec()
	case "mongodb":
		if err := provisionInfo.applyToSpec(&mg.Spec); err != nil {
			return err
		}
	}

	_, err := p.extClient.MongoDBs(mg.Namespace).Create(&mg)

	return err
}

func (p MongoDbProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting mongodb obj %q from namespace %q...", name, namespace)

	mg, err := p.extClient.MongoDBs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if mg.Spec.TerminationPolicy != api.TerminationPolicyWipeOut {
		if err := patchMongoDb(p.extClient, mg, func(in *api.MongoDB) *api.MongoDB {
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}); err != nil {
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

	database := ""
	if dbVal, ok := params["mgDatabase"]; ok {
		database = dbVal.(string)
	}

	userVal, ok := params["mgUser"]
	if ok {
		user = userVal.(string)
	} else {
		mgUser, ok := data["username"]
		if !ok {
			return nil, errors.Errorf("username not found in secret keys")
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

func (p MongoDbProvider) GetProvisionInfo(instanceID, namespace string) (*ProvisionInfo, error) {
	mongodbs, err := p.extClient.MongoDBs(corev1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labels.Set{
			InstanceKey: instanceID,
		}.String(),
	})
	if err != nil {
		return nil, err
	}

	if len(mongodbs.Items) > 0 {
		return provisionInfoFromObjectMeta(mongodbs.Items[0].ObjectMeta)
	}

	return nil, nil
}
