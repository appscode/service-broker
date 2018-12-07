# Installation Guide

## Prerequisites

First you need to have the software services by AppsCode installed in the cluster. Currently AppsCode Service Broker supports the following software service:

 - Kubedb

So we need to have Kubedb installed to go forward. To install Kubedb see [here](https://kubedb.com/docs/0.9.0-rc.1/setup/install/).

To check the installation you need the Service Catalog onto your cluster. So, this document assumes that you've installed Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://svc-cat.io/docs/install/). Optionally you may install the Service Catalog CLI (nammed `svcat`) from the `installing-the-service-catalog-cli` section.

> After satisfying the prerequisites, all commands in this document assume that you're operating out of the root of this repository.

## Install Service Broker

`AppsCode Service Broker` can be installed via a script or as a Helm chart.

- [Script](/docs/setup/install.md#Using-Script)
- [Helm](/docs/setup/install.md#Using-Helm)

### Using Script

To install `AppsCode Service Broker` in your Kubernetes cluster, run the following command:

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
    --storage-class           name of the storage-class for database storage
    --uninstall               uninstall service-broker
```

If you would like to run service broker pod in your own namespace say `my-ns` namespace, first create it and then pass the `--namespace=my-ns` flag:

```console
$ kubectl create namespace my-ns
namespace/my-ns created

$ curl -fsSL https://raw.githubusercontent.com/appscode/service-broker/master/hack/deploy/install.sh \
    | bash -s -- --namespace=my-ns
...
```

If you are using a private Docker registry, you need to pull the following docker images:

 - [appscode/service-broker](https://hub.docker.com/r/appscode/service-broker)

To pass the address of your private registry and optionally a image pull secret use flags `--docker-registry` and `--image-pull-secret` respectively.

```console
$ curl -fsSL https://raw.githubusercontent.com/appscode/service-broker/master/hack/deploy/install.sh \
    | bash -s -- --docker-registry=MY_REGISTRY [--image-pull-secret=SECRET_NAME]
...
```

### Using Helm

`Service Broker` can also be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/appscode/service-broker/tree/master/chart/service-broker) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `my-release`:

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search appscode/service-broker
$ helm install --name my-release --namespace kube-system appscode/service-broker
...
```

To see the detailed configuration options, visit [here](https://github.com/appscode/service-broker/tree/master/chart/service-broker).

### Verify installation

To check whether service broker pod has started or not, run the following command:

```console
# for script installation
$ kubectl get pods --all-namespaces -l app=service-broker --watch
NAMESPACE        NAME                              READY   STATUS    RESTARTS   AGE
kube-system      service-broker-6f4f7b554d-hcvdd   0/1     Pending   0          0s
kube-system      service-broker-6f4f7b554d-hcvdd   0/1   Pending   0     0s
kube-system      service-broker-6f4f7b554d-hcvdd   0/1   ContainerCreating   0     0s
kube-system      service-broker-6f4f7b554d-hcvdd   1/1   Running   0     6s

# for helm installation
kube-system      my-release-service-broker-7d8cc8dcc-q7c6m   0/1   ContainerCreating   0     0s
kube-system      my-release-service-broker-7d8cc8dcc-q7c6m   1/1   Running   0     4s
```

Once the pods is running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm `ClusterServiceBroker`, `ClusterServiceClass`, `ClusterServicePlan` have been registered by the `Service Broker`, run the following command:

```console
$ kubectl get clusterservicebrokers -l app=service-broker
NAME             URL                                                      STATUS   AGE
service-broker   http://service-broker.service-broker.svc.cluster.local   Ready    1m

$ kubectl get clusterserviceclass
NAME            EXTERNAL-NAME   BROKER           AGE
elasticsearch   elasticsearch   service-broker   1m
memcached       memcached       service-broker   1m
mongodb         mongodb         service-broker   1m
mysql           mysql           service-broker   1m
postgresql      postgresql      service-broker   1m
redis           redis           service-broker   1m

$ kubectl get clusterserviceplans
NAME                        EXTERNAL-NAME           BROKER           CLASS           AGE
elasticsearch-6-3           default                 service-broker   elasticsearch   1m
elasticsearch-cluster-6-3   elasticsearch-cluster   service-broker   elasticsearch   1m
ha-postgresql-10-2          ha-postgresql           service-broker   postgresql      1m
memcached-1-5-4             default                 service-broker   memcached       1m
mongodb-3-6                 default                 service-broker   mongodb         1m
mongodb-cluster-3-6         mongodb-cluster         service-broker   mongodb         1m
mysql-8-0                   default                 service-broker   mysql           1m
postgresql-10-2             default                 service-broker   postgresql      1m
redis-4-0                   default                 service-broker   redis           1m
```

You can get the same thing in a different manner using Service Catalog CLI `svcat`.

```console
$ svcat get brokers
       NAME        NAMESPACE                           URL                           STATUS
+----------------+-----------+-----------------------------------------------------+--------+
  service-broker               http://service-broker.kube-system.svc.cluster.local   Ready

$ svcat get classes
      NAME        NAMESPACE                     DESCRIPTION
+---------------+-----------+-------------------------------------------------+
  elasticsearch               The example service from the ElasticSearch
                              database!
  memcached                   The example service from the Memcache database!
  mongodb                     The example service from the MongoDB database!
  mysql                       The example service from the MySQL database!
  postgresql                  The example service from the PostgreSQL
                              database!
  redis                       The example service from the Redis database!

$ svcat get plans
          NAME            NAMESPACE       CLASS                DESCRIPTION
+-----------------------+-----------+---------------+--------------------------------+
  default                             elasticsearch   The default plan for the
                                                      'elasticsearch' service
  elasticsearch-cluster               elasticsearch   This plan is for getting a
                                                      simple elasticsearch cluster
                                                      under the 'elasticsearch'
                                                      service
  default                             memcached       The default plan for the
                                                      'memcached' service
  default                             mongodb         The default plan for the
                                                      'mongodb' service
  mongodb-cluster                     mongodb         This plan is for getting a
                                                      simple mongodb cluster under
                                                      the 'mongodb' service
  default                             mysql           The default plan for the
                                                      'mysql' service
  ha-postgresql                       postgresql      This plan is for getting HA
                                                      postgres database under the
                                                      `postgresql` service
  default                             postgresql      This plan is for getting
                                                      standalone postgres database
                                                      under the `postgresql` service
  default                             redis           The default plan for the
                                                      'redis' service
```

Now, you are ready to use service broker.