package dormantdatabase

import (
	"time"

	crdutils "github.com/appscode/kutil/apiextensions/v1beta1"
	"github.com/appscode/kutil/tools/queue"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	kubedb_informers "github.com/kubedb/apimachinery/client/informers/externalversions/kubedb/v1alpha1"
	api_listers "github.com/kubedb/apimachinery/client/listers/kubedb/v1alpha1"
	amc "github.com/kubedb/apimachinery/pkg/controller"
	"github.com/kubedb/apimachinery/pkg/eventer"
	crd_api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

type Controller struct {
	*amc.Controller
	amc.Config
	// Deleter interface
	deleter amc.Deleter
	// tweakListOptions for watcher
	tweakListOptions func(*metav1.ListOptions)
	// Event Recorder
	recorder record.EventRecorder
	// DormantDatabase
	ddbLister api_listers.DormantDatabaseLister
}

// NewController creates a new DormantDatabase Controller
func NewController(
	controller *amc.Controller,
	deleter amc.Deleter,
	config amc.Config,
	tweakListOptions func(*metav1.ListOptions),
) *Controller {
	// return new DormantDatabase Controller
	return &Controller{
		Controller:       controller,
		deleter:          deleter,
		Config:           config,
		tweakListOptions: tweakListOptions,
		recorder:         eventer.NewEventRecorder(controller.Client, "DormantDatabase Controller"),
	}
}

func (c *Controller) EnsureCustomResourceDefinitions() error {
	crd := []*crd_api.CustomResourceDefinition{
		api.DormantDatabase{}.CustomResourceDefinition(),
	}
	return crdutils.RegisterCRDs(c.ApiExtKubeClient, crd)
}

func (c *Controller) InitInformer() cache.SharedIndexInformer {
	return c.KubedbInformerFactory.InformerFor(&api.DormantDatabase{}, func(client cs.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
		return kubedb_informers.NewFilteredDormantDatabaseInformer(
			client,
			c.WatchNamespace,
			resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			c.tweakListOptions,
		)
	})
}

func (c *Controller) AddEventHandlerFunc(selector labels.Selector) *queue.Worker {
	c.addEventHandler(selector)
	return c.DrmnQueue
}
