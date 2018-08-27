package broker

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/appscode/go/crypto/rand"
	"github.com/golang/glog"
	"github.com/kubedb/service-broker/pkg/db-broker"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
)

// NewBroker is a hook that is called with the Options the program is run
// with. NewBroker is the place where you will initialize your
// Broker Logic the parameters passed in.
func NewBroker(o Options) (*Broker, error) {
	brClient := db_broker.NewClient(o.KubeConfig, o.StorageClass)
	// For example, if your Broker Logic requires a parameter from the command
	// line, you would unpack it from the Options and set it on the
	// Broker Logic here.
	return &Broker{
		Client:    brClient,
		async:     o.Async,
		instances: make(map[string]*exampleInstance, 10),
	}, nil
}

// Broker provides an implementation of broker.Interface
type Broker struct {
	Client *db_broker.Client

	// Indiciates if the broker should handle the requests asynchronously.
	async bool
	// Synchronize go routines.
	sync.RWMutex
	// Add fields here! These fields are provided purely as an example
	instances map[string]*exampleInstance
}

var _ broker.Interface = &Broker{}

func boolPtr(b bool) *bool {
	return &b
}

func (b *Broker) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	// Your catalog broker logic goes here

	response := &broker.CatalogResponse{}
	osbResponse := &osb.CatalogResponse{
		Services: []osb.Service{
			{
				Name:          "mysql",
				ID:            "mysql", //"4f6e6cf6-ffdd-425f-a2c7-3c9258ad246a",
				Description:   "The example service from the MySQL database!",
				Bindable:      true,
				PlanUpdatable: boolPtr(true),
				Metadata: map[string]interface{}{
					"displayName": "Example MySQL DB service",
					"imageUrl":    "http://www.cgtechworld.in/images/Training/technologies/mysql.png",
				},
				Plans: []osb.Plan{
					{
						Name: "default",
						//ID:          rand.WithUniqSuffix("mysql"), //"86064792-7ea2-467b-af93-ac9694d96d5b",
						ID:          "mysql-5-7",
						Description: "The default plan for the 'mysql' service",
						Free:        boolPtr(true),

						// if Free: true, then use Metadata as follows

						//Metadata: map[string]interface{}{
						//	"bullets":[]string{
						//		"20 GB of messages",
						//		"20 connections",
						//	},
						//	"costs":[]map[string]interface{}{
						//		map[string]interface{}{
						//			"amount": map[string]interface{}{
						//				"usd": 99.0,
						//			},
						//			"unit": "MONTHLY",
						//		},
						//		map[string]interface{}{
						//			"amount": map[string]interface{}{
						//				"usd": 0.99,
						//			},
						//			"unit": "1GB of messages over 20GB",
						//		},
						//	},
						//	"displayName":"MySQL Default",
						//},

						//Schemas: &osb.Schemas{
						//	ServiceInstance: &osb.ServiceInstanceSchema{
						//		Create: &osb.InputParametersSchema{
						//			Parameters: map[string]interface{}{
						//				"type": "object",
						//				"properties": map[string]interface{}{
						//					"color": map[string]interface{}{
						//						"type":    "string",
						//						"default": "Clear",
						//						"enum": []string{
						//							"Clear",
						//							"Beige",
						//							"Grey",
						//						},
						//					},
						//				},
						//			},
						//		},
						//	},
						//},
					},
				},
			},
			{
				Name:          "postgresql",
				ID:            "postgresql", //"3948rfjp-9eta-mcvi-s98q-35bth98345ho",
				Description:   "The example service from the PostgreSQL database!",
				Bindable:      true,
				PlanUpdatable: boolPtr(true),
				Metadata: map[string]interface{}{
					"displayName": "Example PostgreSQL DB service",
					"imageUrl":    "https://wiki.postgresql.org/images/3/30/PostgreSQL_logo.3colors.120x120.png",
				},
				Plans: []osb.Plan{
					{
						Name: "default",
						//ID:          rand.WithUniqSuffix("postgresql"), //"30495hkf-vnl0-93ru-yugh-d09345vhjocd",
						ID:          "postgresql-9-6", //"30495hkf-vnl0-93ru-yugh-d09345vhjocd",
						Description: "The default plan for the 'postgresql' service",
						Free:        boolPtr(true),
					},
				},
			},
			{
				Name:          "elasticsearch",
				ID:            "elasticsearch", //"3948rfjp-9eta-mcvi-s98q-35bth98345ho",
				Description:   "The example service from the ElasticSearch database!",
				Bindable:      true,
				PlanUpdatable: boolPtr(true),
				Metadata: map[string]interface{}{
					"displayName": "Example ElasticSearch DB service",
					"imageUrl":    "https://d22e4d61ky6061.cloudfront.net/sites/default/files/Elasticsearch_1.png",
				},
				Plans: []osb.Plan{
					{
						Name: "default",
						//ID:          rand.WithUniqSuffix("postgresql"), //"30495hkf-vnl0-93ru-yugh-d09345vhjocd",
						ID:          "elasticsearch-5-6", //"jkwe487h-fiw4-q987-hdsr-lzsiuhqw486b",
						Description: "The default plan for the 'elasticsearch' service",
						Free:        boolPtr(true),
					},
				},
			},
			{
				Name:          "mongodb",
				ID:            "mongodb",
				Description:   "The example service from the MongoDB database!",
				Bindable:      true,
				PlanUpdatable: boolPtr(true),
				Metadata: map[string]interface{}{
					"displayName": "Example Mongo DB service",
					"imageUrl":    "https://www.vectorlogo.zone/logos/mongodb/mongodb-card.png",
				},
				Plans: []osb.Plan{
					{
						Name:        "default",
						ID:          "mongodb-3-4",
						Description: "The default plan for the 'mongodb' service",
						Free:        boolPtr(true),
					},
				},
			},
			{
				Name:          "redis",
				ID:            "redis",
				Description:   "The example service from the Redis database!",
				Bindable:      true,
				PlanUpdatable: boolPtr(true),
				Metadata: map[string]interface{}{
					"displayName": "Example Redis DB service",
					"imageUrl":    "https://redis.io/images/redis-white.png",
				},
				Plans: []osb.Plan{
					{
						Name:        "default",
						ID:          "redis-4-0",
						Description: "The default plan for the 'redis' service",
						Free:        boolPtr(true),
					},
				},
			},
			{
				Name:          "memcached",
				ID:            "memcached",
				Description:   "The example service from the Memcache database!",
				Bindable:      true,
				PlanUpdatable: boolPtr(true),
				Metadata: map[string]interface{}{
					"displayName": "Example Memcache DB service",
					"imageUrl":    "https://www.veebimajutus.ee/media/uploads/5580065ce3dc0036053bdf2c.png",
				},
				Plans: []osb.Plan{
					{
						Name:        "default",
						ID:          "memcached-1-5-4",
						Description: "The default plan for the 'memcached' service",
						Free:        boolPtr(true),
					},
				},
			},
		},
	}

	glog.Infof("catalog response: %#+v", osbResponse)

	response.CatalogResponse = *osbResponse

	return response, nil
}

func (b *Broker) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	// Your provision logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	response := broker.ProvisionResponse{}
	exampleInstance := &exampleInstance{
		ID:        request.InstanceID,
		ServiceID: request.ServiceID,
		PlanID:    request.PlanID,
		Params:    request.Parameters,
		DbObjName: rand.WithUniqSuffix(request.PlanID),
	}

	// Check to see if this is the same instance
	i := b.instances[request.InstanceID]
	if i != nil {
		if i.Match(exampleInstance) {
			response.Exists = true
			glog.Infof("Instance %s is already exists", request.InstanceID)
			return &response, nil
		} else {
			// Instance ID in use, this is a conflict.
			description := "InstanceID in use"
			glog.Infof("The InstanceID %q is in use", request.InstanceID)
			return nil, osb.HTTPStatusCodeError{
				StatusCode:  http.StatusConflict,
				Description: &description,
			}
		}
	}

	glog.Infof("Provissioning instance %q for %q/%q...", request.InstanceID, request.ServiceID, request.PlanID)
	namespace := request.Context["namespace"].(string)
	err := b.Client.Provision(request.ServiceID, exampleInstance.DbObjName, namespace, request.Parameters)
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}
	glog.Infoln("Provisioning complete")

	b.instances[request.InstanceID] = exampleInstance

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *Broker) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	// Your deprovision logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	glog.Infof("Deprovissioning instance %q for %q/%q...", request.InstanceID, request.ServiceID, request.PlanID)
	instance, ok := b.instances[request.InstanceID]
	if !ok {
		msg := fmt.Sprintf("Instance %q not found", request.InstanceID)
		glog.Infoln(msg)

		return nil, errors.New(msg)
	}

	err := b.Client.Deprovision(request.ServiceID, instance.DbObjName)
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}
	delete(b.instances, request.InstanceID)

	response := broker.DeprovisionResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}
	glog.Infoln("Deprovisioning complete")

	return &response, nil
}

func (b *Broker) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	// Your last-operation logic goes here

	return nil, nil
}

func (b *Broker) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
	// Your bind logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	glog.Infof("Binding instance %q for %q/%q...", request.InstanceID, request.ServiceID, request.PlanID)
	instance, ok := b.instances[request.InstanceID]
	if !ok {
		msg := fmt.Sprintf("Instance %q not found", request.InstanceID)
		glog.Errorln(msg)

		return nil, errors.New(msg)
	}

	creds, err := b.Client.Bind(instance.DbObjName, request.ServiceID, request.PlanID, request.Parameters, instance.Params)
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}

	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: creds,
		},
	}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}
	glog.Infoln("Binding complete")

	return &response, nil
}

func (b *Broker) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	// nothing to do

	response := broker.UnbindResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *Broker) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	// Not supported, do nothing

	response := broker.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *Broker) ValidateBrokerAPIVersion(version string) error {
	return nil
}

// example types

// exampleInstance is intended as an example of a type that holds information about a service instance
type exampleInstance struct {
	ID        string
	ServiceID string
	PlanID    string
	Params    map[string]interface{}
	DbObjName string
}

func (i *exampleInstance) Match(other *exampleInstance) bool {
	return reflect.DeepEqual(i, other)
}
