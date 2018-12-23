package kubedb

import (
	"encoding/json"
	"reflect"

	meta_util "github.com/appscode/kutil/meta"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
)

type Provider interface {
	Metadata() (catalog string, serviceName string)
	Bind(app *appcat.AppBinding, params map[string]interface{}, chartSecrets map[string]interface{}) (*Credentials, error)
	Create(provisionInfo ProvisionInfo) error
	Delete(name, namespace string) error
	GetProvisionInfo(instanceID string) (*ProvisionInfo, error)
}

type ProvisionInfo struct {
	InstanceID  string
	ServiceID   string
	PlanID      string
	Params      map[string]interface{}
	ExtraParams map[string]interface{}

	InstanceName string
	Namespace    string
}

func provisionInfoFromObjectMeta(meta metav1.ObjectMeta) (*ProvisionInfo, error) {
	var provisionInfo ProvisionInfo
	err := json.Unmarshal([]byte(meta.Annotations[ProvisionInfoKey]), &provisionInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal provision info for instance %q", meta.Labels[InstanceKey])
	}
	return &provisionInfo, nil
}

func (p *ProvisionInfo) Match(q *ProvisionInfo) bool {
	return p.InstanceID == q.InstanceID &&
		p.ServiceID == q.ServiceID &&
		p.PlanID == q.PlanID &&
		reflect.DeepEqual(p.Params, q.Params)
}

func (p ProvisionInfo) applyToMetadata(meta *metav1.ObjectMeta) error {
	if _, found := p.Params["metadata"]; found {
		if err := meta_util.Decode(p.Params["metadata"], meta); err != nil {
			return err
		}
	}

	meta.Name = p.InstanceName
	meta.Namespace = p.Namespace

	// set instance id at labels
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[InstanceKey] = p.InstanceID

	// set provision info at annotations
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	if provisionInfoJson, err := json.Marshal(p); err != nil {
		return errors.Wrapf(err, "could not marshal provisioning info %v", p)
	} else {
		meta.Annotations[ProvisionInfoKey] = string(provisionInfoJson)
	}

	return nil
}

func (p ProvisionInfo) applyToSpec(spec interface{}) error {
	if _, found := p.Params["spec"]; !found {
		return errors.New("spec is required for provisioning custom postgres")
	}
	return meta_util.Decode(p.Params["spec"], spec)
}

// ref: https://github.com/osbkit/minibroker/blob/d212fcb0013fe73eae914543525e36a1b1fc91cd/pkg/minibroker/provider.go#L14:6
type Credentials struct {
	Protocol string
	Host     string      `json:"host,omitempty"`
	Port     int32       `json:"port,omitempty"`
	URI      string      `json:"uri,omitempty"`
	Username interface{} `json:"username,omitempty"`
	Password interface{} `json:"password,omitempty"`
	RootCert interface{} `json:"rootCert,omitempty"`
}

// ToMap converts the credentials into the OSB API credentials response
// see https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#device-object
// {
//   "credentials": {
//     "uri": "mysql://mysqluser:pass@mysqlhost:3306/dbname",
//     "username": "mysqluser",
//     "password": "pass",
//     "host": "mysqlhost",
//     "port": 3306,
//     "database": "dbname"
//     }
// }
func (c Credentials) ToMap() (map[string]interface{}, error) {
	var result map[string]interface{}
	err := meta_util.Decode(&c, &result)
	return result, err
}
