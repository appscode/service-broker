# Walkthrough MySQL

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

If we've AppsCode Service Broker installed, then we are ready for going forward. If not, then follow the [installation instructions](/docs/setup/install.md).

This document assumes that you've installed Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.27/docs/install.md) to install Service Catalog. Optionally you may install the Service Catalog CLI, `svcat`. Examples for both `svcat` and `kubectl` are provided so that you may follow this walkthrough using `svcat` or using only `kubectl`.

> All commands in this document assume that you're operating out of the root of this repository.

## Check ClusterServiceClass and ClusterServicePlan for MySQL

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
  memcached                   KubeDB managed Memcache
```

Now, describe the `mysql` class from the `Service Broker`.

```console
$ svcat describe class mysql
  Name:              mysql
  Scope:             cluster
  Description:       KubeDB managed MySQL
  Kubernetes Name:   938a70c5-f2bc-4658-82dd-566bed7797e9
  Status:            Active
  Tags:
  Broker:            appscode-service-broker

Plans:
     NAME               DESCRIPTION
+------------+--------------------------------+
  demo-mysql   Demo MySQL database
  mysql        MySQL database with custom
               specification
```

To view the details of any plan in this class use command `$ svcat describe plan <class_name>/<plan_name>`. For example:

```console
$ svcat describe plan mysql/mysql --scope cluster
  Name:              mysql
  Description:       MySQL database with custom specification
  Kubernetes Name:   6ed1ab9e-a640-4f26-9328-423b2e3816d7
  Status:            Active
  Free:              true
  Class:             mysql

Instances:
No instances defined
```

> Here we,ve used `--scope` flag to specify that our `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServiceBroker` resources are cluster scoped (not namespaced scope)

## Provisioning: Creating a New ServiceInstance

AppsCode Service Broker currently supports two plans for `mysql` class as we can see above. Using `demo-mysql` plan we can provision a demo MySQL database in cluster. And using `mysql` plan we can provision a custom MySQL database with custom [MySQL Spec](https://kubedb.com/docs/0.9.0/concepts/databases/mysql/#mysql-spec) of [MySQL CRD](https://kubedb.com/docs/0.9.0/concepts/databases/mysql).

AppsCode Service Broker accept only metadata and [MySQL Spec](https://kubedb.com/docs/0.9.0/concepts/databases/mysql/#mysql-spec) as parameters for the plans of `mysql` class. The metadata and spec should be provided with key `"metadata"` and `"spec"` respectfully. The metadata is optional for both of the plans available. But the spec is required for the clustom plan and it must be valid.

Since a `ClusterServiceClass` named `mysql` exists in the cluster with a `ClusterServicePlan` named `mysql`, we can create a `ServiceInstance` ponting to them with custom specification as parameter.

> Unlike `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources, `ServiceInstance` resources must be namespaced. The latest version of service catelog supports `ServiceBroker`, `ServiceClass` and `ServicePlan` resources that are namespace scoped and alternative to `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources.

Create the `ServiceInstance`:

```console
$ kubectl create -f docs/examples/mysql-instance.yaml
serviceinstance.servicecatalog.k8s.io/mysqldb created
```

After it is created, the service catalog controller will communicate with the service broker server to initaiate provisioning. Now, see the details:

```console
$ svcat describe instance mysqldb --namespace demo
  Name:        mysqldb
  Namespace:   demo
  Status:      Ready - The instance was provisioned successfully @ 2018-12-26 09:48:09 +0000 UTC
  Class:       mysql
  Plan:        mysql

Parameters:
  metadata:
    labels:
      app: my-mysql
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
    version: 8.0-v1

Bindings:
No bindings defined
```

The yaml configuration of this `ServiceInstance`:

```console
kubectl get serviceinstance mysqldb --namespace demo -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  creationTimestamp: "2018-12-26T09:48:09Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
  name: mysqldb
  namespace: demo
  resourceVersion: "183"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/demo/serviceinstances/mysqldb
  uid: 5a0e34a0-08f3-11e9-9fa4-0242ac110006
spec:
  clusterServiceClassExternalName: mysql
  clusterServiceClassRef:
    name: 938a70c5-f2bc-4658-82dd-566bed7797e9
  clusterServicePlanExternalName: mysql
  clusterServicePlanRef:
    name: 6ed1ab9e-a640-4f26-9328-423b2e3816d7
  externalID: 5a0e342d-08f3-11e9-9fa4-0242ac110006
  parameters:
    metadata:
      labels:
        app: my-mysql
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
      version: 8.0-v1
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
  - lastTransitionTime: "2018-12-26T09:48:09Z"
    message: The instance was provisioned successfully
    reason: ProvisionedSuccessfully
    status: "True"
    type: Ready
  deprovisionStatus: Required
  externalProperties:
    clusterServicePlanExternalID: 6ed1ab9e-a640-4f26-9328-423b2e3816d7
    clusterServicePlanExternalName: mysql
    parameterChecksum: 71a13d7e5fb129a4f6e279eb77a37c2ddef4bd622cfea6326af69e81c7cc9b59
    parameters:
      metadata:
        labels:
          app: my-mysql
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
        version: 8.0-v1
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
$ kubectl create -f docs/examples/mysql-binding.yaml
servicebinding.servicecatalog.k8s.io/mysqldb created
```

Once the `ServiceBinding` resource is created, the service catalog controller initiate binding process by communicating with the service broker server. In general, this step makes the broker server to provide the necessary credentials. Then the service catalog controller will insert them into a Kubernetes `Secret` object.

```console
$ kubectl get servicebindings mysqldb --namespace demo
NAME      SERVICE-INSTANCE   SECRET-NAME   STATUS   AGE
mysqldb   mysqldb            mysqldb       Ready    44s

$ svcat get bindings mysqldb --namespace demo
   NAME     NAMESPACE   INSTANCE   STATUS
+---------+-----------+----------+--------+
  mysqldb   demo        mysqldb    Ready

$ svcat describe bindings mysqldb --namespace demo
  Name:        mysqldb
  Namespace:   demo
  Status:      Ready - Injected bind result @ 2018-12-26 09:51:50 +0000 UTC
  Secret:      mysqldb
  Instance:    mysqldb

Parameters:
  No parameters defined

Secret Data:
  Host       16 bytes
  Password   16 bytes
  Port       4 bytes
  Protocol   5 bytes
  RootCert   4 bytes
  URI        18 bytes
  Username   4 bytes
```

You can see the secret data by passing `--show-secrets` flag to the above command. The yaml configuration of this `ServiceBinding` resource is as follows:

```console
kubectl get servicebindings mysqldb --namespace demo -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  creationTimestamp: "2018-12-26T09:51:50Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
  name: mysqldb
  namespace: demo
  resourceVersion: "187"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/demo/servicebindings/mysqldb
  uid: dda7b269-08f3-11e9-9fa4-0242ac110006
spec:
  externalID: dda7b22d-08f3-11e9-9fa4-0242ac110006
  instanceRef:
    name: mysqldb
  secretName: mysqldb
  userInfo:
    groups:
    - system:masters
    - system:authenticated
    uid: ""
    username: minikube-user
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: "2018-12-26T09:51:50Z"
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

Here, the status has `Ready` condition which means the binding is now ready for use. This binding operation create a `Secret` named `mysqldb` in namespace `demo`.

```console
$ kubectl get secrets --namespace demo
NAME                  TYPE                                  DATA   AGE
default-token-2zx6l   kubernetes.io/service-account-token   3      4h26m
mysqldb               Opaque                                7      6m41s
mysqldb-auth          Opaque                                2      10m
```

## Unbinding: Deleting the ServiceBinding

We can now delete the `ServiceBinding` resource we created in the `Binding` step (it is called `Unbinding` the `ServiceInstance`)

```console
$ kubectl delete servicebinding mysqldb --namespace demo
servicebinding.servicecatalog.k8s.io "mysqldb" deleted

$ svcat unbind mysqldb --namespace demo
deleted mysqldb
```

After completion of unbinding, the `Secret` named `mysqldb` should be deleted.

```console
$ kubectl get secrets --namespace demo
NAME                  TYPE                                  DATA   AGE
default-token-2zx6l   kubernetes.io/service-account-token   3      4h28m
mysqldb-auth          Opaque                                2      12m
```

## Deprovisioning: Deleting the ServiceInstance

After unbinding the `ServiceInstance`, our next step is deleting the `ServiceInstance` resource we created before at the step of provisioning. It is called `Deprovisioning`.

```console
$ kubectl delete serviceinstance mysqldb --namespace demo
serviceinstance.servicecatalog.k8s.io "mysqldb" deleted

$ svcat deprovision mysqldb --namespace demo
deleted mysqldb
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