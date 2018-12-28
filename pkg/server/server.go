package server

import (
	"net/http"

	"github.com/appscode/service-broker/pkg/broker"
	"github.com/gorilla/mux"
	"github.com/pmorie/osb-broker-lib/pkg/metrics"
	"github.com/pmorie/osb-broker-lib/pkg/rest"
	prom "github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/version"
	genericapiserver "k8s.io/apiserver/pkg/server"
)

var (
	Scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(Scheme)
)

func init() {
	// we need to add the options to empty v1
	// TODO fix the server code to avoid this
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	// TODO: keep the generic API server from wanting this
	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

type BrokerServerConfig struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   *broker.Config
}

// BrokerServer contains state for a Kubernetes cluster master/api server.
type BrokerServer struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

func (op *BrokerServer) Run(stopCh <-chan struct{}) error {
	return op.GenericAPIServer.PrepareRun().Run(stopCh)
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *broker.Config
}

type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *BrokerServerConfig) Complete() CompletedConfig {
	completedCfg := completedConfig{
		c.GenericConfig.Complete(),
		c.ExtraConfig,
	}

	completedCfg.GenericConfig.Version = &version.Info{
		Major: "1",
		Minor: "1",
	}

	return CompletedConfig{&completedCfg}
}

// New returns a new instance of BrokerServer from the given config.
func (c completedConfig) New() (*BrokerServer, error) {
	genericServer, err := c.GenericConfig.New("service-broker", genericapiserver.NewEmptyDelegate()) // completion is done in Complete, no need for a second time
	if err != nil {
		return nil, err
	}

	b, err := c.ExtraConfig.New()
	if err != nil {
		return nil, err
	}

	// Prometheus metrics
	reg := prom.NewRegistry()
	osbMetrics := metrics.New()
	reg.MustRegister(osbMetrics)

	api, err := rest.NewAPISurface(b, osbMetrics)
	if err != nil {
		return nil, err
	}
	genericServer.Handler.NonGoRestfulMux.HandlePrefix("/v2/", registerAPIHandlers(api))

	s := &BrokerServer{
		GenericAPIServer: genericServer,
	}
	return s, nil
}

// registerAPIHandlers registers the APISurface endpoints and handlers.
func registerAPIHandlers(api *rest.APISurface) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/catalog", api.GetCatalogHandler).Methods("GET")
	router.HandleFunc("/service_instances/{instance_id}/last_operation", api.LastOperationHandler).Methods("GET")
	router.HandleFunc("/service_instances/{instance_id}", api.ProvisionHandler).Methods("PUT")
	router.HandleFunc("/service_instances/{instance_id}", api.DeprovisionHandler).Methods("DELETE")
	router.HandleFunc("/service_instances/{instance_id}", api.UpdateHandler).Methods("PATCH")
	router.HandleFunc("/service_instances/{instance_id}/service_bindings/{binding_id}", api.BindHandler).Methods("PUT")
	router.HandleFunc("/service_instances/{instance_id}/service_bindings/{binding_id}", api.UnbindHandler).Methods("DELETE")
	return router
}
