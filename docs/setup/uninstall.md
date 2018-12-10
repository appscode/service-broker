# Uninstall Service Broker

To uninstall service broker, run the following command:

```console
$ curl -fsSL https://raw.githubusercontent.com/appscode/service-broker/master/hack/deploy/install.sh | bash -s -- --uninstall
...

configmap "kubedb" deleted
service "service-broker" deleted
deployment.extensions "service-broker" deleted
serviceaccount "service-broker" deleted
clusterrolebinding.rbac.authorization.k8s.io "service-broker" deleted

waiting for service-broker pod to stop running
clusterservicebroker.servicecatalog.k8s.io "service-broker" deleted

Successfully uninstalled service-broker!
```

To uninstall the software services of AppsCode, follow the link below:

 - [Uninstall Kubedb](https://kubedb.com/docs/0.9.0-rc.2/setup/uninstall/)