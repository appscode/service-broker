package framework

import (
	svcat "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	//"github.com/kubernetes-incubator/service-catalog/pkg/svcat/kube"
	"github.com/appscode/go/crypto/rand"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
)

// Framework supports common operations used by e2e tests; it will keep a client & a namespace for you.
type Framework struct {
	BaseName string

	// A Kubernetes and Service Catalog client
	KubeClient           kubernetes.Interface
	ServiceCatalogClient svcat.Interface
	KubedbClient         cs.KubedbV1alpha1Interface
	// Namespace in which all test resources should reside
	Namespace *corev1.Namespace
}

// NewFramework makes a new framework and sets up a BeforeEach/AfterEach for
// you (you can write additional before/after each functions).
func NewFramework(
	baseName string,
	kubeClient kubernetes.Interface,
	serviceCatalogClient svcat.Interface,
	kubedbClient cs.KubedbV1alpha1Interface) *Framework {

	f := &Framework{
		BaseName: baseName,

		KubeClient:           kubeClient,
		ServiceCatalogClient: serviceCatalogClient,
		KubedbClient:         kubedbClient,
	}

	return f
}

func (f *Framework) Invoke() *Invocation {
	return &Invocation{
		Framework: f,
		app:       rand.WithUniqSuffix(f.BaseName),
	}
}

type Invocation struct {
	*Framework
	app string
}

func (f *Invocation) App() string {
	return f.app
}
