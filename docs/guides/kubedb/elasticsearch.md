# Walkthrough Elasticsearch

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

If we've AppsCode Service Broker installed, then we are ready for going forward. If not, then follow the [installation instructions](/docs/setup/install.md).

This document assumes that you've installed Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.27/docs/install.md) to install Service Catalog. Optionally you may install the Service Catalog CLI, `svcat`. Examples for both `svcat` and `kubectl` are provided so that you may follow this walkthrough using `svcat` or using only `kubectl`.

> All commands in this document assume that you're operating out of the root of this repository.

## Check ClusterServiceClass and ClusterServicePlan for Elasticsearch

First, list the available `ClusterServiceClass` resources:

```console
$ kubectl get clusterserviceclasses
NAME                                   EXTERNAL-NAME   BROKER                    AGE
2010d83f-d908-4d9f-879c-ce8f5f527f2a   postgresql      appscode-service-broker   1h
315fc21c-829e-4aa1-8c16-f7921c33550d   elasticsearch   appscode-service-broker   1h
938a70c5-f2bc-4658-82dd-566bed7797e9   mysql           appscode-service-broker   1h
ccfd1c81-e59f-4875-a39f-75ba55320ce0   redis           appscode-service-broker   1h
d690058d-666c-45d8-ba98-fcb9fb47742e   mongodb         appscode-service-broker   1h
d88856cb-fe3f-4473-ba8b-641480da810f   memcached       appscode-service-broker   1h

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

Now, describe the `elasticsearch` class from the `Service Broker`.

```console
$ svcat describe class elasticsearch
  Name:              elasticsearch
  Scope:             cluster
  Description:       KubeDB managed ElasticSearch
  Kubernetes Name:   315fc21c-829e-4aa1-8c16-f7921c33550d
  Status:            Active
  Tags:
  Broker:            appscode-service-broker

Plans:
             NAME                       DESCRIPTION
+----------------------------+--------------------------------+
  demo-elasticsearch-cluster   Demo Elasticsearch cluster
  elasticsearch                Elasticsearch cluster with
                               custom specification
  demo-elasticsearch           Demo Standalone Elasticsearch
                               database
```

To view the details of any plan in this class use command `$ svcat describe plan <class_name>/<plan_name>`. For example:

```console
$ svcat describe plan elasticsearch/elasticsearch --scope cluster
  Name:              elasticsearch
  Description:       Elasticsearch cluster with custom specification
  Kubernetes Name:   6fa212e2-e043-4ae9-91c2-8e5c4403d894
  Status:            Active
  Free:              true
  Class:             elasticsearch

Instances:
No instances defined
```

> Here we,ve used `--scope` flag to specify that our `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServiceBroker` resources are cluster scoped (not namespaced scope)

## Provisioning: Creating a New ServiceInstance

AppsCode Service Broker currently supports three plans for `elasticsearch` class as we can see above. Using `demo-elasticsearch` plan we can provision a demo Elasticsearch database in cluster. Using `demo-elasticsearch-cluster` plan we can provision a demo Elasticsearch database with clustering support in cluster. And using `elasticsearch` plan we can provision a custom Elasticsearch database with custom [Elasticsearch Spec](https://kubedb.com/docs/0.9.0/concepts/databases/elasticsearch/#elasticsearch-spec) of [Elasticsearch CRD](https://kubedb.com/docs/0.9.0/concepts/databases/elasticsearch).

AppsCode Service Broker accept only metadata and [Elasticsearch Spec](https://kubedb.com/docs/0.9.0/concepts/databases/elasticsearch/#elasticsearch-spec) as parameters for the plans of `elasticsearch` class. The metadata and spec should be provided with key `"metadata"` and `"spec"` respectfully. The metadata is optional for all of the plans available. But the spec is required for the clustom plan and it must be valid.

Since a `ClusterServiceClass` named `elasticsearch` exists in the cluster with a `ClusterServicePlan` named `elasticsearch`, we can create a `ServiceInstance` ponting to them with custom specification as parameter.

> Unlike `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources, `ServiceInstance` resources must be namespaced. The latest version of service catelog supports `ServiceBroker`, `ServiceClass` and `ServicePlan` resources that are namespace scoped and alternative to `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources.

Create the `ServiceInstance`:

```console
$ kubectl create -f docs/examples/elasticsearch-instance.yaml
serviceinstance.servicecatalog.k8s.io/elasticsearchdb created
```

After it is created, the service catalog controller will communicate with the service broker server to initaiate provisioning. Now, see the details:

```console
$ svcat describe instance elasticsearchdb --namespace demo
  Name:        elasticsearchdb
  Namespace:   demo
  Status:      Ready - The instance was provisioned successfully @ 2018-12-26 11:16:56 +0000 UTC
  Class:       elasticsearch
  Plan:        elasticsearch

Parameters:
  metadata:
    labels:
      app: my-elasticsearch
  spec:
    enableSSL: true
    storage:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
      storageClassName: standard
    storageType: Durable
    terminationPolicy: WipeOut
    version: 6.3-v1

Bindings:
No bindings defined
```

The yaml configuration of this `ServiceInstance`:

```console
kubectl get serviceinstance elasticsearchdb --namespace demo -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  creationTimestamp: "2018-12-26T11:16:55Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
  name: elasticsearchdb
  namespace: demo
  resourceVersion: "240"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/demo/serviceinstances/elasticsearchdb
  uid: c07af740-08ff-11e9-9fa4-0242ac110006
spec:
  clusterServiceClassExternalName: elasticsearch
  clusterServiceClassRef:
    name: 315fc21c-829e-4aa1-8c16-f7921c33550d
  clusterServicePlanExternalName: elasticsearch
  clusterServicePlanRef:
    name: 6fa212e2-e043-4ae9-91c2-8e5c4403d894
  externalID: c07af6f6-08ff-11e9-9fa4-0242ac110006
  parameters:
    metadata:
      labels:
        app: my-elasticsearch
    spec:
      enableSSL: true
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
      storageType: Durable
      terminationPolicy: WipeOut
      version: 6.3-v1
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
  - lastTransitionTime: "2018-12-26T11:16:56Z"
    message: The instance was provisioned successfully
    reason: ProvisionedSuccessfully
    status: "True"
    type: Ready
  deprovisionStatus: Required
  externalProperties:
    clusterServicePlanExternalID: 6fa212e2-e043-4ae9-91c2-8e5c4403d894
    clusterServicePlanExternalName: elasticsearch
    parameterChecksum: b3e231fcc94d12101a01ce801019f64d3d47ca03c25fc198dc6a1c224e64c166
    parameters:
      metadata:
        labels:
          app: my-elasticsearch
      spec:
        enableSSL: true
        storage:
          accessModes:
          - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
          storageClassName: standard
        storageType: Durable
        terminationPolicy: WipeOut
        version: 6.3-v1
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
$ kubectl create -f docs/examples/elasticsearch-binding.yaml
servicebinding.servicecatalog.k8s.io/elasticsearchdb created
```

Once the `ServiceBinding` resource is created, the service catalog controller initiate binding process by communicating with the service broker server. In general, this step makes the broker server to provide the necessary credentials. Then the service catalog controller will insert them into a Kubernetes `Secret` object.

```console
$ kubectl get servicebindings elasticsearchdb --namespace demo
NAME              SERVICE-INSTANCE   SECRET-NAME       STATUS   AGE
elasticsearchdb   elasticsearchdb    elasticsearchdb   Ready    57s

$ svcat get bindings --namespace demo
       NAME         NAMESPACE      INSTANCE       STATUS
+-----------------+-----------+-----------------+--------+
  elasticsearchdb   demo        elasticsearchdb   Ready

$ svcat describe bindings elasticsearchdb --namespace demo
  Name:        elasticsearchdb
  Namespace:   demo
  Status:      Ready - Injected bind result @ 2018-12-26 11:23:04 +0000 UTC
  Secret:      elasticsearchdb
  Instance:    elasticsearchdb

Parameters:
  No parameters defined

Secret Data:
  Host       24 bytes
  Password   8 bytes
  Port       4 bytes
  Protocol   5 bytes
  RootCert   1520 bytes
  URI        37 bytes
  Username   5 bytes
```

You can see the secret data by passing `--show-secrets` flag to the above command. The yaml configuration of this `ServiceBinding` resource is as follows:

```console
kubectl get servicebindings elasticsearchdb --namespace demo -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  creationTimestamp: "2018-12-26T11:23:04Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
  name: elasticsearchdb
  namespace: demo
  resourceVersion: "250"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/demo/servicebindings/elasticsearchdb
  uid: 9ca83d72-0900-11e9-9fa4-0242ac110006
spec:
  externalID: 9ca83d35-0900-11e9-9fa4-0242ac110006
  instanceRef:
    name: elasticsearchdb
  secretName: elasticsearchdb
  userInfo:
    groups:
    - system:masters
    - system:authenticated
    uid: ""
    username: minikube-user
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: "2018-12-26T11:23:04Z"
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

Here, the status has `Ready` condition which means the binding is now ready for use. This binding operation create a `Secret` named `elasticsearchdb` in namespace `demo`.

```console
$ kubectl get secrets --namespace demo
NAME                   TYPE                                  DATA   AGE
default-token-2qbzg    kubernetes.io/service-account-token   3      6h58m
elasticsearchdb        Opaque                                7      4m14s
elasticsearchdb-auth   Opaque                                9      8m50s
elasticsearchdb-cert   Opaque                                6      8m50s
```

## Unbinding: Deleting the ServiceBinding

We can now delete the `ServiceBinding` resource we created in the `Binding` step (it is called `Unbinding` the `ServiceInstance`)

```console
$ kubectl delete servicebinding elasticsearchdb --namespace demo
servicebinding.servicecatalog.k8s.io "elasticsearchdb" deleted

$ svcat unbind elasticsearchdb --namespace demo
deleted elasticsearchdb
```

After completion of unbinding, the `Secret` named `elasticsearchdb` should be deleted.

```console
$ kubectl get secrets --namespace demo
NAME                   TYPE                                  DATA   AGE
default-token-2qbzg    kubernetes.io/service-account-token   3      6h59m
elasticsearchdb-auth   Opaque                                9      10m
elasticsearchdb-cert   Opaque                                6      10m
```

## Deprovisioning: Deleting the ServiceInstance

After unbinding the `ServiceInstance`, our next step is deleting the `ServiceInstance` resource we created before at the step of provisioning. It is called `Deprovisioning`.

```console
$ kubectl delete serviceinstance elasticsearchdb --namespace demo
serviceinstance.servicecatalog.k8s.io "elasticsearchdb" deleted

$ svcat deprovision elasticsearchdb --namespace demo
deleted elasticsearchdb
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