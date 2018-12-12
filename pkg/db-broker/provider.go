package db_broker

import (
	"encoding/json"
	"fmt"
	"reflect"

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
func (c Credentials) ToMap() map[string]interface{} {
	var result map[string]interface{}
	j, _ := json.Marshal(c)
	json.Unmarshal(j, &result)
	return result
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
