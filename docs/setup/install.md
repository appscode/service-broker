# Installation Guide

## Prerequisites

First you need to have the software services by AppsCode installed in the cluster. Currently AppsCode Service Broker supports the following software service:

 - Kubedb

So we need to have Kubedb installed to go forward. To install Kubedb see [here](https://kubedb.com/docs/0.9.0/setup/install/).

To check the installation for AppsCode Service Broker, we have used [Service Catalog](https://kubernetes.io/docs/concepts/extend-kubernetes/service-catalog/). So, this document assumes that you've installed Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://svc-cat.io/docs/install/). Optionally you may install the Service Catalog CLI (`svcat`) from [Installing the Service Catalog CLI](https://svc-cat.io/docs/install/#installing-the-service-catalog-cli) section.

> After satisfying the prerequisites, all commands in this document assume that you're operating out of the root of this repository.

## Install Service Broker

AppsCode Service Broker can be installed via a script or as a Helm chart.

- [Script](/docs/setup/install.md#Using-Script)
- [Helm](/docs/setup/install.md#Using-Helm)

### Using Script

To install AppsCode Service Broker in your Kubernetes cluster, run the following command:

```console
$ curl -fsSL https://raw.githubusercontent.com/appscode/service-broker/master/hack/deploy/install.sh | bash
...

configmap/kubedb created
deployment.apps/service-broker created
service/service-broker created
serviceaccount/service-broker created
clusterrole.rbac.authorization.k8s.io/service-broker created
clusterrolebinding.rbac.authorization.k8s.io/service-broker created
clusterservicebroker.servicecatalog.k8s.io/service-broker created

waiting until service-broker deployment is ready

Successfully installed service-broker in kube-system namespace!
```

After successful installation, you should have a `service-broker-***` pod running in the `kube-system` namespace.

```console
$ kubectl get pods -n kube-system | grep service-broker
service-broker-***       1/1       Running   0          48s
```

#### Customizing Installer

The installer script and associated yaml files can be found in the [hack/deploy](https://github.com/appscode/service-broker/tree/master/hack/deploy) folder. You can see the full list of flags available to installer using `-h` flag.

```console
$ curl -fsSL https://raw.githubusercontent.com/appscode/service-broker/master/hack/deploy/install.sh | bash -s -- -h
install.sh - install service-broker

install.sh [options]

options:
--------
-h, --help                    show brief help
-n, --namespace=NAMESPACE     specify namespace (default: kube-system)
    --docker-registry         docker registry used to pull service-broker image (default: appscode)
    --image-pull-secret       name of secret used to pull service-broker image
    --port                    port number at which the broker will expose
    --catalogPath             the path of catalogs for different service plans
    --catalogNames            comma separated names of the catalogs for different service plans
    --uninstall               uninstall service-broker
```

If you would like to run service broker pod in your own namespace say `my-ns` namespace, first create it and then pass the `--namespace=my-ns` flag:

```console
$ kubectl create namespace my-ns
namespace/my-ns created

$ curl -fsSL https://raw.githubusercontent.com/appscode/service-broker/master/hack/deploy/install.sh \
    | bash -s -- --namespace=my-ns
```

If you are using a private Docker registry, you need to pull the following docker images:

 - [appscode/service-broker](https://hub.docker.com/r/appscode/service-broker)

To pass the address of your private registry and optionally a image pull secret use flags `--docker-registry` and `--image-pull-secret` respectively.

```console
$ curl -fsSL https://raw.githubusercontent.com/appscode/service-broker/master/hack/deploy/install.sh \
    | bash -s -- --docker-registry=MY_REGISTRY [--image-pull-secret=SECRET_NAME]
```

### Using Helm

`Service Broker` can also be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/appscode/service-broker/tree/master/chart/service-broker) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `appscode-service-broker`:

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search appscode/service-broker
$ helm install appscode/service-broker --name appscode-service-broker --namespace kube-system
```

To see the detailed configuration options, visit [here](https://github.com/appscode/service-broker/tree/master/chart/service-broker).

### Verify installation

To check whether service broker pod has started or not, run the following command:

```console
# for script installation
$ kubectl get pods --all-namespaces -l app=appscode-service-broker --watch
NAMESPACE     NAME                                       READY   STATUS    RESTARTS   AGE
kube-system   appscode-service-broker-76c759f6c7-htqbh   0/1     Pending   0          0s
kube-system   appscode-service-broker-76c759f6c7-htqbh   0/1   ContainerCreating   0     0s
kube-system   appscode-service-broker-76c759f6c7-htqbh   0/1   Running   0     1s
kube-system   appscode-service-broker-76c759f6c7-htqbh   1/1   Running   0     7s

# for helm installation
$ NAMESPACE     NAME                                       READY   STATUS    RESTARTS   AGE
kube-system   appscode-service-broker-85795d8b6f-ntw9v   0/1     Pending   0          0s
kube-system   appscode-service-broker-85795d8b6f-ntw9v   0/1   ContainerCreating   0     0s
kube-system   appscode-service-broker-85795d8b6f-ntw9v   0/1   Running   0     6s
kube-system   appscode-service-broker-85795d8b6f-ntw9v   1/1   Running   0     12s
```

Once the pod is running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm `ClusterServiceBroker`, `ClusterServiceClass`, `ClusterServicePlan` have been registered by the `Service Broker`, follow the sections bellow:

#### Checking a ClusterServiceBroker Resource

```console
$ kubectl get clusterservicebrokers -l app=appscode-service-broker
NAME                      URL                                                            STATUS   AGE
appscode-service-broker   http://appscode-service-broker.kube-system.svc.cluster.local   Ready    1m

$ svcat get brokers
           NAME             NAMESPACE                               URL                                STATUS
+-------------------------+-----------+--------------------------------------------------------------+--------+
  appscode-service-broker               http://appscode-service-broker.kube-system.svc.cluster.local   Ready
```

After creating `ClusterServiceBroker` resource, the service catalog controller responds by querying the broker server to see what services it offers and creates a `ClusterServiceClass` for each of the services.

We can check the status of the broker:

```console
$ svcat describe broker appscode-service-broker
  Name:     appscode-service-broker
  URL:      http://appscode-service-broker.kube-system.svc.cluster.local
  Status:   Ready - Successfully fetched catalog entries from broker @ 2018-12-24 11:24:49 +0000 UTC
```

Here is the yaml configuration of this resource.

```console
kubectl get clusterservicebrokers service-broker -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceBroker
metadata:
  creationTimestamp: "2018-12-24T11:24:15Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
    chart: appscode-service-broker-0.1.0
    heritage: Tiller
    release: appscode-service-broker
  name: appscode-service-broker
  resourceVersion: "604"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/clusterservicebrokers/appscode-service-broker
  uid: 724608cf-076e-11e9-a97c-0242ac110007
spec:
  relistBehavior: Duration
  relistRequests: 0
  url: http://appscode-service-broker.kube-system.svc.cluster.local
status:
  conditions:
  - lastTransitionTime: "2018-12-24T11:24:49Z"
    message: Successfully fetched catalog entries from broker.
    reason: FetchedCatalog
    status: "True"
    type: Ready
  lastCatalogRetrievalTime: "2018-12-24T11:24:49Z"
  reconciledGeneration: 1
```

> Notice that the status says that the broker's catalog of service offerings has been successfully added to our cluster's service catalog.

#### Viewing ClusterServiceClasses

There is a `ClusterServiceClass` for each service provided by the AppsCode Service Broker. To view these `ClusterServiceClass` resources:

```console
$ kubectl get clusterserviceclasses
NAME                                   EXTERNAL-NAME   BROKER                    AGE
2010d83f-d908-4d9f-879c-ce8f5f527f2a   postgresql      appscode-service-broker   33m
315fc21c-829e-4aa1-8c16-f7921c33550d   elasticsearch   appscode-service-broker   33m
938a70c5-f2bc-4658-82dd-566bed7797e9   mysql           appscode-service-broker   33m
ccfd1c81-e59f-4875-a39f-75ba55320ce0   redis           appscode-service-broker   33m
d690058d-666c-45d8-ba98-fcb9fb47742e   mongodb         appscode-service-broker   33m
d88856cb-fe3f-4473-ba8b-641480da810f   memcached       appscode-service-broker   33m

$ svcat get classes
      NAME        NAMESPACE           DESCRIPTION
+---------------+-----------+------------------------------+
  postgresql                  KubeDB managed PostgreSQL
  elasticsearch               KubeDB managed ElasticSearch
  mysql                       KubeDB managed MySQL
  redis                       KubeDB managed Redis
  mongodb                     KubeDB managed MongoDB
  memcached                   KubeDB managed Memcached
```

Here is the details for service `mysql` from KubeDB by `AppsCode`.

```console
$ svcat describe class mysql
  Name:              mysql
  Scope:             cluster
  Description:       KubeDB managed MySQL
  Kubernetes Name:   938a70c5-f2bc-4658-82dd-566bed7797e9
  Status:            Active
  Tags:
  Broker:            appscode-service-broker

Plans:
     NAME               DESCRIPTION
+------------+--------------------------------+
  demo-mysql   Demo MySQL database
  mysql        MySQL database with custom
               specification
```

Here is the yaml configuration of `ClusterServiceClass` named `mysql`.

```console
kubectl get clusterserviceclasses 938a70c5-f2bc-4658-82dd-566bed7797e9 -o yaml
```

> In this command `938a70c5-f2bc-4658-82dd-566bed7797e9` is the name for `ClusterServiceClass` resource having `mysql` as `EXTERNAL-NAME`. It is set by service catalog controller from the `services[].id` field of catalog response returned by broker server. We took this name from `$ kubectl get clusterserviceclasses` command.

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceClass
metadata:
  creationTimestamp: "2018-12-24T11:24:46Z"
  name: 938a70c5-f2bc-4658-82dd-566bed7797e9
  ownerReferences:
  - apiVersion: servicecatalog.k8s.io/v1beta1
    blockOwnerDeletion: false
    controller: true
    kind: ClusterServiceBroker
    name: appscode-service-broker
    uid: 724608cf-076e-11e9-a97c-0242ac110007
  resourceVersion: "588"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/clusterserviceclasses/938a70c5-f2bc-4658-82dd-566bed7797e9
  uid: 84b4998b-076e-11e9-a97c-0242ac110007
spec:
  bindable: true
  bindingRetrievable: false
  clusterServiceBrokerName: appscode-service-broker
  description: KubeDB managed MySQL
  externalID: 938a70c5-f2bc-4658-82dd-566bed7797e9
  externalMetadata:
    displayName: KubeDB managed MySQL
    imageUrl: https://cdn.appscode.com/images/logo/databases/mysql.png
  externalName: mysql
  planUpdatable: true
status:
  removedFromBrokerCatalog: false
```

#### Viewing ClusterServicePlans

There is also a `ClusterServicePlan` for each plan under the broker's services. To view these `ClusterServicePlan` resources:

```console
$ kubectl get clusterserviceplans
NAME                                   EXTERNAL-NAME                BROKER                    CLASS                                  AGE
13373a9b-d5f5-4d9a-88df-d696bbc19071   postgresql                   appscode-service-broker   2010d83f-d908-4d9f-879c-ce8f5f527f2a   52m
1fd1abf1-e8e1-44a2-8214-bf0fd1ce9417   demo-mysql                   appscode-service-broker   938a70c5-f2bc-4658-82dd-566bed7797e9   52m
2f05622b-724d-458f-abc8-f223b1afa0b9   demo-elasticsearch-cluster   appscode-service-broker   315fc21c-829e-4aa1-8c16-f7921c33550d   52m
41818203-0e2d-4d30-809f-a60c8c73dae8   demo-ha-postgresql           appscode-service-broker   2010d83f-d908-4d9f-879c-ce8f5f527f2a   52m
45716530-cadb-4247-b06a-24a34200d734   redis                        appscode-service-broker   ccfd1c81-e59f-4875-a39f-75ba55320ce0   52m
498c12a6-7a68-4983-807b-75737f99062a   demo-mongodb                 appscode-service-broker   d690058d-666c-45d8-ba98-fcb9fb47742e   52m
4b6ad8a7-272e-4cfd-bb38-5b9d4bd3962f   demo-redis                   appscode-service-broker   ccfd1c81-e59f-4875-a39f-75ba55320ce0   52m
6af19c54-7757-42e5-bb74-b8350037c4a2   demo-mongodb-cluster         appscode-service-broker   d690058d-666c-45d8-ba98-fcb9fb47742e   52m
6ed1ab9e-a640-4f26-9328-423b2e3816d7   mysql                        appscode-service-broker   938a70c5-f2bc-4658-82dd-566bed7797e9   52m
6fa212e2-e043-4ae9-91c2-8e5c4403d894   elasticsearch                appscode-service-broker   315fc21c-829e-4aa1-8c16-f7921c33550d   52m
af1ce2dc-5734-4e41-aaa2-8aa6a58d688f   demo-memcached               appscode-service-broker   d88856cb-fe3f-4473-ba8b-641480da810f   52m
c4bcf392-7ebb-4623-a79d-13d00d761d56   demo-postgresql              appscode-service-broker   2010d83f-d908-4d9f-879c-ce8f5f527f2a   52m
c4e99557-3a81-452e-b9cf-660f01c155c0   demo-elasticsearch           appscode-service-broker   315fc21c-829e-4aa1-8c16-f7921c33550d   52m
d40e49b2-f8fb-4d47-96d3-35089bd0942d   memcached                    appscode-service-broker   d88856cb-fe3f-4473-ba8b-641480da810f   52m
e8f87ba6-0711-42db-a663-a3c75b78a541   mongodb                      appscode-service-broker   d690058d-666c-45d8-ba98-fcb9fb47742e   52m

$              NAME              NAMESPACE       CLASS                DESCRIPTION
+----------------------------+-----------+---------------+--------------------------------+
  postgresql                               postgresql      PostgreSQL database with
                                                           custom specification
  demo-ha-postgresql                       postgresql      Demo HA PostgreSQL database
  demo-postgresql                          postgresql      Demo Standalone PostgreSQL
                                                           database
  demo-elasticsearch                       elasticsearch   Demo Standalone Elasticsearch
                                                           database
  demo-elasticsearch-cluster               elasticsearch   Demo Elasticsearch cluster
  elasticsearch                            elasticsearch   Elasticsearch cluster with
                                                           custom specification
  demo-mysql                               mysql           Demo MySQL database
  mysql                                    mysql           MySQL database with custom
                                                           specification
  redis                                    redis           Redis with custom
                                                           specification
  demo-redis                               redis           Demo Redis
  demo-mongodb-cluster                     mongodb         Demo MongoDB cluster
  demo-mongodb                             mongodb         Demo Standalone MongoDB
                                                           database
  mongodb                                  mongodb         MongoDB database with custom
                                                           specification
  demo-memcached                           memcached       Demo Memcached
  memcached                                memcached       Memcached with custom
                                                           specification
```

As an example, to view the details of the `demo-mysql` plan of `mysql` class:

```console
$ svcat describe plan mysql/demo-mysql --scope cluster
  Name:              demo-mysql
  Description:       Demo MySQL database
  Kubernetes Name:   1fd1abf1-e8e1-44a2-8214-bf0fd1ce9417
  Status:            Active
  Free:              true
  Class:             mysql

Instances:
No instances defined
```

> Here we,ve used `--scope` flag to specify that our `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServiceBroker` resources are cluster scoped (not namespaced scope)

Here is the yaml configuration of `ClusterServicePlan` named `demo-mysql` of `ClusterServiceClass` named `mysql`.

```console
kubectl get clusterserviceplans 1fd1abf1-e8e1-44a2-8214-bf0fd1ce9417 -o yaml
```

> In this command `1fd1abf1-e8e1-44a2-8214-bf0fd1ce9417` is the name for `ClusterServicePlan` resource with having `demo-mysql` as `EXTERNAL-NAME`. It is set by service catalog controller from the `services[].plans[].id` field of catalog response returned by broker server. We took this name from `$ kubectl get clusterserviceplans` command.

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServicePlan
metadata:
  creationTimestamp: "2018-12-24T11:24:48Z"
  name: 1fd1abf1-e8e1-44a2-8214-bf0fd1ce9417
  ownerReferences:
  - apiVersion: servicecatalog.k8s.io/v1beta1
    blockOwnerDeletion: false
    controller: true
    kind: ClusterServiceBroker
    name: appscode-service-broker
    uid: 724608cf-076e-11e9-a97c-0242ac110007
  resourceVersion: "602"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/clusterserviceplans/1fd1abf1-e8e1-44a2-8214-bf0fd1ce9417
  uid: 8604b82a-076e-11e9-a97c-0242ac110007
spec:
  clusterServiceBrokerName: appscode-service-broker
  clusterServiceClassRef:
    name: 938a70c5-f2bc-4658-82dd-566bed7797e9
  description: Demo MySQL database
  externalID: 1fd1abf1-e8e1-44a2-8214-bf0fd1ce9417
  externalName: demo-mysql
  free: true
status:
  removedFromBrokerCatalog: false
```