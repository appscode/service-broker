package db_broker

import (
	"encoding/json"
	"fmt"
	"reflect"

	meta_util "github.com/appscode/kutil/meta"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Provider interface {
	Bind(service corev1.Service, params map[string]interface{}, chartSecrets map[string]interface{}) (*Credentials, error)
	Create(provisionInfo ProvisionInfo, namespace string) error
	Delete(name, namespace string) error
	GetProvisionInfo(instanceID, namespace string) (*ProvisionInfo, error)
}

type ProvisionInfo struct {
	InstanceID string
	ServiceID  string
	PlanID     string
	Params     map[string]interface{}

	InstanceName string
}

func instanceFromObjectMeta(meta metav1.ObjectMeta) (*ProvisionInfo, error) {
	var provisionInfo ProvisionInfo
	err := json.Unmarshal([]byte(meta.Annotations[ProvisionInfoKey]), &provisionInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshall provision info for instance %q", meta.Labels[InstanceKey])
	}
	return &provisionInfo, nil
}

func (p *ProvisionInfo) Match(q *ProvisionInfo) bool {
	return p.InstanceID == q.InstanceID &&
		p.ServiceID == q.ServiceID &&
		p.PlanID == q.PlanID &&
		reflect.DeepEqual(p.Params, q.Params)
}

func (p ProvisionInfo) applyToMetadata(meta *metav1.ObjectMeta, namespace string) error {
	var (
		err error
	)

	if _, found := p.Params["metadata"]; found {
		if err := meta_util.Decode(p.Params["metadata"], meta); err != nil {
			return err
		}
	}

	// set instance id at labels
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[InstanceKey] = p.InstanceID

	// set provision info at annotations
	var provisionInfoJson []byte
	if provisionInfoJson, err = json.Marshal(p); err != nil {
		return errors.Wrapf(err, "could not marshall provisioning info %v", p)
	}

	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	meta.Annotations[ProvisionInfoKey] = string(provisionInfoJson)

	meta.Name = p.InstanceName
	meta.Namespace = namespace

	return nil
}

func (p ProvisionInfo) applyToSpec(spec interface{}) error {
	var (
		err error
	)

	if _, found := p.Params["spec"]; !found {
		return errors.New("spec is required for provisioning custom postgres")
	}

	if err = meta_util.Decode(p.Params["spec"], spec); err != nil {
		return err
	}

	return nil
}

type Credentials struct {
	Protocol string
	URI      string `json:"uri,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     int32  `json:"port,omitempty"`
	Database string `json:"database,omitempty"`
	RootCert string `json:"rootCert,omitempty"`
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
	j, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(j, &result)
	return result, err
}

func buildURI(c Credentials) string {
	var uri = fmt.Sprintf("%s://", c.Protocol)
	if c.Username != "" {
		uri = fmt.Sprintf("%s%s:%s@", uri, c.Username, c.Password)
	}
	uri = fmt.Sprintf("%s%s:%d", uri, c.Host, c.Port)
	if c.Database != "" {
		uri = fmt.Sprintf("%s/%s", uri, c.Database)
	}

	return uri
	//if c.Database == "" {
	//	return fmt.Sprintf("%s://%s:%s@%s:%d",
	//		c.Protocol, c.Username, c.Password, c.Host, c.Port)
	//}
	//
	//return fmt.Sprintf("%s://%s:%s@%s:%d/%s",
	//	c.Protocol, c.Username, c.Password, c.Host, c.Port, c.Database)
}

func buildHostFromService(service corev1.Service) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace)
}
