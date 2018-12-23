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

type MySQLProvider struct {
	extClient cs.KubedbV1alpha1Interface
}

func NewMySQLProvider(config *rest.Config) Provider {
	return &MySQLProvider{
		extClient: cs.NewForConfigOrDie(config),
	}
}

func demoMySQLSpec() api.MySQLSpec {
	return api.MySQLSpec{
		Version:           jsonTypes.StrYo(demoMySQLVersion),
		StorageType:       api.StorageTypeEphemeral,
		TerminationPolicy: api.TerminationPolicyWipeOut,
	}
}

func (p MySQLProvider) Metadata() (string, string) {
	return "kubedb", "mysql"
}

func (p MySQLProvider) Create(provisionInfo ProvisionInfo) error {
	var my api.MySQL

	// set metadata from provision info
	if err := provisionInfo.applyToMetadata(&my.ObjectMeta); err != nil {
		return err
	}

	// set postgres spec
	switch provisionInfo.PlanID {
	case planMySQLDemo:
		my.Spec = demoMySQLSpec()
	case planMySQL:
		if err := provisionInfo.applyToSpec(&my.Spec); err != nil {
			return err
		}
	}

	glog.Infof("Creating mysql obj %q in namespace %q...", my.Name, my.Namespace)
	_, err := p.extClient.MySQLs(my.Namespace).Create(&my)

	return err
}

func (p MySQLProvider) Delete(name, namespace string) error {
	glog.Infof("Deleting mysql obj %q from namespace %q...", name, namespace)

	my, err := p.extClient.MySQLs(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if my.Spec.TerminationPolicy != api.TerminationPolicyWipeOut {
		if err := patchMySQL(p.extClient, my, func(in *api.MySQL) *api.MySQL {
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}); err != nil {
			return err
		}
	}

	if err := p.extClient.MySQLs(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (p MySQLProvider) Bind(
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

	username, ok := data["username"]
	if !ok {
		return nil, errors.Errorf(`"username" not found in secret keys for %s %s/%s`, app.Spec.Type, app.Namespace, app.Name)
	}

	password, ok := data["password"]
	if !ok {
		return nil, errors.Errorf(`"password" not found in secret keys for %s %s/%s`, app.Spec.Type, app.Namespace, app.Name)
	}

	return &Credentials{
		Protocol: app.Spec.ClientConfig.Service.Scheme,
		Host:     host,
		Port:     port,
		URI:      uri,
		Username: username,
		Password: password,
	}, nil
}

func (p MySQLProvider) GetProvisionInfo(instanceID string) (*ProvisionInfo, error) {
	mysqls, err := p.extClient.MySQLs(corev1.NamespaceAll).List(metav1.ListOptions{
		LabelSelector: labels.Set{
			InstanceKey: instanceID,
		}.String(),
	})
	if err != nil || len(mysqls.Items) == 0 {
		return nil, err
	}

	if len(mysqls.Items) > 1 {
		var instances []string
		for _, mysql := range mysqls.Items {
			instances = append(instances, fmt.Sprintf("%s/%s", mysql.Namespace, mysql.Name))
		}

		return nil, errors.Errorf("%d MySQLs with instance id %s found: %s",
			len(mysqls.Items), instanceID, strings.Join(instances, ", "))
	}
	return provisionInfoFromObjectMeta(mysqls.Items[0].ObjectMeta)
}
