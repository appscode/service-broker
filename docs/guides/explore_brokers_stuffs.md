# Explore through Broker Stuffs

This document assumes that you've installed Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.27/docs/install.md). Optionally you may install the Service Catalog CLI, svcat. Examples for both `svcat` and `kubectl` are provided so that you may follow this walkthrough using `svcat` or using only `kubectl`.

## Checking a ClusterServiceBroker Resource

We've created a `ClusterServiceBroker` resource in the service-catalog API server, querying service catalog returns with this resource:

```console
$ kubectl get clusterservicebrokers -l app=service-broker
NAME             URL                                                      STATUS   AGE
service-broker   http://service-broker.service-broker.svc.cluster.local   Ready    41s

$ svcat get brokers
       NAME        NAMESPACE                            URL                             STATUS
+----------------+-----------+--------------------------------------------------------+--------+
  service-broker               http://service-broker.service-broker.svc.cluster.local   Ready
```

After creating `ClusterServiceBroker` resource, the service catalog controller responds by querying the broker server to see what services it offers and creates a `ClusterServiceClass` for each.

We can check the status of the broker:

```console
$ svcat describe broker service-broker
  Name:     service-broker
  URL:      http://service-broker.service-broker.svc.cluster.local
  Status:   Ready - Successfully fetched catalog entries from broker @ 2018-12-03 05:04:37 +0000 UTC
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
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"servicecatalog.k8s.io/v1beta1","kind":"ClusterServiceBroker","metadata":{"annotations":{},"labels":{"app":"service-broker"},"name":"service-broker"},"spec":{"url":"http://service-broker.service-broker.svc.cluster.local"}}
  creationTimestamp: 2018-12-03T05:04:05Z
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: service-broker
  name: service-broker
  resourceVersion: "963"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/clusterservicebrokers/service-broker
  uid: dbab1ac4-f6b8-11e8-89f4-0242ac110003
spec:
  relistBehavior: Duration
  relistRequests: 0
  url: http://service-broker.service-broker.svc.cluster.local
status:
  conditions:
  - lastTransitionTime: 2018-12-03T05:04:37Z
    message: Successfully fetched catalog entries from broker.
    reason: FetchedCatalog
    status: "True"
    type: Ready
  lastCatalogRetrievalTime: 2018-12-03T05:04:37Z
  reconciledGeneration: 1
```

> Notice that the status says that the broker's catalog of service offerings has been successfully added to our cluster's service catalog.

## Viewing ClusterServiceClasses

There is a `ClusterServiceClass` for each service that the AppsCode Service Broker provides. To view these `ClusterServiceClass` resources:

```console
$ kubectl get clusterserviceclasses -o=custom-columns=NAME:.metadata.name,EXTERNAL\ NAME:.spec.externalName
NAME            EXTERNAL NAME
elasticsearch   elasticsearch
memcached       memcached
mongodb         mongodb
mysql           mysql
postgresql      postgresql
redis           redis

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
```

> **NOTE:** The above kubectl command uses a custom set of columns. The **`NAME`** field is the Kubernetes name of the `ClusterServiceClass` and the **`EXTERNAL NAME`** field is the human-readable name for the service that the broker returns.

Here is the details for service `mysql` from KubeDB by `AppsCode`.

```console
$ svcat describe class mysql
  Name:              mysql
  Scope:             cluster
  Description:       The example service from the MySQL database!
  Kubernetes Name:   mysql
  Status:            Active
  Tags:
  Broker:            service-broker

Plans:
   NAME              DESCRIPTION
+---------+--------------------------------+
  default   The default plan for the
            'mysql' service
```

Here is the yaml configuration of `ClusterServiceClass` named `mysql`.

```console
kubectl get clusterserviceclasses mysql -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServiceClass
metadata:
  creationTimestamp: 2018-12-03T05:04:35Z
  name: mysql
  ownerReferences:
  - apiVersion: servicecatalog.k8s.io/v1beta1
    blockOwnerDeletion: false
    controller: true
    kind: ClusterServiceBroker
    name: service-broker
    uid: dbab1ac4-f6b8-11e8-89f4-0242ac110003
  resourceVersion: "948"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/clusterserviceclasses/mysql
  uid: edbd5a70-f6b8-11e8-89f4-0242ac110003
spec:
  bindable: true
  bindingRetrievable: false
  clusterServiceBrokerName: service-broker
  description: The example service from the MySQL database!
  externalID: mysql
  externalMetadata:
    displayName: Example MySQL DB service
    imageUrl: http://www.cgtechworld.in/images/Training/technologies/mysql.png
  externalName: mysql
  planUpdatable: true
status:
  removedFromBrokerCatalog: false
```

## Viewing ClusterServicePlans

There is also a `ClusterServicePlan` for each plan under the broker's services. To view these `ClusterServicePlan` resources:

```console
$ kubectl get clusterserviceplans -o=custom-columns=NAME:.metadata.name,EXTERNAL\ NAME:.spec.externalName
NAME                        EXTERNAL NAME
elasticsearch-6-3           default
elasticsearch-cluster-6-3   elasticsearch-cluster
ha-postgresql-10-2          ha-postgresql
memcached-1-5-4             default
mongodb-3-6                 default
mongodb-cluster-3-6         mongodb-cluster
mysql-8-0                   default
postgresql-10-2             default
redis-4-0                   default

$ svcat get plans
NAME                        EXTERNAL NAME
elasticsearch-6-3           default
elasticsearch-cluster-6-3   elasticsearch-cluster
ha-postgresql-10-2          ha-postgresql
memcached-1-5-4             default
mongodb-3-6                 default
mongodb-cluster-3-6         mongodb-cluster
mysql-8-0                   default
postgresql-10-2             default
redis-4-0                   default
```

As an example, to view the details of the `default` plan of `mysql` class:

```console
$ svcat describe plan mysql/default --scope cluster
  Name:              default
  Description:       The default plan for the 'mysql' service
  Kubernetes Name:   mysql-8-0
  Status:            Active
  Free:              true
  Class:             mysql

Instances:
No instances defined
```

> Here we,ve used `--scope` flag to specify that our `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServiceBroker` resources are cluster scoped (not namespaced scope)

Here is the yaml configuration of `ClusterServicePlan` named `default` of `ClusterServiceClass` named `mysql`.

```console
kubectl get clusterserviceplans mysql-8-0 -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ClusterServicePlan
metadata:
  creationTimestamp: 2018-12-03T05:04:35Z
  name: mysql-8-0
  ownerReferences:
  - apiVersion: servicecatalog.k8s.io/v1beta1
    blockOwnerDeletion: false
    controller: true
    kind: ClusterServiceBroker
    name: service-broker
    uid: dbab1ac4-f6b8-11e8-89f4-0242ac110003
  resourceVersion: "954"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/clusterserviceplans/mysql-8-0
  uid: edd9682e-f6b8-11e8-89f4-0242ac110003
spec:
  clusterServiceBrokerName: service-broker
  clusterServiceClassRef:
    name: mysql
  description: The default plan for the 'mysql' service
  externalID: mysql-8-0
  externalName: default
  free: true
status:
  removedFromBrokerCatalog: false
```