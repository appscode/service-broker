package e2e

import (
	"github.com/kubedb/service-broker/test/e2e/framework"
	"github.com/kubedb/service-broker/test/util"
	v1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func newTestBroker(name, url string) *v1beta1.ClusterServiceBroker {
	return &v1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}
}

var _ = Describe("[service-catalog] ClusterServiceBroker", func() {
	var (
		f *framework.Invocation

		brokerName      string
		brokerNamespace string
		//BrokerImageFlag = brokerImageFlag
	)

	BeforeEach(func() {
		f = root.Invoke()

		brokerName = f.App()
		brokerNamespace = f.Namespace.Name

		By("Creating a service account for service broker")
		_, err := f.KubeClient.CoreV1().
			ServiceAccounts(brokerNamespace).
			Create(NewServiceBrokerServiceAccount(brokerName, brokerNamespace))
		Expect(err).NotTo(HaveOccurred())

		By("Creating a cluster-admin custerrolebinding for service broker")
		_, err = f.KubeClient.RbacV1().
			ClusterRoleBindings().
			Create(NewServiceBrokerClusterRoleBinding(brokerName, brokerNamespace))
		Expect(err).NotTo(HaveOccurred())

		By("Creating a service broker deployment")
		deploy, err := f.KubeClient.AppsV1().
			Deployments(brokerNamespace).
			Create(NewServiceBrokerDeployment(brokerName, brokerNamespace, brokerImageFlag))
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for pod to be running")
		pod, err := framework.GetBrokerPod(f.KubeClient, deploy)
		Expect(err).NotTo(HaveOccurred())
		err = framework.WaitForPodRunningInNamespace(f.KubeClient, pod)
		Expect(err).NotTo(HaveOccurred())

		By("Creating a service broker service")
		_, err = f.KubeClient.CoreV1().
			Services(f.Namespace.Name).
			Create(NewServiceBrokerService(brokerName, brokerNamespace))
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for service endpoint")
		err = framework.WaitForEndpoint(f.KubeClient, f.Namespace.Name, brokerName)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		By("Deleting the service account")
		err := f.KubeClient.CoreV1().ServiceAccounts(brokerNamespace).Delete(brokerName, nil)
		Expect(err).NotTo(HaveOccurred())
		By("Deleting the custerrolebinding")
		err = f.KubeClient.RbacV1().ClusterRoleBindings().Delete(brokerName, nil)
		Expect(err).NotTo(HaveOccurred())
		By("Deleting the service broker deployment")
		err = f.KubeClient.AppsV1().Deployments(brokerNamespace).Delete(brokerName, nil)
		Expect(err).NotTo(HaveOccurred())
		By("Deleting the user broker service")
		err = f.KubeClient.CoreV1().Services(f.Namespace.Name).Delete(brokerName, nil)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should become ready", func() {
		By("Making sure the ServiceBroker does not exist before creating it")
		if _, err := f.ServiceCatalogClient.ServicecatalogV1beta1().ClusterServiceBrokers().Get(brokerName, metav1.GetOptions{}); err == nil {
			By("deleting the ServiceBroker if it does exist")
			err = f.ServiceCatalogClient.ServicecatalogV1beta1().ClusterServiceBrokers().Delete(brokerName, nil)
			Expect(err).NotTo(HaveOccurred(), "failed to delete the broker")

			By("Waiting for the ServiceBroker to not exist after deleting it")
			err = util.WaitForBrokerToNotExist(f.ServiceCatalogClient.ServicecatalogV1beta1(), brokerName)
			Expect(err).NotTo(HaveOccurred())
		}

		By("Creating a Broker")
		url := "http://" + brokerName + "." + brokerNamespace + ".svc.cluster.local"

		broker, err := f.ServiceCatalogClient.ServicecatalogV1beta1().ClusterServiceBrokers().Create(newTestBroker(brokerName, url))
		Expect(err).NotTo(HaveOccurred())
		By("Waiting for Broker to be ready")
		err = util.WaitForBrokerCondition(f.ServiceCatalogClient.ServicecatalogV1beta1(),
			broker.Name,
			v1beta1.ServiceBrokerCondition{
				Type:   v1beta1.ServiceBrokerConditionReady,
				Status: v1beta1.ConditionTrue,
			})
		Expect(err).NotTo(HaveOccurred())

		By("Deleting the Broker")
		err = f.ServiceCatalogClient.ServicecatalogV1beta1().ClusterServiceBrokers().Delete(brokerName, nil)
		Expect(err).NotTo(HaveOccurred())

		By("Waiting for Broker to not exist")
		err = util.WaitForBrokerToNotExist(f.ServiceCatalogClient.ServicecatalogV1beta1(), brokerName)
		Expect(err).NotTo(HaveOccurred())
	})
})
