---
title: CoreOS Prometheus Operator | AppsCode Service Broker
description: Monitoring AppsCode Service Broker Using CoreOS Prometheus Operator
menu:
  product_service-broker_0.3.0:
    identifier: coreos-monitoring
    name: Prometheus Operator
    parent: monitoring-guides
    weight: 30
product_name: service-broker
menu_name: product_service-broker_0.3.0
section_menu_id: guides
---
> New to AppsCode Service Broker? Please start [here](/docs/concepts/README.md).

# Monitoring AppsCode Service Broker Using CoreOS Prometheus Operator

CoreOS [prometheus-operator](https://github.com/coreos/prometheus-operator) provides simple and Kubernetes native way to deploy and configure Prometheus server. This tutorial will show you how to use CoreOS Prometheus operator for monitoring AppsCode Service Broker.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy Prometheus operator and respective resources.

  ```console
  $ kubectl create ns monitoring
  namespace/monitoring created
  ```

- We need a CoreOS [prometheus-operator](https://github.com/coreos/prometheus-operator) instance running. If you don't already have a running instance, deploy one following the docs from [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/coreos-operator/README.md).

## Enable Monitoring in AppsCode Service Broker

Enable Prometheus monitoring using `prometheus.io/coreos-operator` agent while installing AppsCode Service Broker. To know details about how to enable monitoring see [here](/docs/guides/monitoring/overview.md#how-to-enable-monitoring).

Let's install AppsCode Service Broker with monitoring enabled.

**Helm:**

```console
$ helm install appscode/service-broker --name appscode-service-broker --namespace kube-system \
  --set monitoring.enabled=true \
  --set monitoring.agent=prometheus.io/coreos-operator \
  --set monitoring.prometheus.namespace=monitoring \
  --set monitoring.serviceMonitor.labels.k8s-app=prometheus
```

This will create a `ServiceMonitor` crd with name `appscode-service-broker` in `monitoring` namespace for monitoring endpoints of `appscode-service-broker` service. This `ServiceMonitor` will have label `k8s-app: prometheus` as we have set it through `--set monitoring.serviceMonitor.labels.k8s-app=prometheus` flag. This label will be used by Prometheus crd to select this `ServiceMonitor`.

Let's check the ServiceMonitor crd using following command,

```yaml
$ kubectl get servicemonitor -n monitoring appscode-service-broker -o yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  creationTimestamp: 2019-01-09T12:15:47Z
  generation: 1
  labels:
    k8s-app: prometheus
  name: appscode-service-broker
  namespace: monitoring
  resourceVersion: "39617"
  selfLink: /apis/monitoring.coreos.com/v1/namespaces/monitoring/servicemonitors/appscode-service-broker
  uid: 4be916f8-1408-11e9-85c4-0800278ac612
spec:
  endpoints:
  - bearerTokenFile: /var/run/secrets/kubernetes.io/serviceaccount/token
    port: api
    scheme: https
    tlsConfig:
      caFile: /etc/prometheus/secrets/appscode-service-broker-apiserver-cert/tls.crt
      serverName: appscode-service-broker.kube-system.svc
  namespaceSelector:
    matchNames:
    - kube-system
  selector:
    matchLabels:
      app: service-broker
      release: appscode-service-broker
```

AppsCode Service Broker exports  metrics in TLS secured `api` endpoint. So, we have have added flowing two section in `ServicMonitor` specification.

- `tlsConfig` section to establish TLS secured connection.
- `bearerTokenFile` to authorize Prometheus server to AppsCode Service Broker.

Installation process has created a secret named `appscode-service-broker-apiserver-cert` in `monitoring` namespace as we have specified it through `--set monitoring.prometheus.namespace=monitoring`. This secret holds the public certificate of AppsCode Service Broker that has been specified in `tlsConfig` section.

Verify that the secret `appscode-service-broker-apiserver-cert` has been created in `monitoring` namespace.

```console
$ kubectl get secret -n monitoring -l=app=service-broker
NAME                                     TYPE                DATA   AGE
appscode-service-broker-apiserver-cert   kubernetes.io/tls   2      5m40s
```

We are going to specify this secret in [Prometheus](https://github.com/coreos/prometheus-operator/blob/master/Documentation/design.md#prometheus) crd specification. CoreOS Prometheus will mount this secret in `/etc/prometheus/secret/appscode-service-broker-apiserver-cert` directory of respective Prometheus server pod.

Here, `tlsConfig.caFile` indicates the certificate to use for TLS secured connection and `tlsConfig.serverName` is used to verify hostname for which this certificate is valid.

 `bearerTokenFile` denotes the `ServiceAccount` token of the Prometheus server that is going to scape metrics from AppsCode Service Broker. Kubernetes automatically mount it in `/var/run/secrets/kubernetes.io/serviceaccount/token` directory of Prometheus pod. For, an RBAC enabled cluster, we have to grand some permissions to this `ServiceAccount`.

## Configure Prometheus Server

Now, we have to create or configure a `Prometheus` crd to selects above `ServiceMonitor`.

### Configure Existing Prometheus Server

If you already have a Prometheus crd and respective Prometheus server running, you have to update this Prometheus crd to select `appscode-service-broker` ServiceMonitor.

At first, add the ServiceMonitor's  label `k8s-app: prometheus` in `spec.serviceMonitorSelector.matchLabels` field of Prometheus crd.

```yaml
serviceMonitorSelector:
  matchLabels:
    k8s-app: prometheus
```

Then, add secret name `appscode-service-broker-apiserver-cert` in `spec.secrets` section.

```yaml
secrets:
  - appscode-service-broker-apiserver-cert
```

>**Warning:** Updating Prometheus crd specification will cause restart of your Prometheus server. If you don't use a persistent volume for Prometheus storage, you will lost your previously scrapped data.

### Deploy New Prometheus Server

If you don't have any existing Prometheus server running, you have to create a Prometheus crd. CoreOS prometheus operator will deploy respective Prometheus server automatically.

**Create RBAC:**

If you are using an RBAC enabled cluster, you have to give necessary RBAC permissions for Prometheus. Let's create necessary RBAC stuffs for Prometheus,

```console
$ kubectl apply -f https://raw.githubusercontent.com/appscode/third-party-tools/master/monitoring/prometheus/builtin/artifacts/rbac.yaml
clusterrole.rbac.authorization.k8s.io/prometheus created
serviceaccount/prometheus created
clusterrolebinding.rbac.authorization.k8s.io/prometheus created
```

>YAML for the RBAC resources created above can be found [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/artifacts/rbac.yaml).

**Create Prometheus:**

Below is the YAML of `Prometheus` crd that we are going to create for this tutorial,

```yaml
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: prometheus
  namespace: monitoring # use same namespace as ServiceMonitor crd
  labels:
    prometheus: prometheus
spec:
  replicas: 1
  serviceAccountName: prometheus
  serviceMonitorSelector:
    matchLabels:
      k8s-app: prometheus # change this according to your setup
  secrets:
    - appscode-service-broker-apiserver-cert
  resources:
    requests:
      memory: 400Mi
```

Here, `spec.serviceMonitorSelector` is used to select the `ServiceMonitor` crd that is created by AppsCode Service Broker. We have provided `appscode-service-broker-apiserver-cert` secret in `spec.secrets` field. This will be mounted in Prometheus pod.

Let's create the `Prometheus` object we have shown above,

```console
$ kubectl apply -f docs/examples/monitoring/prometheus.yaml
prometheus.monitoring.coreos.com/prometheus created
```

CoreOS prometheus operator watches for `Prometheus` crd. Once a `Prometheus` crd is created, it generates respective configuration and creates a `StatefulSet` to run Prometheus server.

Let's check `StatefulSet` has been created,

```console
$ kubectl get statefulset -n monitoring
NAME                    DESIRED   CURRENT   AGE
prometheus-prometheus   1         1         31s
```

### Verify Monitoring Metrics

Prometheus server is listening to port `9090`. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

At first, let's check if the Prometheus pod is in `Running` state.

```console
$ kubectl get pod prometheus-prometheus-0 -n monitoring
NAME                      READY   STATUS    RESTARTS   AGE
prometheus-prometheus-0   3/3     Running   1          71s
```

Now, run following command on a separate terminal to forward 9090 port of `prometheus-prometheus-0` pod,

```console
$ kubectl port-forward -n monitoring prometheus-prometheus-0 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `api` endpoint of `appscode-service-broker` service as target.

<p align="center">
  <img alt="Prometheus Target" height="100%" src="/docs/images/monitoring/coreos-prom-target.png" style="padding:10px">
</p>

Check the labels marked with red rectangle. These labels confirm that the metrics are coming from AppsCode Service Broker through `api` endpoint of  `appscode-service-broker` service.

Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create beautiful dashboard with collected metrics.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```console
# cleanup Prometheus resources
kubectl delete -n monitoring prometheus prometheus
kubectl delete -n monitoring secret appscode-service-broker-apiserver-cert
kubectl delete -n monitoring servicemonitor appscode-service-broker

# delete namespace
kubectl delete ns monitoring
```

To uninstall AppsCode Service Broker follow this [guide](/docs/setup/uninstall.md).

## Next Steps

- Learn what metrics AppsCode Service Broker exports from [here](/docs/guides/monitoring/overview.md).
- Learn how to monitor AppsCode Service Broker using builtin Prometheus operator from [here](/docs/guides/monitoring/builtin.md).
