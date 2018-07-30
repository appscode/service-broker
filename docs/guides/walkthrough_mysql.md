# Walkthrough MySQL

If we've `Kubedb Service Broker` installed, then we are ready for going forward. If not, then the [installation instructions](/docs/setup/install.md) are ready.

This document assumes that you've installed Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.27/docs/install.md). Optionally you may install the Service Catalog CLI, svcat. Examples for both svcat and kubectl are provided so that you may follow this walkthrough using svcat or using only kubectl.

> All commands in this document assume that you're operating out of the root of this repository.

## Check ClusterServiceClass and ClusterServicePlan for MySQL

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
$ kubectl get clusterserviceplans -o=custom-columns=NAME:.metadata.name,EXTERNAL\ NAME:.spec.externalName
NAME                EXTERNAL NAME
elasticsearch-5-6   default
memcached-1-5-4     default
mongodb-3-4         default
mysql-5-7           default
postgresql-9-6      default
redis-4-0           default

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

## Provisioning: Creating a New ServiceInstance

Since a `ClusterServiceClass` named `mysql` exists in the cluster with a `ClusterServicePlan` named `default`, we can create a `ServiceInstance` ponting to them.

> Unlike `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources, `ServiceInstance` resources must be namespaced. The latest version of service catelog supports `ServiceBroker`, `ServiceClass` and `ServicePlan` resources that are namespace scoped and alternative to `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources.

Create the `ServiceInstance`:

```console
$ kubectl create -f docs/examples/mysql-instance.yaml
serviceinstance.servicecatalog.k8s.io "my-broker-mysql-instance" created
```

After it is created, the service catalog controller will communicate with the `Kubedb Service Broker` server to initaiate provisioning. Now, see the details:

```console
$ svcat describe instance my-broker-mysql-instance --namespace service-broker
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
kubectl get serviceinstance my-broker-mysql-instance --namespace service-broker -o yaml
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

## Binding: Creating a ServiceBinding for this ServiceInstance

We've now a `ServiceInstance` ready. To use this we've to bind it. So, create a `ServiceBinding` resource:

```console
$ kubectl create -f docs/examples/mysql-binding.yaml
servicebinding.servicecatalog.k8s.io "my-broker-mysql-binding" created
```

Once the `ServiceBinding` resource is created, the service catalog controller initiate binding process by communicating with `Kubedb Service Broker` server. In general, this step makes the broker server to provide the necessary credentials. Then the service catalog controller will insert them into a Kubernetes `Secret` object.

```console
$ kubectl get servicebindings my-broker-mysql-binding --namespace service-broker -o=custom-columns=NAME:.metadata.name,INSTANCE\ REF:.spec.instanceRef.name,SECRET\ NAME:.spec.secretName
NAME                      INSTANCE REF               SECRET NAME
my-broker-mysql-binding   my-broker-mysql-instance   my-broker-mysql-secret

$ svcat get bindings --namespace service-broker
           NAME               NAMESPACE              INSTANCE           STATUS
+-------------------------+----------------+--------------------------+--------+
  my-broker-mysql-binding   service-broker   my-broker-mysql-instance   Ready

$ svcat describe bindings my-broker-mysql-binding --namespace service-broker
  Name:        my-broker-mysql-binding
  Namespace:   service-broker
  Status:      Ready - Injected bind result @ 2018-07-27 06:19:23 +0000 UTC  
  Secret:      my-broker-mysql-secret
  Instance:    my-broker-mysql-instance

Parameters:
  No parameters defined

Secret Data:
  Protocol   2 bytes
  host       49 bytes  
  password   16 bytes  
  port       4 bytes
  uri        81 bytes  
  user       4 bytes
  username   4 bytes
```

You can see the secret data by passing `--show-secrets` flag to the above command. The yaml configuration of this `ServiceBinding` resource is as follows:

```console
kubectl get servicebindings my-broker-mysql-binding --namespace service-broker -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  creationTimestamp: 2018-07-27T06:19:23Z
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: service-broker
  name: my-broker-mysql-binding
  namespace: service-broker
  resourceVersion: "46"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/service-broker/servicebindings/my-broker-mysql-binding
  uid: 014acaa7-9165-11e8-996f-0242ac110007
spec:
  externalID: 014aca51-9165-11e8-996f-0242ac110007
  instanceRef:
    name: my-broker-mysql-instance
  secretName: my-broker-mysql-secret
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: 2018-07-27T06:19:23Z
    message: Injected bind result
    reason: InjectedBindResult
    status: "True"
    type: Ready
  externalProperties: {}
  orphanMitigationInProgress: false
  reconciledGeneration: 1
  unbindStatus: Required
```

Here, the status has `Ready` condition which means the binding is now ready for use. This binding operation create a `Secret` named `my-broker-mysql-secret` in namespace `service-broker`.

```console
$ kubectl get secrets --namespace service-broker
NAME                         TYPE                                  DATA      AGE
default-token-4hlpm          kubernetes.io/service-account-token   3         3m
my-broker-mysql-secret       Opaque                                7         14s
mysql-5-7-rea3nf-auth        Opaque                                2         45s
service-broker-token-zk868   kubernetes.io/service-account-token   3         2m
```

## Unbinding: Deleting the ServiceBinding

We can now delete the `ServiceBinding` resource we created in the `Binding` step (it is called `Unbinding` the `ServiceInstance`)

```console
$ kubectl delete servicebinding my-broker-mysql-binding --namespace service-broker
servicebinding.servicecatalog.k8s.io "my-broker-mysql-binding" deleted

$ svcat unbind my-broker-mysql-instance --namespace service-broker
deleted my-broker-mysql-binding
```

After completion of unbinding, the `Secret` named `my-broker-mysql-secret` should be deleted.

```console
$ kubectl get secrets --namespace service-broker
NAME                         TYPE                                  DATA      AGE
default-token-4hlpm          kubernetes.io/service-account-token   3         4m
mysql-5-7-rea3nf-auth        Opaque                                2         2m
service-broker-token-zk868   kubernetes.io/service-account-token   3         4m
```

## Deprovisioning: Deleting the ServiceInstance

After unbinding the `ServiceInstance`, our next step is deleting the `ServiceInstance` resource we created before at the step of provisioning. It is called `Deprovisioning`.

```console
$ kubectl delete serviceinstance my-broker-mysql-instance --namespace service-broker
serviceinstance.servicecatalog.k8s.io "my-broker-mysql-instance" deleted

$ svcat deprovision my-broker-mysql-instance --namespace service-broker
deleted my-broker-mysql-instance
```

## Cleanup

Now, we've to clean the cluster. For this, just [uninstall](/docs/setup/uninstall.md) the broker. It'll delete the `ClusterServiceBroker` resource. Then service catalog controller automatically delete all `ClusterServiceClass` and `ClusterServicePlan` resources that came from that broker.

```console
$ kubectl get clusterserviceclasses
No resources found.

$ svcat get classes
  NAME   NAMESPACE   DESCRIPTION  
+------+-----------+-------------+
```