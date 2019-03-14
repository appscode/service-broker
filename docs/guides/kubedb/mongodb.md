---
title: MongoDB | AppsCode Service Broker
menu:
  product_service-broker_0.2.0:
    identifier: mongodb-kubedb
    name: Mongodb
    parent: kubedb-guides
    weight: 30
product_name: service-broker
menu_name: product_service-broker_0.2.0
section_menu_id: guides
---
> New to AppsCode Service Broker? Please start [here](/docs/concepts/README.md).

# MongoDB Walk-through

This tutorial will show you how to use AppsCode Service Broker to provision and deprovision an MongoDB cluster and bind to the MongoDB service.

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Then install Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.38/docs/install.md) for Service Catalog. Optionally you may install the Service Catalog CLI, `svcat`. Examples for both `svcat` and `kubectl` are provided so that you may follow this walk-through using `svcat` or using only `kubectl`.

If you've AppsCode Service Broker installed, then we are ready for the next step. If not, follow the [instructions](/docs/setup/install.md) to install KubeDB and AppsCode Service Broker.

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> All commands in this document assume that you're operating out of the root of this repository.

## Check ClusterServiceClass and ClusterServicePlan for Mongodb

First, list the available `ClusterServiceClass` resources:

```console
$ kubectl get clusterserviceclasses
NAME                                   EXTERNAL-NAME   BROKER                    AGE
2010d83f-d908-4d9f-879c-ce8f5f527f2a   postgresql      appscode-service-broker   2h
315fc21c-829e-4aa1-8c16-f7921c33550d   elasticsearch   appscode-service-broker   2h
938a70c5-f2bc-4658-82dd-566bed7797e9   mysql           appscode-service-broker   2h
ccfd1c81-e59f-4875-a39f-75ba55320ce0   redis           appscode-service-broker   2h
d690058d-666c-45d8-ba98-fcb9fb47742e   mongodb         appscode-service-broker   2h
d88856cb-fe3f-4473-ba8b-641480da810f   memcached       appscode-service-broker   2h

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

Now, describe the `mongodb` class from the `Service Broker`.

```console
$ svcat describe class mongodb
  Name:              mongodb
  Scope:             cluster
  Description:       KubeDB managed MongoDB
  Kubernetes Name:   d690058d-666c-45d8-ba98-fcb9fb47742e
  Status:            Active
  Tags:
  Broker:            appscode-service-broker

Plans:
          NAME                    DESCRIPTION
+----------------------+--------------------------------+
  demo-mongodb           Demo Standalone MongoDB
                         database
  demo-mongodb-cluster   Demo MongoDB cluster
  mongodb                MongoDB database with custom
                         specification
```

To view the details of any plan in this class use command `$ svcat describe plan <class_name>/<plan_name>`. For example:

```console
$ svcat describe plan mongodb/mongodb --scope cluster
  Name:              mongodb
  Description:       MongoDB database with custom specification
  Kubernetes Name:   e8f87ba6-0711-42db-a663-a3c75b78a541
  Status:            Active
  Free:              true
  Class:             mongodb

Instances:
No instances defined
```

> Here we,ve used `--scope` flag to specify that our `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServiceBroker` resources are cluster scoped (not namespaced scope)

## Provisioning: Creating a New ServiceInstance

AppsCode Service Broker currently supports three plans for `mongodb` class as we can see above. Using `demo-mongodb` plan we can provision a demo MongoDB database. Using `demo-mongodb-cluster` plan we can provision a demo MongoDB database with clustering support. And using `mongodb` plan we can provision a custom MongoDB database with the full functionality of a [MongoDB CRD](https://kubedb.com/docs/0.10.0/concepts/databases/mongodb).

AppsCode Service Broker accepts only metadata and [MongoDB Spec](https://kubedb.com/docs/0.10.0/concepts/databases/mongodb/#mongodb-spec) as parameters for the plans of `mongodb` class. The metadata and spec should be provided with key `"metadata"` and `"spec"` respectively. The metadata is optional for all of the plans available. But the spec is required for the custom plan and it must be valid.

Since a `ClusterServiceClass` named `mongodb` exists in the cluster with a `ClusterServicePlan` named `mongodb`, we can create a `ServiceInstance` pointing to them with custom specification as parameters.

> Unlike `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources, `ServiceInstance` resources must be namespaced. The latest version of .service catalog supports `ServiceBroker`, `ServiceClass` and `ServicePlan` resources that are namespace scoped and alternative to `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources.

Create the `ServiceInstance`:

```console
$ kubectl create -f docs/examples/mongodb-instance.yaml
serviceinstance.servicecatalog.k8s.io/mongodb created
```

After it is created, the service catalog controller will communicate with the service broker server to initiate provisioning. Now, see the details:

```console
$ svcat describe instance mongodb --namespace demo
  Name:        mongodb
  Namespace:   demo
  Status:      Ready - The instance was provisioned successfully @ 2018-12-26 09:10:34 +0000 UTC
  Class:       mongodb
  Plan:        mongodb

Parameters:
  metadata:
    labels:
      app: my-mongodb
  spec:
    storage:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
      storageClassName: standard
    storageType: Durable
    terminationPolicy: WipeOut
    version: 3.4-v1

Bindings:
No bindings defined
```

The yaml configuration of this `ServiceInstance`:

```console
kubectl get serviceinstance mongodb --namespace demo -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  creationTimestamp: "2018-12-26T09:10:33Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
  name: mongodb
  namespace: demo
  resourceVersion: "157"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/demo/serviceinstances/mongodb
  uid: 1989092b-08ee-11e9-9fa4-0242ac110006
spec:
  clusterServiceClassExternalName: mongodb
  clusterServiceClassRef:
    name: d690058d-666c-45d8-ba98-fcb9fb47742e
  clusterServicePlanExternalName: mongodb
  clusterServicePlanRef:
    name: e8f87ba6-0711-42db-a663-a3c75b78a541
  externalID: 198908d8-08ee-11e9-9fa4-0242ac110006
  parameters:
    metadata:
      labels:
        app: my-mongodb
    spec:
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      storageType: Durable
      terminationPolicy: WipeOut
      version: 3.4-v1
  updateRequests: 0
  userInfo:
    groups:
    - system:masters
    - system:authenticated
    uid: ""
    username: minikube-user
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: "2018-12-26T09:10:34Z"
    message: The instance was provisioned successfully
    reason: ProvisionedSuccessfully
    status: "True"
    type: Ready
  deprovisionStatus: Required
  externalProperties:
    clusterServicePlanExternalID: e8f87ba6-0711-42db-a663-a3c75b78a541
    clusterServicePlanExternalName: mongodb
    parameterChecksum: 28e8a2c60d61c5feed7b353472b4a0db00bf9868dc946a9ae113a041fe7e8ea4
    parameters:
      metadata:
        labels:
          app: my-mongodb
      spec:
        storage:
          accessModes:
          - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
          storageClassName: standard
        storageType: Durable
        terminationPolicy: WipeOut
        version: 3.4-v1
    userInfo:
      groups:
      - system:masters
      - system:authenticated
      uid: ""
      username: minikube-user
  observedGeneration: 1
  orphanMitigationInProgress: false
  provisionStatus: Provisioned
  reconciledGeneration: 1
```

## Binding: Creating a ServiceBinding for this ServiceInstance

We have a ready `ServiceInstance`. To use this service, we can bind to it. AppsCode Service Broker currently supports no parameter for binding. Now, create a `ServiceBinding` resource:

```console
$ kubectl create -f docs/examples/mongodb-binding.yaml
servicebinding.servicecatalog.k8s.io/mongodb created
```

Once the `ServiceBinding` resource is created, the service catalog controller initiates binding process by communicating with the service broker server. In general, the broker server returns the necessary credentials in this step. Then the service catalog controller will insert them into a Kubernetes `Secret` object.

```console
$ kubectl get servicebindings mongodb --namespace demo
NAME      SERVICE-INSTANCE   SECRET-NAME   STATUS   AGE
mongodb   mongodb            mongodb       Ready    2m

$ svcat get bindings mongodb --namespace demo
   NAME     NAMESPACE   INSTANCE   STATUS
+---------+-----------+----------+--------+
  mongodb   demo        mongodb    Ready

$ svcat describe bindings mongodb --namespace demo
  Name:        mongodb
  Namespace:   demo
  Status:      Ready - Injected bind result @ 2018-12-26 09:15:42 +0000 UTC
  Secret:      mongodb
  Instance:    mongodb

Parameters:
  No parameters defined

Secret Data:
  Host       16 bytes
  Password   16 bytes
  Port       5 bytes
  Protocol   7 bytes
  RootCert   4 bytes
  URI        32 bytes
  Username   4 bytes
```

You can see the secret data by passing `--show-secrets` flag to the above command. The yaml configuration of this `ServiceBinding` resource is as follows:

```console
kubectl get servicebindings mongodb --namespace demo -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  creationTimestamp: "2018-12-26T09:15:42Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
  name: mongodb
  namespace: demo
  resourceVersion: "161"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/demo/servicebindings/mongodb
  uid: d18d903d-08ee-11e9-9fa4-0242ac110006
spec:
  externalID: d18d8fb6-08ee-11e9-9fa4-0242ac110006
  instanceRef:
    name: mongodb
  secretName: mongodb
  userInfo:
    groups:
    - system:masters
    - system:authenticated
    uid: ""
    username: minikube-user
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: "2018-12-26T09:15:42Z"
    message: Injected bind result
    reason: InjectedBindResult
    status: "True"
    type: Ready
  externalProperties:
    userInfo:
      groups:
      - system:masters
      - system:authenticated
      uid: ""
      username: minikube-user
  orphanMitigationInProgress: false
  reconciledGeneration: 1
  unbindStatus: Required
```

Here, the status has `Ready` condition which means the binding is now ready for use. This binding operation creates a `Secret` named `mongodb` in namespace `demo`.

```console
$ kubectl get secrets --namespace demo
NAME                  TYPE                                  DATA   AGE
default-token-2zx6l   kubernetes.io/service-account-token   3      3h52m
mongodb               Opaque                                7      8m51s
mongodb-auth          Opaque                                2      13m
```

## Unbinding: Deleting the ServiceBinding

We can now delete the `ServiceBinding` resource we created in the `Binding` step (it is called `Unbinding` the `ServiceInstance`)

```console
$ kubectl delete servicebinding mongodb --namespace demo
servicebinding.servicecatalog.k8s.io "mongodb" deleted

$ svcat unbind mongodb --namespace demo
deleted mongodb
```

After completion of unbinding, the `Secret` named `mongodb` should be deleted.

```console
$ kubectl get secrets --namespace demo
NAME                  TYPE                                  DATA   AGE
default-token-2zx6l   kubernetes.io/service-account-token   3      3h54m
mongodb-auth          Opaque                                2      15m
```

## Deprovisioning: Deleting the ServiceInstance

After unbinding the `ServiceInstance`, our next step is deleting the `ServiceInstance` resource we provisioned before. It is called `Deprovisioning`.

```console
$ kubectl delete serviceinstance mongodb --namespace demo
serviceinstance.servicecatalog.k8s.io "mongodb" deleted

$ svcat deprovision mongodb --namespace demo
deleted mongodb
```

## Cleanup

To cleanup the cluster, just [uninstall](/docs/setup/uninstall.md) the broker. It'll delete the `ClusterServiceBroker` resource. Then service catalog controller automatically deletes all `ClusterServiceClass` and `ClusterServicePlan` resources that came from that broker.

```console
$ kubectl get clusterserviceclasses
No resources found.

$ svcat get classes
  NAME   NAMESPACE   DESCRIPTION
+------+-----------+-------------+
```