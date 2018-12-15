package broker

import (
	"net/http"
	"sync"

	"github.com/appscode/go/crypto/rand"
	db_broker "github.com/appscode/service-broker/pkg/db-broker"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
)

// NewBroker is a hook that is called with the Options the program is run
// with. NewBroker is the place where you will initialize your
// Broker Logic the parameters passed in.
func NewBroker(s *ExtraOptions) (*Broker, error) {
	brClient := db_broker.NewClient(s.KubeConfig, s.StorageClass)
	// For example, if your Broker Logic requires a parameter from the command
	// line, you would unpack it from the Options and set it on the
	// Broker Logic here.
	return &Broker{
		Client:       brClient,
		async:        s.Async,
		catalogPath:  s.CatalogPath,
		catalogNames: s.CatalogNames,
	}, nil
}

// Broker provides an implementation of broker.Interface
type Broker struct {
	Client *db_broker.Client

	// Indiciates if the broker should handle the requests asynchronously.
	async bool

	// The path for catalogs
	catalogPath string
	// names of the catalogs those will provided by the broker
	catalogNames []string

	// Synchronize go routines.
	sync.RWMutex
}

var _ broker.Interface = &Broker{}

func (b *Broker) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	// Your catalog broker logic goes here
	services, err := b.Client.GetCatalog(b.catalogPath, b.catalogNames...)
	if err != nil {
		return nil, err
	}

	return &broker.CatalogResponse{
		CatalogResponse: osb.CatalogResponse{
			Services: services,
		},
	}, nil
}

func (b *Broker) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	// Your provision logic goes here

	// example implementation:
	b.Lock()
	defer b.Unlock()

	response := broker.ProvisionResponse{}
	curProvisionInfo := &db_broker.ProvisionInfo{
		InstanceID:   request.InstanceID,
		ServiceID:    request.ServiceID,
		PlanID:       request.PlanID,
		Params:       request.Parameters,
		InstanceName: rand.WithUniqSuffix(request.PlanID),
	}

	// Check to see if this is the same instance
	provisionInfo, err := b.Client.GetProvisionInfo(b.catalogNames, request.InstanceID, request.ServiceID)
	if err != nil {
		return nil, err
	}
	if provisionInfo != nil {
		if provisionInfo.Match(curProvisionInfo) {
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

	glog.Infof("Provisioning instance %q for %q/%q...", request.InstanceID, request.ServiceID, request.PlanID)
	err = b.Client.Provision(b.catalogNames, *curProvisionInfo)
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}
	glog.Infoln("Provisioning complete")

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

	glog.Infof("Deprovisioning instance %q for %q/%q...", request.InstanceID, request.ServiceID, request.PlanID)
	provisionInfo, err := b.Client.GetProvisionInfo(b.catalogNames, request.InstanceID, request.ServiceID)
	if err != nil {
		return nil, err
	} else if provisionInfo == nil {
		return nil, errors.Errorf("Instance %q not found", request.InstanceID)
	}

	err = b.Client.Deprovision(b.catalogNames, request.ServiceID, provisionInfo.InstanceName)
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}

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
	provisionInfo, err := b.Client.GetProvisionInfo(b.catalogNames, request.InstanceID, request.ServiceID)
	if err != nil {
		return nil, errors.Wrapf(err, "Instance %q not found", request.InstanceID)
	}

	creds, err := b.Client.Bind(b.catalogNames, request.ServiceID, request.PlanID, request.Parameters, *provisionInfo)
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

	glog.Infof("Unbinding instance %q for %q/%q...", request.InstanceID, request.ServiceID, request.PlanID)
	response := broker.UnbindResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}
	glog.Infoln("Unbinding complete")

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
