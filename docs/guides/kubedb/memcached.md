# Walkthrough Memcached

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

If we've AppsCode Service Broker installed, then we are ready for going forward. If not, then follow the [installation instructions](/docs/setup/install.md).

This document assumes that you've installed Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.27/docs/install.md) to install Service Catalog. Optionally you may install the Service Catalog CLI, `svcat`. Examples for both `svcat` and `kubectl` are provided so that you may follow this walkthrough using `svcat` or using only `kubectl`.

> All commands in this document assume that you're operating out of the root of this repository.

## Check ClusterServiceClass and ClusterServicePlan for Memcached

First, list the available `ClusterServiceClass` resources:

```console
$ kubectl get clusterserviceclasses
NAME                                   EXTERNAL-NAME   BROKER                    AGE
2010d83f-d908-4d9f-879c-ce8f5f527f2a   postgresql      appscode-service-broker   5m
315fc21c-829e-4aa1-8c16-f7921c33550d   elasticsearch   appscode-service-broker   5m
938a70c5-f2bc-4658-82dd-566bed7797e9   mysql           appscode-service-broker   5m
ccfd1c81-e59f-4875-a39f-75ba55320ce0   redis           appscode-service-broker   5m
d690058d-666c-45d8-ba98-fcb9fb47742e   mongodb         appscode-service-broker   5m
d88856cb-fe3f-4473-ba8b-641480da810f   memcached       appscode-service-broker   5m

$ svcat get classes
      NAME        NAMESPACE           DESCRIPTION
+---------------+-----------+------------------------------+
  postgresql                  KubeDB managed PostgreSQL
  elasticsearch               KubeDB managed ElasticSearch
  mysql                       KubeDB managed MySQL
  redis                       KubeDB managed Redis
  mongodb                     KubeDB managed MongoDB
  memcached                   KubeDB managed Memcache
```

Now, describe `memcached` class from the `Service Broker`.

```console
$ svcat describe class memcached
  Name:              memcached
  Scope:             cluster
  Description:       KubeDB managed Memcache
  Kubernetes Name:   d88856cb-fe3f-4473-ba8b-641480da810f
  Status:            Active
  Tags:
  Broker:            appscode-service-broker

Plans:
       NAME                 DESCRIPTION
+----------------+--------------------------------+
  demo-memcached   Demo Memcached
  memcached        Memcached with custom
                   specification
```

To view the details of any plan in this class use command `$ svcat describe plan <class_name>/<plan_name>`. For example:

```console
$ svcat describe plan memcached/memcached --scope cluster
  Name:              memcached
  Description:       Memcached with custom specification
  Kubernetes Name:   d40e49b2-f8fb-4d47-96d3-35089bd0942d
  Status:            Active
  Free:              true
  Class:             memcached

Instances:
No instances defined
```

> Here we,ve used `--scope` flag to specify that our `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServiceBroker` resources are cluster scoped (not namespaced scope)

## Provisioning: Creating a New ServiceInstance

AppsCode Service Broker currently supports two plans for `memcached` class as we can see above. Using `demo-memcached` plan we can provision a demo Memcached database in cluster. And using `memcached` plan we can provision a custom Memcached database with custom [Memcached Spec](https://kubedb.com/docs/0.9.0/concepts/databases/memcached/#memcached-spec) of [Memcached CRD](https://kubedb.com/docs/0.9.0/concepts/databases/memcached).

AppsCode Service Broker accept only metadata and [Memcached Spec](https://kubedb.com/docs/0.9.0/concepts/databases/memcached/#memcached-spec) as parameters for the plans of `memcached` class. The metadata and spec should be provided with key `"metadata"` and `"spec"` respectfully. The metadata is optional for both of the plans available. But the spec is required for the clustom plan and it must be valid.

Since a `ClusterServiceClass` named `memcached` exists in the cluster with a `ClusterServicePlan` named `memcached`, we can create a `ServiceInstance` ponting to them with custom specification as parameter.

> Unlike `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources, `ServiceInstance` resources must be namespaced. The latest version of service catelog supports `ServiceBroker`, `ServiceClass` and `ServicePlan` resources that are namespace scoped and alternative to `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources.

Create the `ServiceInstance`:

```console
$ kubectl create -f docs/examples/memcached-instance.yaml
serviceinstance.servicecatalog.k8s.io/memcached created
```

After it is created, the service catalog controller will communicate with the service broker server to initaiate provisioning. Now, see the details:

```console
$ svcat describe instance memcached --namespace demo
  Name:        memcached
  Namespace:   demo
  Status:      Ready - The instance was provisioned successfully @ 2018-12-26 05:38:02 +0000 UTC
  Class:       memcached
  Plan:        memcached

Parameters:
  metadata:
    labels:
      app: my-memcached
  spec:
    podTemplate:
      spec:
        resources:
          limits:
            cpu: 500m
            memory: 128Mi
          requests:
            cpu: 250m
            memory: 64Mi
    replicas: 3
    terminationPolicy: WipeOut
    version: 1.5.4-v1

Bindings:
No bindings defined
```

The yaml configuration of this `ServiceInstance`:

```console
kubectl get serviceinstance memcached --namespace demo -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  creationTimestamp: "2018-12-26T05:38:01Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
  name: memcached
  namespace: demo
  resourceVersion: "101"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/demo/serviceinstances/memcached
  uid: 6878eb72-08d0-11e9-9fa4-0242ac110006
spec:
  clusterServiceClassExternalName: memcached
  clusterServiceClassRef:
    name: d88856cb-fe3f-4473-ba8b-641480da810f
  clusterServicePlanExternalName: memcached
  clusterServicePlanRef:
    name: d40e49b2-f8fb-4d47-96d3-35089bd0942d
  externalID: 6878eb3d-08d0-11e9-9fa4-0242ac110006
  parameters:
    metadata:
      labels:
        app: my-memcached
    spec:
      podTemplate:
        spec:
          resources:
            limits:
              cpu: 500m
              memory: 128Mi
            requests:
              cpu: 250m
              memory: 64Mi
      replicas: 3
      terminationPolicy: WipeOut
      version: 1.5.4-v1
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
  - lastTransitionTime: "2018-12-26T05:38:02Z"
    message: The instance was provisioned successfully
    reason: ProvisionedSuccessfully
    status: "True"
    type: Ready
  deprovisionStatus: Required
  externalProperties:
    clusterServicePlanExternalID: d40e49b2-f8fb-4d47-96d3-35089bd0942d
    clusterServicePlanExternalName: memcached
    parameterChecksum: 9627fe20432ac96997f9ff1983cb3cb6e3b1a2d14184f2e44761a0eb13c31993
    parameters:
      metadata:
        labels:
          app: my-memcached
      spec:
        podTemplate:
          spec:
            resources:
              limits:
                cpu: 500m
                memory: 128Mi
              requests:
                cpu: 250m
                memory: 64Mi
        replicas: 3
        terminationPolicy: WipeOut
        version: 1.5.4-v1
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

We've now a `ServiceInstance` ready. To use this we've to bind it. AppsCode Service Broker currently supports no parameter for binding. So we didn't use any parameter for it. Now create a `ServiceBinding` resource:

```console
$ kubectl create -f docs/examples/memcached-binding.yaml
servicebinding.servicecatalog.k8s.io/memcached created
```

Once the `ServiceBinding` resource is created, the service catalog controller initiate binding process by communicating with the service broker server. In general, this step makes the broker server to provide the necessary credentials. Then the service catalog controller will insert them into a Kubernetes `Secret` object.

```console
$ kubectl get servicebindings memcached --namespace demo
NAME        SERVICE-INSTANCE   SECRET-NAME   STATUS   AGE
memcached   memcached          memcached     Ready    1m

$ svcat get bindings memcached --namespace demo
    NAME      NAMESPACE   INSTANCE    STATUS
+-----------+-----------+-----------+--------+
  memcached   demo        memcached   Ready

$ svcat describe bindings memcached --namespace demo
  Name:        memcached
  Namespace:   demo
  Status:      Ready - Injected bind result @ 2018-12-26 07:58:54 +0000 UTC
  Secret:      memcached
  Instance:    memcached

Parameters:
  No parameters defined

Secret Data:
  Host       18 bytes
  Password   4 bytes
  Port       5 bytes
  Protocol   0 bytes
  RootCert   4 bytes
  URI        26 bytes
  Username   4 bytes
```

You can see the secret data by passing `--show-secrets` flag to the above command. The yaml configuration of this `ServiceBinding` resource is as follows:

```console
kubectl get servicebindings memcached --namespace demo -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  creationTimestamp: "2018-12-26T07:58:53Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
  name: memcached
  namespace: demo
  resourceVersion: "132"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/demo/servicebindings/memcached
  uid: 16bb8d67-08e4-11e9-9fa4-0242ac110006
spec:
  externalID: 16bb8d1a-08e4-11e9-9fa4-0242ac110006
  instanceRef:
    name: memcached
  secretName: memcached
  userInfo:
    groups:
    - system:masters
    - system:authenticated
    uid: ""
    username: minikube-user
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: "2018-12-26T07:58:54Z"
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

Here, the status has `Ready` condition which means the binding is now ready for use. This binding operation create a `Secret` named `memcached` in namespace `demo`.

```console
$ kubectl get secrets --namespace demo
NAME                  TYPE                                  DATA   AGE
default-token-2zx6l   kubernetes.io/service-account-token   3      153m
memcached             Opaque                                7      5m57s
```

## Unbinding: Deleting the ServiceBinding

We can now delete the `ServiceBinding` resource we created in the `Binding` step (it is called `Unbinding` the `ServiceInstance`)

```console
$ kubectl delete servicebinding memcached --namespace demo
servicebinding.servicecatalog.k8s.io "memcached" deleted

$ svcat unbind memcached --namespace demo
deleted memcached
```

After completion of unbinding, the `Secret` named `memcached` should be deleted.

```console
$ kubectl get secrets --namespace demo
NAME                  TYPE                                  DATA   AGE
default-token-2zx6l   kubernetes.io/service-account-token   3      174m
```

## Deprovisioning: Deleting the ServiceInstance

After unbinding the `ServiceInstance`, our next step is deleting the `ServiceInstance` resource we created before at the step of provisioning. It is called `Deprovisioning`.

```console
$ kubectl delete serviceinstance memcached --namespace demo
serviceinstance.servicecatalog.k8s.io "memcached" deleted

$ svcat deprovision memcached --namespace demo
deleted memcached
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