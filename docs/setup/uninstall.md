# Uninstall Service Broker

To uninstall Service broker, run the following command:

```console
$ curl -fsSL https://raw.githubusercontent.com/kubedb/service-broker/master/hack/deploy/service-broker.sh | bash -s -- uninstall
...

checking kubeconfig context
minikube

validatingwebhookconfiguration.admissionregistration.k8s.io "validators.kubedb.com" deleted
mutatingwebhookconfiguration.admissionregistration.k8s.io "mutators.kubedb.com" deleted
apiservice.apiregistration.k8s.io "v1alpha1.mutators.kubedb.com" deleted
apiservice.apiregistration.k8s.io "v1alpha1.validators.kubedb.com" deleted
deployment.extensions "kubedb-operator" deleted
service "kubedb-operator" deleted
secret "kubedb-server-cert" deleted
serviceaccount "kubedb-operator" deleted
clusterrolebinding.rbac.authorization.k8s.io "kubedb-operator" deleted
clusterrolebinding.rbac.authorization.k8s.io "kubedb-server-auth-delegator" deleted
clusterrole.rbac.authorization.k8s.io "kubedb-operator" deleted
rolebinding.rbac.authorization.k8s.io "kubedb-server-extension-server-authentication-reader" deleted

Successfully uninstalled KubeDB!

service "service-broker" deleted
deployment.extensions "service-broker" deleted
serviceaccount "service-broker" deleted
clusterrolebinding.rbac.authorization.k8s.io "service-broker" deleted

waiting for service-broker pod to stop running
clusterservicebroker.servicecatalog.k8s.io "service-broker" deleted
namespace "service-broker" deleted

Successfully uninstalled service-broker!
```