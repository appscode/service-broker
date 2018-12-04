package e2e

import (
	kubedb_util "github.com/appscode/service-broker/pkg/db-broker"
	"github.com/appscode/service-broker/test/e2e/framework"
	"github.com/appscode/service-broker/test/util"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[service-catalog]", func() {
	var (
		f *framework.Invocation

		brokerName      string
		brokerNamespace string

		serviceclassName  string
		serviceclassID    string
		serviceplanName   string
		serviceplanID     string
		instanceName      string
		bindingName       string
		bindingsecretName string
		waitForCRDBeReady func() error

		test func()
	)

	BeforeEach(func() {
		f = root.Invoke()

		brokerName = f.BaseName
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

		By("Creating configmap for catalogs")
		cm, err := NewCatalogConfigMap(brokerName, brokerNamespace)
		Expect(err).NotTo(HaveOccurred())
		_, err = f.KubeClient.CoreV1().
			ConfigMaps(brokerNamespace).
			Create(cm)
		Expect(err).NotTo(HaveOccurred())

		By("Creating a service broker deployment")
		deploy, err := f.KubeClient.AppsV1().
			Deployments(brokerNamespace).
			Create(NewServiceBrokerDeployment(brokerName, brokerNamespace, brokerImageFlag, storageClass))
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

		test = func() {
			By("Making sure the ServiceBroker does not exist before creating it")
			if _, err := f.ServiceCatalogClient.ServicecatalogV1beta1().ClusterServiceBrokers().Get(brokerName, metav1.GetOptions{}); err == nil {
				By("deleting the ServiceBroker if it exists")
				err = f.ServiceCatalogClient.ServicecatalogV1beta1().ClusterServiceBrokers().Delete(brokerName, nil)
				Expect(err).NotTo(HaveOccurred(), "failed to delete the broker")

				By("Waiting for the ServiceBroker to not exist after deleting it")
				err = util.WaitForBrokerToNotExist(f.ServiceCatalogClient.ServicecatalogV1beta1(), brokerName)
				Expect(err).NotTo(HaveOccurred())
			}

			By("Creating a ClusterServiceBroker")
			url := "http://" + brokerName + "." + brokerNamespace + ".svc.cluster.local"
			broker := &v1beta1.ClusterServiceBroker{
				ObjectMeta: metav1.ObjectMeta{
					Name: brokerName,
				},
				Spec: v1beta1.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
						URL: url,
					},
				},
			}
			broker, err := f.ServiceCatalogClient.ServicecatalogV1beta1().ClusterServiceBrokers().Create(broker)
			Expect(err).NotTo(HaveOccurred(), "failed to create ClusterServiceBroker")

			By("Waiting for ClusterServiceBroker to be ready")
			err = util.WaitForBrokerCondition(f.ServiceCatalogClient.ServicecatalogV1beta1(),
				broker.Name,
				v1beta1.ServiceBrokerCondition{
					Type:   v1beta1.ServiceBrokerConditionReady,
					Status: v1beta1.ConditionTrue,
				},
			)
			Expect(err).NotTo(HaveOccurred(), "failed to wait ClusterServiceBroker to be ready")

			By("Waiting for ClusterServiceClass to be ready")
			err = util.WaitForClusterServiceClassToExist(f.ServiceCatalogClient.ServicecatalogV1beta1(), serviceclassID)
			Expect(err).NotTo(HaveOccurred(), "failed to wait serviceclass to be ready")

			By("Waiting for ClusterServicePlan to be ready")
			err = util.WaitForClusterServicePlanToExist(f.ServiceCatalogClient.ServicecatalogV1beta1(), serviceplanID)
			Ω(err).ShouldNot(HaveOccurred(), "serviceplan never became ready")

			// Provisioning a ServiceInstance and binding to it
			//By("Creating a namespace")
			//testnamespace, err := framework.CreateKubeNamespace(testns, f.KubeClient)
			//Expect(err).NotTo(HaveOccurred(), "failed to create kube namespace")

			By("Creating a ServiceInstance")
			instance := &v1beta1.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      instanceName,
					Namespace: brokerNamespace,
				},
				Spec: v1beta1.ServiceInstanceSpec{
					PlanReference: v1beta1.PlanReference{
						ClusterServiceClassExternalName: serviceclassName,
						ClusterServicePlanExternalName:  serviceplanName,
					},
				},
			}
			instance, err = f.ServiceCatalogClient.ServicecatalogV1beta1().ServiceInstances(brokerNamespace).Create(instance)
			Expect(err).NotTo(HaveOccurred(), "failed to create instance")
			Expect(instance).NotTo(BeNil())

			By("Waiting for ServiceInstance to be ready")
			err = util.WaitForInstanceCondition(f.ServiceCatalogClient.ServicecatalogV1beta1(),
				brokerNamespace,
				instanceName,
				v1beta1.ServiceInstanceCondition{
					Type:   v1beta1.ServiceInstanceConditionReady,
					Status: v1beta1.ConditionTrue,
				},
			)
			Expect(err).NotTo(HaveOccurred(), "failed to wait instance to be ready")

			By("Waiting for database crd obj to be ready")
			Expect(waitForCRDBeReady()).NotTo(HaveOccurred())

			// Make sure references have been resolved
			By("References should have been resolved before ServiceInstance is ready ")
			sc, err := f.ServiceCatalogClient.ServicecatalogV1beta1().ServiceInstances(brokerNamespace).Get(instanceName, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred(), "failed to get ServiceInstance after binding")
			Expect(sc.Spec.ClusterServiceClassRef).NotTo(BeNil())
			Expect(sc.Spec.ClusterServicePlanRef).NotTo(BeNil())
			Expect(sc.Spec.ClusterServiceClassRef.Name).To(Equal(serviceclassID))
			Expect(sc.Spec.ClusterServicePlanRef.Name).To(Equal(serviceplanID))

			// Binding to the ServiceInstance
			By("Creating a ServiceBinding")
			binding := &v1beta1.ServiceBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      bindingName,
					Namespace: brokerNamespace,
				},
				Spec: v1beta1.ServiceBindingSpec{
					ServiceInstanceRef: v1beta1.LocalObjectReference{
						Name: instanceName,
					},
					SecretName: bindingsecretName,
				},
			}
			binding, err = f.ServiceCatalogClient.ServicecatalogV1beta1().ServiceBindings(brokerNamespace).Create(binding)
			Expect(err).NotTo(HaveOccurred(), "failed to create binding")
			Expect(binding).NotTo(BeNil())

			By("Waiting for ServiceBinding to be ready")
			_, err = util.WaitForBindingCondition(f.ServiceCatalogClient.ServicecatalogV1beta1(),
				brokerNamespace,
				bindingName,
				v1beta1.ServiceBindingCondition{
					Type:   v1beta1.ServiceBindingConditionReady,
					Status: v1beta1.ConditionTrue,
				},
			)
			Expect(err).NotTo(HaveOccurred(), "failed to wait binding to be ready")

			By("Secret should have been created after binding")
			err = framework.WaitForCreatingSecret(f.KubeClient, bindingsecretName, brokerNamespace)
			Expect(err).NotTo(HaveOccurred(), "failed to create secret after binding")

			// Unbinding from the ServiceInstance
			By("Deleting the ServiceBinding")
			err = f.ServiceCatalogClient.ServicecatalogV1beta1().ServiceBindings(brokerNamespace).Delete(bindingName, nil)
			Expect(err).NotTo(HaveOccurred(), "failed to delete the binding")

			By("Waiting for ServiceBinding to not exist")
			err = util.WaitForBindingToNotExist(f.ServiceCatalogClient.ServicecatalogV1beta1(), brokerNamespace, bindingName)
			Expect(err).NotTo(HaveOccurred())

			By("Secret should been deleted after delete the binding")
			_, err = f.KubeClient.CoreV1().Secrets(brokerNamespace).Get(bindingsecretName, metav1.GetOptions{})
			Expect(err).To(HaveOccurred())

			// Deprovisioning the ServiceInstance
			//By("Patching the ServiceInstance")
			//err = util.WaitForInstanceToBePatched(f.ServiceCatalogClient.ServicecatalogV1beta1(), instance)
			//Expect(err).NotTo(HaveOccurred(), "failed to patch the instance")

			By("Deleting the ServiceInstance")
			err = f.ServiceCatalogClient.ServicecatalogV1beta1().ServiceInstances(brokerNamespace).Delete(instanceName, nil)
			Expect(err).NotTo(HaveOccurred(), "failed to delete the instance")

			By("Waiting for ServiceInstance to not exist")
			err = util.WaitForInstanceToNotExist(f.ServiceCatalogClient.ServicecatalogV1beta1(), brokerNamespace, instanceName)
			Expect(err).NotTo(HaveOccurred())

			// Deleting ClusterServiceBroker and ClusterServiceClass
			By("Deleting the ClusterServiceBroker")
			err = f.ServiceCatalogClient.ServicecatalogV1beta1().ClusterServiceBrokers().Delete(brokerName, nil)
			Expect(err).NotTo(HaveOccurred(), "failed to delete the broker")

			By("Waiting for ClusterServiceBroker to not exist")
			err = util.WaitForBrokerToNotExist(f.ServiceCatalogClient.ServicecatalogV1beta1(), brokerName)
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for ClusterServiceClass to not exist")
			err = util.WaitForClusterServiceClassToNotExist(f.ServiceCatalogClient.ServicecatalogV1beta1(), serviceclassID)
			Expect(err).NotTo(HaveOccurred())

			By("Waiting for ClusterServicePlan to not exist")
			err = util.WaitForClusterServicePlanToNotExist(f.ServiceCatalogClient.ServicecatalogV1beta1(), serviceplanID)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterEach(func() {
		//rc, err := f.KubeClient.CoreV1().Pods(brokerNamespace).GetLogs(brokerPodName, &v1.PodLogOptions{}).Stream()
		//defer rc.Close()
		//if err != nil {
		//	framework.Logf("Error getting logs for pod %s: %v", brokerName, err)
		//} else {
		//	buf := new(bytes.Buffer)
		//	buf.ReadFrom(rc)
		//	framework.Logf("Pod %s has the following logs:\n%sEnd %s logs", brokerName, buf.String(), brokerName)
		//}

		By("Deleting the service account")
		err := f.KubeClient.CoreV1().ServiceAccounts(brokerNamespace).Delete(brokerName, nil)
		Expect(err).NotTo(HaveOccurred())
		By("Deleting the custerrolebinding")
		err = f.KubeClient.RbacV1().ClusterRoleBindings().Delete(brokerName, nil)
		Expect(err).NotTo(HaveOccurred())
		By("Deleting the configmap for catalog")
		err = f.KubeClient.CoreV1().ConfigMaps(brokerNamespace).Delete(brokerName, nil)
		Expect(err).NotTo(HaveOccurred())
		By("Deleting the service broker deployment")
		err = f.KubeClient.AppsV1().Deployments(brokerNamespace).Delete(brokerName, nil)
		Expect(err).NotTo(HaveOccurred())
		By("Deleting the user broker service")
		err = f.KubeClient.CoreV1().Services(f.Namespace.Name).Delete(brokerName, nil)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Test MySQL broker service", func() {
		JustBeforeEach(func() {
			serviceclassName = "mysql"
			serviceclassID = "mysql"

			instanceName = "test-mysqldb"
			bindingName = "test-mysql-binding"
			bindingsecretName = "test-mysql-secret"
			waitForCRDBeReady = func() error {
				my, err := f.KubedbClient.MySQLs(brokerNamespace).List(metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				return kubedb_util.WaitForMySQLBeReady(f.KubedbClient, my.Items[0].Name, brokerNamespace)
			}
		})

		It("Runs through the default plan", func() {
			serviceplanName = "default"
			serviceplanID = "mysql-8-0"
			test()
		})
	})

	Context("Test PostgreSQL broker service", func() {
		JustBeforeEach(func() {
			serviceclassName = "postgresql"
			serviceclassID = "postgresql"

			instanceName = "test-postgresqldb"
			bindingName = "test-postgresql-binding"
			bindingsecretName = "test-postgresql-secret"
			waitForCRDBeReady = func() error {
				pg, err := f.KubedbClient.Postgreses(brokerNamespace).List(metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				return kubedb_util.WaitForPostgreSQLBeReady(f.KubedbClient, pg.Items[0].Name, brokerNamespace)
			}

		})

		It("Runs through the default plan", func() {
			serviceplanName = "default"
			serviceplanID = "postgresql-10-2"
			test()
		})

		It("Runs through the ha-postgresql plan", func() {
			serviceplanName = "ha-postgresql"
			serviceplanID = "ha-postgresql-10-2"
			test()
		})
	})

	Context("Test Elasticsearch broker service", func() {
		JustBeforeEach(func() {
			serviceclassName = "elasticsearch"
			serviceclassID = "elasticsearch"

			instanceName = "test-elasticsearchdb"
			bindingName = "test-elasticsearch-binding"
			bindingsecretName = "test-elasticsearch-secret"
			waitForCRDBeReady = func() error {
				es, err := f.KubedbClient.Elasticsearches(brokerNamespace).List(metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				return kubedb_util.WaitForElasticsearchBeReady(f.KubedbClient, es.Items[0].Name, brokerNamespace)
			}

		})

		It("Runs through the default plan", func() {
			serviceplanName = "default"
			serviceplanID = "elasticsearch-6-3"
			test()
		})

		It("Runs through the elasticsearch-cluster plan", func() {
			serviceplanName = "elasticsearch-cluster"
			serviceplanID = "elasticsearch-cluster-6-3"
			test()
		})
	})

	Context("Test MongoDb broker service", func() {
		JustBeforeEach(func() {
			serviceclassName = "mongodb"
			serviceclassID = "mongodb"

			instanceName = "test-mongodb"
			bindingName = "test-mongodb-binding"
			bindingsecretName = "test-mongodb-secret"
			waitForCRDBeReady = func() error {
				mg, err := f.KubedbClient.MongoDBs(brokerNamespace).List(metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				return kubedb_util.WaitForMongoDbBeReady(f.KubedbClient, mg.Items[0].Name, brokerNamespace)
			}
		})

		It("Runs through the default plan", func() {
			serviceplanName = "default"
			serviceplanID = "mongodb-3-6"
			test()
		})

		It("Runs through the mongodb-cluster plan", func() {
			serviceplanName = "mongodb-cluster"
			serviceplanID = "mongodb-cluster-3-6"
			test()
		})
	})

	Context("Test Redis broker service", func() {
		JustBeforeEach(func() {
			serviceclassName = "redis"
			serviceclassID = "redis"

			instanceName = "test-redisdb"
			bindingName = "test-redis-binding"
			bindingsecretName = "test-redis-secret"
			waitForCRDBeReady = func() error {
				rd, err := f.KubedbClient.Redises(brokerNamespace).List(metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				return kubedb_util.WaitForRedisBeReady(f.KubedbClient, rd.Items[0].Name, brokerNamespace)
			}
		})

		It("Runs through the default plan", func() {
			serviceplanName = "default"
			serviceplanID = "redis-4-0"
			test()
		})
	})

	Context("Test Memcached broker service", func() {
		JustBeforeEach(func() {
			serviceclassName = "memcached"
			serviceclassID = "memcached"

			instanceName = "test-memcachedb"
			bindingName = "test-memcached-binding"
			bindingsecretName = "test-memcached-secret"
			waitForCRDBeReady = func() error {
				mc, err := f.KubedbClient.Memcacheds(brokerNamespace).List(metav1.ListOptions{})
				Expect(err).NotTo(HaveOccurred())
				return kubedb_util.WaitForMemcachedBeReady(f.KubedbClient, mc.Items[0].Name, brokerNamespace)
			}
		})

		It("Runs through the default plan", func() {
			serviceplanName = "default"
			serviceplanID = "memcached-1-5-4"
			test()
		})
	})
})
