# Walkthrough MySQL

This document assumes that you've installed Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.27/docs/install.md). Optionally you may install the Service Catalog CLI, svcat. Examples for both svcat and kubectl are provided so that you may follow this walkthrough using svcat or using only kubectl.

> All commands in this document assume that you're operating out of the root of this repository.

## Check MySQL ClusterServiceClass and It's ClusterServicePlan

First, list the available `ClusterServiceClass` resources:

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
                              Redis database!
```

> **NOTE:** The above kubectl command uses a custom set of columns. The **`NAME`** field is the Kubernetes name of the `ClusterServiceClass` and the **`EXTERNAL NAME`** field is the human-readable name for the service that the broker returns.

Now, describe the `ClusterServiceClass` `mysql` from `Kubedb Service Broker`.

```console
$ svcat describe class mysql
  Name:          mysql
  Description:   The example service from the MySQL database!
  UUID:          mysql
  Status:        Active
  Tags:
  Broker:        service-broker

Plans:
   NAME              DESCRIPTION
+---------+--------------------------------+
  default   The default plan for the
            'mysql' service
```

To view the details of the `default` plan of `mysql` class:

```console
$ kubectl get clusterserviceplans mysql-5-7
NAME        AGE
mysql-5-7   3h

$ svcat get plan mysql/default
   NAME     CLASS            DESCRIPTION
+---------+-------+--------------------------------+
  default   mysql   The default plan for the
                    'mysql' service

$ svcat describe plan mysql/default
  Name:          default
  Description:   The default plan for the 'mysql' service  
  UUID:          mysql-5-7
  Status:        Active
  Free:          true
  Class:         mysql

Instances:
No instances defined
```

## Creating a New ServiceInstance

Since a `ClusterServiceClass` named `mysql` exists in the cluster with a `ClusterServicePlan` named `default`, we can create a `ServiceInstance` ponting to them.

> Unlike `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources, `ServiceInstance` resources must be namespaced. The latest version of service catelog supports `ServiceBroker`, `ServiceClass` and `ServicePlan` resources that are namespace scoped and alternative to `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources.

Create the `ServiceInstance`:

```console
$ kubectl create -f docs/examples/mysql-instance.yaml 
serviceinstance.servicecatalog.k8s.io "my-broker-mysql-instance" created
```

After it is created, the service catalog controller will communicate with the `Kubedb Service Broker` server to initaiate provisioning. Now, see the details:

```console
$ svcat describe instance my-broker-mysql-instance -n service-broker
  Name:        my-broker-mysql-instance
  Namespace:   service-broker
  Status:      Ready - The instance was provisioned successfully @ 2018-07-26 12:38:37 +0000 UTC  
  Class:       mysql
  Plan:        default

Parameters:
  No parameters defined

Bindings:
No bindings defined
```

The yaml configuration of this `ServiceInstance`:

```console
kubectl get serviceinstance my-broker-mysql-instance -n service-broker -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  creationTimestamp: 2018-07-26T12:38:36Z
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: service-broker
  name: my-broker-mysql-instance
  namespace: service-broker
  resourceVersion: "206"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/service-broker/serviceinstances/my-broker-mysql-instance
  uid: d0aadec5-90d0-11e8-9b93-0242ac110007
spec:
  clusterServiceClassExternalName: mysql
  clusterServiceClassRef:
    name: mysql
  clusterServicePlanExternalName: default
  clusterServicePlanRef:
    name: mysql-5-7
  externalID: d0aade36-90d0-11e8-9b93-0242ac110007
  updateRequests: 0
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: 2018-07-26T12:38:37Z
    message: The instance was provisioned successfully
    reason: ProvisionedSuccessfully
    status: "True"
    type: Ready
  deprovisionStatus: Required
  externalProperties:
    clusterServicePlanExternalID: mysql-5-7
    clusterServicePlanExternalName: default
  observedGeneration: 1
  orphanMitigationInProgress: false
  provisionStatus: Provisioned
  reconciledGeneration: 1
  ```