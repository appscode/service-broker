package broker

import (
	"net/http"
	"sync"

	dbsvc "github.com/appscode/service-broker/pkg/kubedb"
	"github.com/golang/glog"
	svcat_cs "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/pmorie/osb-broker-lib/pkg/broker"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// NewBroker is a hook that is called with the Options the program is run
// with. NewBroker is the place where you will initialize your
// Broker Logic the parameters passed in.
func NewBroker(s *ExtraOptions) (*Broker, error) {
	config, err := clientcmd.BuildConfigFromFlags("", s.KubeConfig)
	if err != nil {
		return nil, err
	}
	config.Burst = 100
	config.QPS = 100

	svccatClient, err := svcat_cs.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dbClient := dbsvc.NewClient(config, s.StorageClass)
	// For example, if your Broker Logic requires a parameter from the command
	// line, you would unpack it from the Options and set it on the
	// Broker Logic here.
	return &Broker{
		Client:       dbClient,
		svccatClient: svccatClient,
		async:        s.Async,
		catalogPath:  s.CatalogPath,
		catalogNames: s.CatalogNames,
	}, nil
}

// Broker provides an implementation of broker.Interface
type Broker struct {
	Client *dbsvc.Client

	svccatClient svcat_cs.ServicecatalogV1beta1Interface

	// Indicates if the broker should handle the requests asynchronously.
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
	b.Lock()
	defer b.Unlock()

	namespace := request.Context["namespace"].(string)
	response := broker.ProvisionResponse{}
	curProvisionInfo := &dbsvc.ProvisionInfo{
		InstanceID: request.InstanceID,
		ServiceID:  request.ServiceID,
		PlanID:     request.PlanID,
		Params:     request.Parameters,
		Namespace:  namespace,
	}

	// use name of ServiceInstance as instance crd name
	// ref: https://github.com/kubernetes-incubator/service-catalog/issues/2532
	svcinstances, err := b.svccatClient.ServiceInstances(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, svcinstance := range svcinstances.Items {
		if svcinstance.Spec.ExternalID == request.InstanceID {
			curProvisionInfo.InstanceName = svcinstance.Name
		}
	}
	if curProvisionInfo.InstanceName == "" {
		return nil, errors.Errorf("failed get name of ServiceInstance %s/%s", namespace, request.InstanceID)
	}

	// Check to see if this is the same instance
	provisionInfo, err := b.Client.GetProvisionInfo(request.InstanceID, request.ServiceID)
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
	err = b.Client.Provision(*curProvisionInfo)
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
	provisionInfo, err := b.Client.GetProvisionInfo(request.InstanceID, request.ServiceID)
	if err != nil {
		return nil, err
	} else if provisionInfo == nil {
		return nil, errors.Errorf("Instance %q not found", request.InstanceID)
	}

	err = b.Client.Deprovision(request.ServiceID, provisionInfo.InstanceName, provisionInfo.Namespace)
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
	provisionInfo, err := b.Client.GetProvisionInfo(request.InstanceID, request.ServiceID)
	if err != nil {
		return nil, errors.Wrapf(err, "Instance %q not found", request.InstanceID)
	}

	creds, err := b.Client.Bind(request.ServiceID, request.PlanID, request.Parameters, *provisionInfo)
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
