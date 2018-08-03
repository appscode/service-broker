# Walkthrough Memcached

If we've `Kubedb Service Broker` installed, then we are ready for going forward. If not, then the [installation instructions](/docs/setup/install.md) are ready.

This document assumes that you've installed Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.27/docs/install.md). Optionally you may install the Service Catalog CLI, svcat. Examples for both svcat and kubectl are provided so that you may follow this walkthrough using svcat or using only kubectl.

> All commands in this document assume that you're operating out of the root of this repository.

## Check ClusterServiceClass and ClusterServicePlan for Memcached

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
                              ElasticSearch database!
  memcached                   The example service from the
                              Memcache database!
  mongodb                     The example service from the
                              MongoDB database!
  mysql                       The example service from the
                              MySQL database!
  postgresql                  The example service from the
                              PostgreSQL database!
  redis                       The example service from the
                              Redis database!
```

> **NOTE:** The above kubectl command uses a custom set of columns. The **`NAME`** field is the Kubernetes name of the `ClusterServiceClass` and the **`EXTERNAL NAME`** field is the human-readable name for the service that the broker returns.

Now, describe the `ClusterServiceClass` named `memcached` from `Kubedb Service Broker`.

```console
$ svcat describe class memcached
  Name:          memcached
  Description:   The example service from the Memcache database!  
  UUID:          memcached
  Status:        Active
  Tags:
  Broker:        service-broker

Plans:
   NAME              DESCRIPTION
+---------+--------------------------------+
  default   The default plan for the
            'memcached' service
```

To view the details of the `default` plan of `memcached` class:

```console
$ kubectl get clusterserviceplans -o=custom-columns=NAME:.metadata.name,EXTERNAL\ NAME:.spec.externalName
NAME                EXTERNAL NAME
elasticsearch-5-6   default
memcached-1-5-4     default
mongodb-3-4         default
mysql-5-7           default
postgresql-9-6      default
memcached-4-0           default

$ svcat get plan memcached/default
   NAME       CLASS              DESCRIPTION
+---------+-----------+--------------------------------+
  default   memcached   The default plan for the
                        'memcached' service

$ svcat describe plan memcached/default
  Name:          default
  Description:   The default plan for the 'memcached' service  
  UUID:          memcached-1-5-4
  Status:        Active
  Free:          true
  Class:         memcached

Instances:
No instances defined
```

## Provisioning: Creating a New ServiceInstance

Since a `ClusterServiceClass` named `memcached` exists in the cluster with a `ClusterServicePlan` named `default`, we can create a `ServiceInstance` ponting to them.

> Unlike `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources, `ServiceInstance` resources must be namespaced. The latest version of service catelog supports `ServiceBroker`, `ServiceClass` and `ServicePlan` resources that are namespace scoped and alternative to `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources.

Create the `ServiceInstance`:

```console
$ kubectl create -f docs/examples/memcached-instance.yaml
serviceinstance.servicecatalog.k8s.io "my-broker-memcached-instance" created
```

After it is created, the service catalog controller will communicate with the `Kubedb Service Broker` server to initaiate provisioning. Now, see the details:

```console
$ svcat describe instance my-broker-memcached-instance --namespace service-broker
  Name:        my-broker-memcached-instance
  Namespace:   service-broker
  Status:      Ready - The instance was provisioned successfully @ 2018-07-30 08:57:04 +0000 UTC  
  Class:       memcached
  Plan:        default

Parameters:
  No parameters defined

Bindings:
No bindings defined
```

The yaml configuration of this `ServiceInstance`:

```console
kubectl get serviceinstance my-broker-memcached-instance --namespace service-broker -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  creationTimestamp: 2018-07-30T08:57:03Z
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: service-broker
  name: my-broker-memcached-instance
  namespace: service-broker
  resourceVersion: "201"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/service-broker/serviceinstances/my-broker-memcached-instance
  uid: 87737598-93d6-11e8-bf8e-0242ac110008
spec:
  clusterServiceClassExternalName: memcached
  clusterServiceClassRef:
    name: memcached
  clusterServicePlanExternalName: default
  clusterServicePlanRef:
    name: memcached-1-5-4
  externalID: 87737507-93d6-11e8-bf8e-0242ac110008
  updateRequests: 0
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: 2018-07-30T08:57:04Z
    message: The instance was provisioned successfully
    reason: ProvisionedSuccessfully
    status: "True"
    type: Ready
  deprovisionStatus: Required
  externalProperties:
    clusterServicePlanExternalID: memcached-1-5-4
    clusterServicePlanExternalName: default
  observedGeneration: 1
  orphanMitigationInProgress: false
  provisionStatus: Provisioned
  reconciledGeneration: 1
```

## Binding: Creating a ServiceBinding for this ServiceInstance

We've now a `ServiceInstance` ready. To use this we've to bind it. So, create a `ServiceBinding` resource:

```console
$ kubectl create -f docs/examples/memcached-binding.yaml
servicebinding.servicecatalog.k8s.io "my-broker-memcached-binding" created
```

Once the `ServiceBinding` resource is created, the service catalog controller initiate binding process by communicating with `Kubedb Service Broker` server. In general, this step makes the broker server to provide the necessary credentials. Then the service catalog controller will insert them into a Kubernetes `Secret` object.

```console
$ kubectl get servicebindings my-broker-memcached-binding --namespace service-broker -o=custom-columns=NAME:.metadata.name,INSTANCE\ REF:.spec.instanceRef.name,SECRET\ NAME:.spec.secretName
NAME                          INSTANCE REF                   SECRET NAME
my-broker-memcached-binding   my-broker-memcached-instance   my-broker-memcached-secret

$ svcat get bindings --namespace service-broker
             NAME                 NAMESPACE                INSTANCE             STATUS  
+-----------------------------+----------------+------------------------------+--------+
  my-broker-memcached-binding   service-broker   my-broker-memcached-instance   Ready

$ svcat describe bindings my-broker-memcached-binding --namespace service-broker
  Name:        my-broker-memcached-binding
  Namespace:   service-broker
  Status:      Ready - Injected bind result @ 2018-07-30 08:58:40 +0000 UTC  
  Secret:      my-broker-memcached-secret
  Instance:    my-broker-memcached-instance

Parameters:
  No parameters defined

Secret Data:
  Protocol   2 bytes
  host       55 bytes  
  port       5 bytes
  uri        66 bytes  
```

You can see the secret data by passing `--show-secrets` flag to the above command. The yaml configuration of this `ServiceBinding` resource is as follows:

```console
kubectl get servicebindings my-broker-memcached-binding --namespace service-broker -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  creationTimestamp: 2018-07-30T08:58:39Z
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: service-broker
  name: my-broker-memcached-binding
  namespace: service-broker
  resourceVersion: "204"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/service-broker/servicebindings/my-broker-memcached-binding
  uid: c0938699-93d6-11e8-bf8e-0242ac110008
spec:
  externalID: c093865d-93d6-11e8-bf8e-0242ac110008
  instanceRef:
    name: my-broker-memcached-instance
  secretName: my-broker-memcached-secret
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: 2018-07-30T08:58:40Z
    message: Injected bind result
    reason: InjectedBindResult
    status: "True"
    type: Ready
  externalProperties: {}
  orphanMitigationInProgress: false
  reconciledGeneration: 1
  unbindStatus: Required
```

Here, the status has `Ready` condition which means the binding is now ready for use. This binding operation create a `Secret` named `my-broker-memcached-secret` in namespace `service-broker`.

```console
$ kubectl get secrets --namespace service-broker
NAME                         TYPE                                  DATA      AGE
default-token-c5qbd          kubernetes.io/service-account-token   3         4h
my-broker-memcached-secret   Opaque                                4         3m
service-broker-token-m6hsm   kubernetes.io/service-account-token   3         4h
```

## Unbinding: Deleting the ServiceBinding

We can now delete the `ServiceBinding` resource we created in the `Binding` step (it is called `Unbinding` the `ServiceInstance`)

```console
$ kubectl delete servicebinding my-broker-memcached-binding --namespace service-broker
servicebinding.servicecatalog.k8s.io "my-broker-memcached-binding" deleted

$ svcat unbind my-broker-memcached-instance --namespace service-broker
deleted my-broker-memcached-binding
```

After completion of unbinding, the `Secret` named `my-broker-memcached-secret` should be deleted.

```console
$ kubectl get secrets --namespace service-broker
NAME                         TYPE                                  DATA      AGE
default-token-c5qbd          kubernetes.io/service-account-token   3         4h
service-broker-token-m6hsm   kubernetes.io/service-account-token   3         4h
```

## Deprovisioning: Deleting the ServiceInstance

After unbinding the `ServiceInstance`, our next step is deleting the `ServiceInstance` resource we created before at the step of provisioning. It is called `Deprovisioning`.

```console
$ kubectl delete serviceinstance my-broker-memcached-instance --namespace service-broker
serviceinstance.servicecatalog.k8s.io "my-broker-memcached-instance" deleted

$ svcat deprovision my-broker-memcached-instance --namespace service-broker
deleted my-broker-memcached-instance
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