package db_broker

import (
	jsonTypes "github.com/appscode/go/encoding/json/types"
	"github.com/appscode/go/types"
	"github.com/appscode/kutil"
	"github.com/golang/glog"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
)

type PostgreSQLProvider struct {
	extClient cs.KubedbV1alpha1Interface
	//mysqls map[string]*api.MySQL
}

func NewPostgreSQLProvider(kubeConfig string) Provider {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err)
	}
	return &MySQLProvider{
		extClient: cs.NewForConfigOrDie(config),
		//mysqls: make(map[string]*api.MySQL),
	}
}

func NewPostgresObj(name, namespace string) *api.Postgres {
	return &api.Postgres{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: api.PostgresSpec{
			Version:    jsonTypes.StrYo("9.6"),
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

func (p PostgreSQLProvider) Create(name, namespace string) error {
	glog.Infof("Create(%q, %q) error {}", name, namespace)
	pgObj := NewPostgresObj(name, namespace)
	//p.mysqls[name] = myObj

	_, err := p.extClient.Postgreses(pgObj.Namespace).Create(pgObj)
	if err != nil {
		return err
	}
	err = wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		pgsql, err := p.extClient.Postgreses(pgObj.Namespace).Get(pgObj.Name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return pgsql.Status.Phase == api.DatabasePhaseRunning, nil
	})

	glog.Infof("Create(%q, %q) error {} complete\n", name, namespace)
	return err
}

func (p PostgreSQLProvider) Delete(name, namespace string) error {
	glog.Infof("Delete(%q %q) error {}", name, namespace)
	//meta := p.mysqls[name].ObjectMeta
	if err := p.extClient.Postgreses(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	err := wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
		dormantDatabase, err := p.extClient.DormantDatabases(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		dormantDatabase, _, err = util.PatchDormantDatabase(p.extClient, dormantDatabase, func(in *api.DormantDatabase) *api.DormantDatabase {
			in.Spec.WipeOut = true
			return in
		})
		return true, nil
	})
	if err != nil {
		return err
	}

	glog.Infof("Delete(%q %q) error {} complete", name, namespace)
	return p.extClient.DormantDatabases(namespace).Delete(name, deleteInBackground())
	//if p.extClient.DormantDatabases(meta.Namespace).Delete(meta.Name, deleteInBackground()); err != nil {
	//	return err
	//}
	//err = wait.PollImmediate(kutil.RetryInterval, kutil.ReadinessTimeout, func() (bool, error) {
	//	dormantDatabase, err := p.extClient.DormantDatabases(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
	//	if err != nil {
	//		return false, nil
	//	}
	//	if err != nil {
	//		return false, nil
	//	}
	//	return len(podList.Items) == 0, nil
	//})
}

func (p PostgreSQLProvider) Bind(
	service corev1.Service,
	params map[string]interface{},
	data map[string]interface{}) (*Credentials, error) {

	if len(service.Spec.Ports) == 0 {
		return nil, errors.Errorf("no ports found")
	}
	svcPort := service.Spec.Ports[0]

	host := buildHostFromService(service)

	database := "postgress"
	if dbVal, ok := params["pgsqlDatabase"]; ok {
		database = dbVal.(string)
	}

	var user, password string
	userVal, ok := params["pgsqlUser"]
	if ok {
		user = userVal.(string)
	} else {
		user = "postgres"

		rootPassword, ok := data["POSTGRES_PASSWORD"]
		if !ok {
			return nil, errors.Errorf("pgsql-password not found in secret keys")
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
