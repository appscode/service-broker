# Uninstall Service Broker

To uninstall Service broker, run the following command:

```console
$ curl -fsSL https://raw.githubusercontent.com/kubedb/service-broker/master/hack/dev/build.sh | bash -s -- --uninstall [--namespace=NAMESPACE]
...
service "service-broker" deleted
deployment.extensions "service-broker" deleted
serviceaccount "service-broker" deleted
clusterrolebinding.rbac.authorization.k8s.io "service-broker" deleted

waiting for service-broker pod to stop running
clusterservicebroker.servicecatalog.k8s.io "service-broker" deleted
namespace "service-broker" deleted

Successfully uninstalled service-broker!
```