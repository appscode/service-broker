# Installation Guide

## Prerequisites

- This document assumes that you've installed Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.27/docs/install.md). Optionally you may install the Service Catalog CLI, svcat.
- You also need `Kubedb` to be installed. Please see the [installation instructions](https://kubedb.com/docs/0.8.0/setup/install).

> After satisfying the prerequisites, all commands in this document assume that you're operating out of the root of this repository.

## Install Service Broker

`Kubedb Service Broker` can be installed via a script or as a Helm chart.

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="true">Script</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="helm-tab" data-toggle="tab" href="#helm" role="tab" aria-controls="helm" aria-selected="false">Helm</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="script" role="tabpanel" aria-labelledby="script-tab">

### Using Script

To install `Kubedb Service Broker` in your Kubernetes cluster, run the following command:

```console
$ curl -fsSL https://raw.githubusercontent.com/kubedb/service-broker/master/hack/dev/build.sh | bash -s -- run
...
```

After successful installation, you should have a `service-broker-***` pod running in the `service-broker` namespace.

```console
$ kubectl get pods -n service-broker | grep service-broker
stash-operator-846d47f489-jrb58       1/1       Running   0          48s
```

#### Customizing Installer

The installer script and associated yaml files can be found in the [hack/dev](https://github.com/kubedb/service-broker/tree/master/hack/dev) folder. You can see the full list of flags available to installer using `-h` flag.

```console
$ curl -fsSL https://raw.githubusercontent.com/kubedb/service-broker/master/hack/dev/build.sh | bash -s -- -h
build.sh

build.sh [commands] [options]

commands:
---------
build            builds and push the docker image for service-broker
run              installs service-broker
uninstall        uninstalls service-broker

options:
--------
-h, --help                         show brief help
-n, --namespace=NAMESPACE          specify namespace (default: service-broker)
    --docker-registry              docker registry used to pull stash images (default: shudipta)
```

If you would like to run Service broker pod in `my-ns` namespace, pass the `--namespace=my-ns` flag:

```console
$ kubectl create namespace my-ns
namespace "my-ns" created

$ curl -fsSL https://raw.githubusercontent.com/kubedb/service-broker/master/hack/dev/build.sh | bash -s -- --namespace=my-ns
...
```

If you are using a private Docker registry, you need to pull the following image:

- [shudipta/db-broker](https://hub.docker.com/r/shudipta/db-broker/)

To pass the address of your private registry and optionally a image pull secret use flags `--docker-registry` and `--image-pull-secret` respectively.

```console
$ curl -fsSL https://raw.githubusercontent.com/kubedb/service-broker/master/hack/dev/build.sh \
    | bash -s -- --docker-registry=MY_REGISTRY [--image-pull-secret=SECRET_NAME]
...
```

### Verify installation

To check whether Service broker pod has started or not, run the following command:

```console
$ kubectl get pods --all-namespaces -l app=service-broker --watch
NAMESPACE        NAME                              READY     STATUS    RESTARTS   AGE
service-broker   service-broker-6974dcff7f-87cgm   0/1       Pending   0          0s
service-broker   service-broker-6974dcff7f-87cgm   0/1       Pending   0         1s
service-broker   service-broker-6974dcff7f-87cgm   0/1       ContainerCreating   0         2s
service-broker   service-broker-6974dcff7f-87cgm   1/1       Running   0         26s
```

Once the pods is running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm ClusterServiceBroker, ClusterServiceClass, ClusterServicePlan have been registered by the `Kubedb Service Broker`, run the following command:

```console
$ kubectl get clusterservicebrokers -l app=service-broker
NAME             AGE
service-broker   2m

$ kubectl get clusterserviceclass
NAME            AGE
elasticsearch   2m
memcached       2m
mongodb         2m
mysql           2m
postgresql      2m
redis           2m

$ kubectl get clusterserviceplans
NAME                AGE
elasticsearch-5-6   3m
memcached-1-5-4     3m
mongodb-3-4         3m
mysql-5-7           3m
postgresql-9-6      3m
redis-4-0           3m
```

You can get the same thing in a different manner using Service Catalog CLI `svcat`.

```console
$ svcat get brokers
       NAME                                 URL                             STATUS  
+----------------+--------------------------------------------------------+--------+
  service-broker   http://service-broker.service-broker.svc.cluster.local   Ready

$ svcat get classes
      NAME        NAMESPACE            DESCRIPTION
+---------------+-----------+--------------------------------+
  elasticsearch               The example service from the
                              ElasticSearch database
  memcached                   The example service from the
                              Memcache database
  mongodb                     The example service from the
                              MongoDB database
  mysql                       The example service from the
                              MySQL database
  postgresql                  The example service from the
                              PostgreSQL database
  redis                       The example service from the
                              Redis database

$ svcat get plans
   NAME         CLASS                DESCRIPTION
+---------+---------------+--------------------------------+
  default   elasticsearch   The default plan for the
                            'elasticsearch' service
  default   memcached       The default plan for the
                            'memcached' service
  default   mongodb         The default plan for the
                            'mongodb' service
  default   mysql           The default plan for the
                            'mysql' service
  default   postgresql      The default plan for the
                            'postgresql' service
  default   redis           The default plan for the
                            'redis' service
```

Now, you are ready to use Service broker.