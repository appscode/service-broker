# Walkthrough Postgres

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

If we've AppsCode Service Broker installed, then we are ready for going forward. If not, then follow the [installation instructions](/docs/setup/install.md).

This document assumes that you've installed Service Catalog onto your cluster. If you haven't, please see the [installation instructions](https://github.com/kubernetes-incubator/service-catalog/blob/v0.1.27/docs/install.md) to install Service Catalog. Optionally you may install the Service Catalog CLI, `svcat`. Examples for both `svcat` and `kubectl` are provided so that you may follow this walkthrough using `svcat` or using only `kubectl`.

> All commands in this document assume that you're operating out of the root of this repository.

## Check ClusterServiceClass and ClusterServicePlan for Postgres

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

Now, describe the `postgresql` class from the `Service Broker`.

```console
$ svcat describe class postgresql
  Name:              postgresql
  Scope:             cluster
  Description:       KubeDB managed PostgreSQL
  Kubernetes Name:   2010d83f-d908-4d9f-879c-ce8f5f527f2a
  Status:            Active
  Tags:
  Broker:            appscode-service-broker

Plans:
         NAME                   DESCRIPTION
+--------------------+--------------------------------+
  postgresql           PostgreSQL database with
                       custom specification
  demo-ha-postgresql   Demo HA PostgreSQL database
  demo-postgresql      Demo Standalone PostgreSQL
                       database
```

To view the details of any plan in this class use command `$ svcat describe plan <class_name>/<plan_name>`. For example:

```console
$ svcat describe plan postgresql/postgresql --scope cluster
  Name:              postgresql
  Description:       PostgreSQL database with custom specification
  Kubernetes Name:   13373a9b-d5f5-4d9a-88df-d696bbc19071
  Status:            Active
  Free:              true
  Class:             postgresql

Instances:
No instances defined
```

> Here we,ve used `--scope` flag to specify that our `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServiceBroker` resources are cluster scoped (not namespaced scope)

## Provisioning: Creating a New ServiceInstance

AppsCode Service Broker currently supports three plans for `postgresql` class as we can see above. Using `demo-postgresql` plan we can provision a demo Postgres database in cluster. Using `demo-ha-postgresql` plan we can provision a demo HA Postgres database in cluster. And using `postgresql` plan we can provision a custom Postgres database with custom [Postgres Spec](https://kubedb.com/docs/0.9.0/concepts/databases/postgres/#postgres-spec) of [Postgres CRD](https://kubedb.com/docs/0.9.0/concepts/databases/postgres).

AppsCode Service Broker accept only metadata and [Postgres Spec](https://kubedb.com/docs/0.9.0/concepts/databases/postgres/#postgres-spec) as parameters for the plans of `postgresql` class. The metadata and spec should be provided with key `"metadata"` and `"spec"` respectfully. The metadata is optional for all of the plans available. But the spec is required for the clustom plan and it must be valid.

Since a `ClusterServiceClass` named `postgresql` exists in the cluster with a `ClusterServicePlan` named `postgresql`, we can create a `ServiceInstance` ponting to them with custom specification as parameter.

> Unlike `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources, `ServiceInstance` resources must be namespaced. The latest version of service catelog supports `ServiceBroker`, `ServiceClass` and `ServicePlan` resources that are namespace scoped and alternative to `ClusterServiceBroker`, `ClusterServiceClass` and `ClusterServicePlan` resources.

Create the `ServiceInstance`:

```console
$ kubectl create -f docs/examples/postgresql-instance.yaml
serviceinstance.servicecatalog.k8s.io/postgresqldb created
```

After it is created, the service catalog controller will communicate with the service broker server to initaiate provisioning. Now, see the details:

```console
$ svcat describe instance postgresqldb --namespace demo
  Name:        postgresqldb
  Namespace:   demo
  Status:      Ready - The instance was provisioned successfully @ 2018-12-26 10:17:55 +0000 UTC
  Class:       postgresql
  Plan:        postgresql

Parameters:
  metadata:
    labels:
      app: my-postgres
  spec:
    storage:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 50Mi
      storageClassName: standard
    terminationPolicy: WipeOut
    version: 10.2-v1

Bindings:
No bindings defined
```

The yaml configuration of this `ServiceInstance`:

```console
kubectl get serviceinstance postgresqldb --namespace demo -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceInstance
metadata:
  creationTimestamp: "2018-12-26T10:17:54Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
  name: postgresqldb
  namespace: demo
  resourceVersion: "202"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/demo/serviceinstances/postgresqldb
  uid: 82223da0-08f7-11e9-9fa4-0242ac110006
spec:
  clusterServiceClassExternalName: postgresql
  clusterServiceClassRef:
    name: 2010d83f-d908-4d9f-879c-ce8f5f527f2a
  clusterServicePlanExternalName: postgresql
  clusterServicePlanRef:
    name: 13373a9b-d5f5-4d9a-88df-d696bbc19071
  externalID: 82223d44-08f7-11e9-9fa4-0242ac110006
  parameters:
    metadata:
      labels:
        app: my-postgres
    spec:
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 50Mi
        storageClassName: standard
      terminationPolicy: WipeOut
      version: 10.2-v1
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
  - lastTransitionTime: "2018-12-26T10:17:55Z"
    message: The instance was provisioned successfully
    reason: ProvisionedSuccessfully
    status: "True"
    type: Ready
  deprovisionStatus: Required
  externalProperties:
    clusterServicePlanExternalID: 13373a9b-d5f5-4d9a-88df-d696bbc19071
    clusterServicePlanExternalName: postgresql
    parameterChecksum: cdc7984a164fef77f2027f285e7b2bdc931a861461726f5aec961c812a3745e0
    parameters:
      metadata:
        labels:
          app: my-postgres
      spec:
        storage:
          accessModes:
          - ReadWriteOnce
          resources:
            requests:
              storage: 50Mi
          storageClassName: standard
        terminationPolicy: WipeOut
        version: 10.2-v1
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
$ kubectl create -f docs/examples/postgresql-binding.yaml
servicebinding.servicecatalog.k8s.io/postgresqldb created
```

Once the `ServiceBinding` resource is created, the service catalog controller initiate binding process by communicating with the service broker server. In general, this step makes the broker server to provide the necessary credentials. Then the service catalog controller will insert them into a Kubernetes `Secret` object.

```console
$ kubectl get servicebindings postgresqldb --namespace demo
NAME           SERVICE-INSTANCE   SECRET-NAME    STATUS   AGE
postgresqldb   postgresqldb       postgresqldb   Ready    1m

$ svcat get bindings postgresqldb --namespace demo
      NAME       NAMESPACE     INSTANCE     STATUS
+--------------+-----------+--------------+--------+
  postgresqldb   demo        postgresqldb   Ready

$ svcat describe bindings postgresqldb --namespace demo
  Name:        postgresqldb
  Namespace:   demo
  Status:      Ready - Injected bind result @ 2018-12-26 10:21:12 +0000 UTC
  Secret:      postgresqldb
  Instance:    postgresqldb

Parameters:
  No parameters defined

Secret Data:
  Host       21 bytes
  Password   16 bytes
  Port       4 bytes
  Protocol   10 bytes
  RootCert   4 bytes
  URI        56 bytes
  Username   8 bytes
```

You can see the secret data by passing `--show-secrets` flag to the above command. The yaml configuration of this `ServiceBinding` resource is as follows:

```console
kubectl get servicebindings postgresqldb --namespace demo -o yaml
```

Output:

```yaml
apiVersion: servicecatalog.k8s.io/v1beta1
kind: ServiceBinding
metadata:
  creationTimestamp: "2018-12-26T10:21:11Z"
  finalizers:
  - kubernetes-incubator/service-catalog
  generation: 1
  labels:
    app: appscode-service-broker
  name: postgresqldb
  namespace: demo
  resourceVersion: "206"
  selfLink: /apis/servicecatalog.k8s.io/v1beta1/namespaces/demo/servicebindings/postgresqldb
  uid: f7be4740-08f7-11e9-9fa4-0242ac110006
spec:
  externalID: f7be46f6-08f7-11e9-9fa4-0242ac110006
  instanceRef:
    name: postgresqldb
  secretName: postgresqldb
  userInfo:
    groups:
    - system:masters
    - system:authenticated
    uid: ""
    username: minikube-user
status:
  asyncOpInProgress: false
  conditions:
  - lastTransitionTime: "2018-12-26T10:21:12Z"
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

Here, the status has `Ready` condition which means the binding is now ready for use. This binding operation create a `Secret` named `postgresqldb` in namespace `demo`.

```console
$ kubectl get secrets --namespace demo
NAME                       TYPE                                  DATA   AGE
default-token-2zx6l        kubernetes.io/service-account-token   3      4h53m
postgresqldb               Opaque                                7      3m28s
postgresqldb-auth          Opaque                                2      6m45s
postgresqldb-token-vll77   kubernetes.io/service-account-token   3      6m45s
```

## Unbinding: Deleting the ServiceBinding

We can now delete the `ServiceBinding` resource we created in the `Binding` step (it is called `Unbinding` the `ServiceInstance`)

```console
$ kubectl delete servicebinding postgresqldb --namespace demo
servicebinding.servicecatalog.k8s.io "postgresqldb" deleted

$ svcat unbind postgresqldb --namespace demo
deleted postgresqldb
```

After completion of unbinding, the `Secret` named `postgresqldb` should be deleted.

```console
$ kubectl get secrets --namespace demo
NAME                       TYPE                                  DATA   AGE
default-token-2zx6l        kubernetes.io/service-account-token   3      4h55m
postgresqldb-auth          Opaque                                2      8m49s
postgresqldb-token-vll77   kubernetes.io/service-account-token   3      8m49s
```

## Deprovisioning: Deleting the ServiceInstance

After unbinding the `ServiceInstance`, our next step is deleting the `ServiceInstance` resource we created before at the step of provisioning. It is called `Deprovisioning`.

```console
$ kubectl delete serviceinstance postgresqldb --namespace demo
serviceinstance.servicecatalog.k8s.io "postgresqldb" deleted

$ svcat deprovision postgresqldb --namespace demo
deleted postgresqldb
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