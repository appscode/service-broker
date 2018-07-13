package main

import (
	"github.com/appscode/kutil/tools/clientcmd"
	svcat "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func main() {
	clientConfig, err :=  clientcmd.BuildConfigFromContext("/home/ac/.kube/config", "")
	if err != nil {
		panic(err)
	}
	serviceCatalogClient, err := svcat.NewForConfig(clientConfig)
	if err != nil {
		panic(err)
	}

	brokerName := "abc"
	brokerNamespace := "abc-ns"
	url := "http://" + brokerName + "." + brokerNamespace + ".svc.cluster.local"
	_, err = serviceCatalogClient.ServicecatalogV1beta1().ClusterServiceBrokers().Create(newTestBroker(brokerName, url))
	if err!= nil {
		panic(err)
	}
}
