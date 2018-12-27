# Redis Walk-through

This tutorial will show you how to use AppsCode Service Broker to provision and deprovision an Redis cluster and bind to the Redis service.

Before we start, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube). Then install Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.38/docs/install.md) for Service Catalog. Optionally you may install the Service Catalog CLI, `svcat`. Examples for both `svcat` and `kubectl` are provided so that you may follow this walk-through using `svcat` or using only `kubectl`.

If you've AppsCode Service Broker installed, then we are ready for the next step. If not, follow the [instructions](/docs/setup/install.md) to install KubeDB and AppsCode Service Broker.

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> All commands in this document assume that you're operating out of the root of this repository.

## Check ClusterServiceClass and ClusterServicePlan for Redis

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
  memcached                   KubeDB managed Memcached
```

Now, describe the `redis` class from the `Service Broker`.

```console
$ svcat describe class redis
  Name:              redis
  Scope:             cluster
  Description:       KubeDB managed Redis
  Kubernetes Name:   ccfd1c81-e59f-4875-a39f-75ba55320ce0
  Status:            Active
  Tags:
  Broker:            appscode-service-broker

Plans:
     NAME               DESCRIPTION
+------------+--------------------------------+
  redis        Redis with custom
               specification
  demo-redis   Demo Redis
```

To view the details of any plan in this class use command `$ svcat describe plan <class_name>/<plan_name>`. For example:

```console
$ svcat describe plan redis/redis --scope cluster
  Name:              redis
  Description:       Redis with custom specification
  Kubernetes Name:   45716530-cadb-4247-b06a-24a34200d734
  Status:            Active
  Free:              true
  Class:             redis

Instances:
No instances defined
```

> Here we,ve used `--scope` flag to specify that our `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServiceBroker` resources are cluster scoped (not namespaced scope)

## Provisioning: Creating a New ServiceInstance

AppsCode Service Broker currently supports two plans for `redis` class as we can see above. Using `demo-redis` plan we can provision a demo Redis database. And using `redis` plan we can provision a custom Redis database with the full functionality of a [Redis CRD](https://kubedb.com/docs/0.9.0/concepts/databases/redis).

AppsCode Service Broker accepts only metadata and [Redis Spec](https://kubedb.com/docs/0.9.0/concepts/databases/redis/#redis-spec) as parameters for the plans of `redis` class. The metadata and spec should be provided with key `"metadata"` and `"spec"` respectively. The metadata is optional for both of the plans available. But the spec is required for the custom plan and it must be valid.

Since a `ClusterServiceClass` named `redis` exists in the cluster with a `ClusterServicePlan` named `redis`, we can create a `ServiceInstance` pointing to them with custom specification as parameters.

> Unlike `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources, `ServiceInstance` resources must be namespaced. The latest version of .service catalog supports `ServiceBroker`, `ServiceClass` and `ServicePlan` resources that are namespace scoped and alternative to `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources.

Create the `ServiceInstance`:

```console
$ kubectl create -f docs/examples/redis-instance.yaml
serviceinstance.servicecatalog.k8s.io/redisdb created
```

After it is created, the service catalog controller will communicate with the service broker server to initiate provisioning. Now, see the details:

```console
$ svcat describe instance redisdb --namespace demo
svcat describe instance redisdb --namespace demo
  Name:        redisdb
  Namespace:   demo
  Status:      Ready - The instance was provisioned successfully @ 2018-12-26 10:36:33 +0000 UTC
  Class:       redis
  Plan:        redis

Parameters:
  metadata:
    labels:
      app: my-redis
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
    version: 4.0-v1

Bindings:
No bindings defined
```

The yaml configuration of this `ServiceInstance`:

```console
kubectl get serviceinstance redisdb --namespace demo -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  creationTimestamp: "2018-12-26T10:36:32Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
  name: redisdb
  namespace: demo
  resourceVersion: "219"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/demo/serviceinstances/redisdb
  uid: 1cbc6b93-08fa-11e9-9fa4-0242ac110006
spec:
  clusterServiceClassExternalName: redis
  clusterServiceClassRef:
    name: ccfd1c81-e59f-4875-a39f-75ba55320ce0
  clusterServicePlanExternalName: redis
  clusterServicePlanRef:
    name: 45716530-cadb-4247-b06a-24a34200d734
  externalID: 1cbc6afe-08fa-11e9-9fa4-0242ac110006
  parameters:
    metadata:
      labels:
        app: my-redis
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
      version: 4.0-v1
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
  - lastTransitionTime: "2018-12-26T10:36:33Z"
    message: The instance was provisioned successfully
    reason: ProvisionedSuccessfully
    status: "True"
    type: Ready
  deprovisionStatus: Required
  externalProperties:
    clusterServicePlanExternalID: 45716530-cadb-4247-b06a-24a34200d734
    clusterServicePlanExternalName: redis
    parameterChecksum: 8df2444bef8793149d8563446f7c9557ba89b17e4d3ee48ed6baa256717cb267
    parameters:
      metadata:
        labels:
          app: my-redis
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
        version: 4.0-v1
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
$ kubectl create -f docs/examples/redis-binding.yaml
servicebinding.servicecatalog.k8s.io/redisdb created
```

Once the `ServiceBinding` resource is created, the service catalog controller initiates binding process by communicating with the service broker server. In general, the broker server returns the necessary credentials in this step. Then the service catalog controller will insert them into a Kubernetes `Secret` object.

```console
$ kubectl get servicebindings redisdb --namespace demo
NAME      SERVICE-INSTANCE   SECRET-NAME   STATUS   AGE
redisdb   redisdb            redisdb       Ready    40s

$ svcat get bindings redisdb --namespace demo
   NAME     NAMESPACE   INSTANCE   STATUS
+---------+-----------+----------+--------+
  redisdb   demo        redisdb    Ready

$ svcat describe bindings redisdb --namespace demo
  Name:        redisdb
  Namespace:   demo
  Status:      Ready - Injected bind result @ 2018-12-26 10:54:25 +0000 UTC
  Secret:      redisdb
  Instance:    redisdb

Parameters:
  No parameters defined

Secret Data:
  Host       16 bytes
  Password   4 bytes
  Port       4 bytes
  Protocol   5 bytes
  RootCert   4 bytes
  URI        29 bytes
  Username   4 bytes
```

You can see the secret data by passing `--show-secrets` flag to the above command. The yaml configuration of this `ServiceBinding` resource is as follows:

```console
kubectl get servicebindings redisdb --namespace demo -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  creationTimestamp: "2018-12-26T10:54:25Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
  name: redisdb
  namespace: demo
  resourceVersion: "225"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/demo/servicebindings/redisdb
  uid: 9bf494ac-08fc-11e9-9fa4-0242ac110006
spec:
  externalID: 9bf49471-08fc-11e9-9fa4-0242ac110006
  instanceRef:
    name: redisdb
  secretName: redisdb
  userInfo:
    groups:
    - system:masters
    - system:authenticated
    uid: ""
    username: minikube-user
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: "2018-12-26T10:54:25Z"
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

Here, the status has `Ready` condition which means the binding is now ready for use. This binding operation creates a `Secret` named `redisdb` in namespace `demo`.

```console
$ kubectl get secrets --namespace demo
NAME                  TYPE                                  DATA   AGE
default-token-2zx6l   kubernetes.io/service-account-token   3      5h26m
redisdb               Opaque                                7      3m46s
```

## Unbinding: Deleting the ServiceBinding

We can now delete the `ServiceBinding` resource we created in the `Binding` step (it is called `Unbinding` the `ServiceInstance`)

```console
$ kubectl delete servicebinding redisdb --namespace demo
servicebinding.servicecatalog.k8s.io "redisdb" deleted

$ svcat unbind redisdb --namespace demo
deleted redisdb
```

After completion of unbinding, the `Secret` named `redisdb` should be deleted.

```console
$ kubectl get secrets --namespace demo
NAME                  TYPE                                  DATA   AGE
default-token-2zx6l   kubernetes.io/service-account-token   3      5h28m
```

## Deprovisioning: Deleting the ServiceInstance

After unbinding the `ServiceInstance`, our next step is deleting the `ServiceInstance` resource we provisioned before. It is called `Deprovisioning`.

```console
$ kubectl delete serviceinstance redisdb --namespace demo
serviceinstance.servicecatalog.k8s.io "redisdb" deleted

$ svcat deprovision redisdb --namespace demo
deleted redisdb
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