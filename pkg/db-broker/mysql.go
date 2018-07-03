package db_broker

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type MySQLProvider struct{}

func (p MySQLProvider) Bind(services []corev1.Service, params map[string]interface{}, chartSecrets map[string]interface{}) (*Credentials, error) {
	service := services[0]
	if len(service.Spec.Ports) == 0 {
		return nil, errors.Errorf("no ports found")
	}
	svcPort := service.Spec.Ports[0]

	host := buildHostFromService(service)

	database := ""
	dbVal, ok := params["mysqlDatabase"]
	if ok {
		database = dbVal.(string)
	}

	var user, password string
	userVal, ok := params["mysqlUser"]
	if ok {
		user = userVal.(string)

		passwordVal, ok := chartSecrets["mysql-password"]
		if !ok {
			return nil, errors.Errorf("mysql-password not found in secret keys")
		}
		password = passwordVal.(string)
	} else {
		user = "root"

		rootPassword, ok := chartSecrets["mysql-root-password"]
		if !ok {
			return nil, errors.Errorf("mysql-root-password not found in secret keys")
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