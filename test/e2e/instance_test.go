package e2e

import (
	"time"

	"github.com/appscode/service-broker/test/e2e/framework"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// how long to wait for an instance to be deleted.
	instanceDeleteTimeout = 60 * time.Second
)

func newTestInstance(name, namespace, serviceClassName, planName string) *v1beta1.ServiceInstance {
	return &v1beta1.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServicePlanExternalName:  planName,
				ClusterServiceClassExternalName: serviceClassName,
			},
		},
	}
}

// createInstance in the specified namespace
func createInstance(c clientset.Interface, namespace string, instance *v1beta1.ServiceInstance) (*v1beta1.ServiceInstance, error) {
	return c.ServicecatalogV1beta1().ServiceInstances(namespace).Create(instance)
}

// deleteInstance with the specified namespace and name
func deleteInstance(c clientset.Interface, namespace, name string) error {
	return c.ServicecatalogV1beta1().ServiceInstances(namespace).Delete(name, nil)
}

// waitForInstanceToBeDeleted waits for the instance to be removed.
func waitForInstanceToBeDeleted(c clientset.Interface, namespace, name string) error {
	return wait.Poll(framework.Poll, instanceDeleteTimeout, func() (bool, error) {
		_, err := c.ServicecatalogV1beta1().ServiceInstances(namespace).Get(name, metav1.GetOptions{})
		if err == nil {
			framework.Logf("waiting for service instance %s to be deleted", name)
			return false, nil
		}
		if errors.IsNotFound(err) {
			framework.Logf("verified service instance %s is deleted", name)
			return true, nil
		}
		return false, err
	})
}

var _ = Describe("[service-catalog] ServiceInstance", func() {
	var (
		f                 *framework.Invocation
		instanceName      string
		instanceNamespace string
	)

	BeforeEach(func() {
		f = root.Invoke()
		instanceName = f.BaseName + "-instance"
		instanceNamespace = f.Namespace.Name
	})

	It("should verify an Instance can be deleted if referenced service class does not exist.", func() {
		By("Creating an Instance")
		instance := newTestInstance(instanceName, instanceNamespace, "no-service-class", "no-service-plan")
		instance, err := createInstance(f.ServiceCatalogClient, instanceNamespace, instance)
		Expect(err).NotTo(HaveOccurred())
		By("Deleting the Instance")
		err = deleteInstance(f.ServiceCatalogClient, instanceNamespace, instanceName)
		Expect(err).NotTo(HaveOccurred())
		err = waitForInstanceToBeDeleted(f.ServiceCatalogClient, instanceNamespace, instanceName)
		Expect(err).NotTo(HaveOccurred())
	})
})
