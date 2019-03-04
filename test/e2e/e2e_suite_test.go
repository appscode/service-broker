package e2e

import (
	"testing"
	"time"

	"github.com/appscode/service-broker/test/e2e/framework"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	svcat "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"kmodules.xyz/client-go/logs"
	"kmodules.xyz/client-go/tools/clientcmd"
)

const (
	TIMEOUT = 20 * time.Minute
)

var (
	root *framework.Framework
)

func TestE2e(t *testing.T) {
	logs.InitLogs()
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(TIMEOUT)
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "e2e Suite", []Reporter{junitReporter})
}

// BeforeSuite gets a client and makes a namespace.
var _ = BeforeSuite(func() {
	By("Creating a kubernetes client")
	clientConfig, err := clientcmd.BuildConfigFromContext(options.KubeConfig, options.KubeContext)
	Expect(err).NotTo(HaveOccurred())

	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	Expect(err).NotTo(HaveOccurred())

	By("Creating a service catalog client")
	serviceCatalogClient, err := svcat.NewForConfig(clientConfig)
	Expect(err).NotTo(HaveOccurred())

	By("Creating a kubedb client")
	kubedbClient, err := cs.NewForConfig(clientConfig)
	Expect(err).NotTo(HaveOccurred())

	root = framework.NewFramework("test-broker", kubeClient, serviceCatalogClient, kubedbClient)

	By("Building a namespace api object")
	namespace, err := framework.CreateKubeNamespace(root.BaseName, root.KubeClient)
	Expect(err).NotTo(HaveOccurred())

	root.Namespace = namespace
})

// To make sure that this framework cleans up after itself, no matter what,
// we install a Cleanup action before each test and clear it after.  If we
// should abort, the AfterSuite hook should run all Cleanup actions.

// AfterEach deletes the namespace, after reading its events.
var _ = AfterSuite(func() {
	err := framework.DeleteKubeNamespace(root.KubeClient, root.Namespace.Name)
	Expect(err).NotTo(HaveOccurred())
})
