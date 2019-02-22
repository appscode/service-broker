---
title: Builtin Prometheus | AppsCode Service Broker
description: Monitoring AppsCode Service Broker with builtin Prometheus
menu:
  product_service-broker_0.1.0:
    identifier: builtin-monitoring
    name: Builtin Prometheus
    parent: monitoring-guides
    weight: 20
product_name: service-broker
menu_name: product_service-broker_0.1.0
section_menu_id: guides
---
> New to AppsCode Service Broker? Please start [here](/docs/concepts/README.md).

# Monitoring AppsCode Service Broker with builtin Prometheus

This tutorial will show you how to configure builtin [Prometheus](https://github.com/prometheus/prometheus) scrapper to monitor AppsCode Service Broker.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://github.com/kubernetes/minikube).

- If you are not familiar with how to configure Prometheus to scrape metrics from various Kubernetes resources, please read the tutorial from [here](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin).

- To keep Prometheus resources isolated, we are going to use a separate namespace called `monitoring` to deploy respective monitoring resources.

  ```console
  $ kubectl create ns monitoring
  namespace/monitoring created
  ```

## Enable AppsCode Service Broker Monitoring

Enable Prometheus monitoring using `prometheus.io/builtin` agent while installing AppsCode Service Broker. To know details about how to enable monitoring see [here](/docs/guides/monitoring/overview.md#how-to-enable-monitoring).

Let's install AppsCode Service Broker with monitoring enabled.

**Helm:**

```console
$ helm install appscode/service-broker --name appscode-service-broker --namespace kube-system \
  --set monitoring.enabled=true \
  --set monitoring.agent=prometheus.io/builtin \
  --set monitoring.prometheus.namespace=monitoring \
```

This will add necessary annotations to `appscode-service-broker` service created in `kube-system` namespace. Prometheus server will scrape metrics using those annotations. Let's check which annotations are added to the service,

```yaml
$ kubectl get service -n kube-system appscode-service-broker -o yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "8443"
    prometheus.io/scheme: https
    prometheus.io/scrape: "true"
  creationTimestamp: 2019-01-09T07:03:55Z
  labels:
    app: service-broker
    chart: service-broker-0.1.0
    heritage: Tiller
    release: appscode-service-broker
  name: appscode-service-broker
  namespace: kube-system
  resourceVersion: "10502"
  selfLink: /api/v1/namespaces/kube-system/services/appscode-service-broker
  uid: ba851ced-13dc-11e9-85c4-0800278ac612
spec:
  clusterIP: 10.106.25.160
  ports:
  - name: api
    port: 443
    protocol: TCP
    targetPort: 8443
  selector:
    app: service-broker
    release: appscode-service-broker
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}
```

Here, `prometheus.io/scrape: "true"` annotation indicates that Prometheus should scrape metrics for this service.

The following three annotations point to `api` endpoints which provides Prometheus metrics.

```ini
prometheus.io/path: /metrics
prometheus.io/port: "8443"
prometheus.io/scheme: https
```

Now, we are ready to configure our Prometheus server to scrape those metrics.

## Configure Prometheus Server

We have to configure a Prometheus scrapping job to scrape the metrics using this service. We are going to configure scrapping job similar to this [kubernetes-service-endpoints](https://github.com/appscode/third-party-tools/tree/master/monitoring/prometheus/builtin#kubernetes-service-endpoints) job. However, as we are going to collect metrics from a TLS secured endpoint, we have to add following configurations:
- [tls_config](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#tls_config) section to establish TLS secured connection.
- `bearer_token_file` to authorize Prometheus server to AppsCode Service Broker.

AppsCode Service Broker has created a secret named `appscode-service-broker-apiserver-cert` in `monitoring` namespace as we have specified it through `--set monitoring.prometheus.namespace=monitoring`. This secret holds the public certificate that is necessary to establish TLS secured connection with AppsCode Service Broker.

Verify that the secret `appscode-service-broker-apiserver-cert` has been created in `monitoring` namespace.

```console
$ kubectl get secret -n monitoring -l=app=service-broker
NAME                                     TYPE                DATA   AGE
appscode-service-broker-apiserver-cert   kubernetes.io/tls   2      3h24m
```

We are going to mount this secret in `/etc/prometheus/secret/appscode-service-broker-apiserver-cert` directory of Prometheus deployment.

Let's configure a Prometheus scrapping job to collect the metrics.

```yaml
- job_name: appscode-service-broker
  kubernetes_sd_configs:
  - role: endpoints
  # we have to provide certificate to establish tls secure connection
  tls_config:
    # public certificate of the AppsCode Service Broker that has been mounted in "/etc/prometheus/secret/<tls secret name>" directory of prometheus server
    ca_file: /etc/prometheus/secret/appscode-service-broker-apiserver-cert/tls.crt
    # dns name for which the certificate is valid
    server_name: appscode-service-broker.kube-system.svc
  # bearer_token_file is required for authorizing prometheus server to AppsCode Service Broker
  bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
  # by default Prometheus server select all kubernetes services as possible target.
  # relabel_config is used to filter only desired endpoints
  relabel_configs:
  # keep only those services that has "prometheus.io/scrape: true" anootation.
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
    regex: true
    action: keep
  # keep only those services that has "app: service-broker" label
  - source_labels: [__meta_kubernetes_service_label_app]
    regex: service-broker
    action: keep
  # keep only those services that has endpoint named "api"
  - source_labels: [__meta_kubernetes_endpoint_port_name]
    regex: api
    action: keep
  # read the metric path from "prometheus.io/path: <path>" annotation
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
    regex: (.+)
    target_label: __metrics_path__
    action: replace
  # read the scrapping scheme from "prometheus.io/scheme: <scheme>" annotation
  - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
    action: replace
    target_label: __scheme__
    regex: (https?)
  # read the port from "prometheus.io/port: <port>" annotation and update scrapping address accordingly
  - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
    action: replace
    target_label: __address__
    regex: ([^:]+)(?::\d+)?;(\d+)
    replacement: $1:$2
  # add service namespace as label to the scrapped metrics
  - source_labels: [__meta_kubernetes_namespace]
    separator: ;
    regex: (.*)
    target_label: namespace
    replacement: $1
    action: replace
  # add service name as label to the scrapped metrics
  - source_labels: [__meta_kubernetes_service_name]
    separator: ;
    regex: (.*)
    target_label: service
    replacement: $1
    action: replace
```

Note that, `bearer_token_file` denotes the `ServiceAccount` token of the Prometheus server. Kubernetes automatically mount it in `/var/run/secrets/kubernetes.io/serviceaccount/token` directory of Prometheus pod. For, an RBAC enabled cluster, we have to grand some permissions to this `ServiceAccount`.

### Configure Existing Prometheus Server

If you already have a Prometheus server running, update the respective `ConfigMap` and add above scrapping job.

Then, you have to mount `appscode-service-broker-apiserver-cert` secret in Prometheus deployment. Add the secret as volume:

```yaml
volumes:
- name: appscode-service-broker-apiserver-cert
  secret:
    defaultMode: 420
    secretName: appscode-service-broker-apiserver-cert
    items: # avoid mounting private key
    - key: tls.crt
      path: tls.crt
```

Then, mount this volume in `/etc/prometheus/secret/appscode-service-broker-apiserver-cert` directory.

```yaml
volumeMounts:
- name: appscode-service-broker-apiserver-cert # mount the secret volume with public certificate of the AppsCode Service Broker
  mountPath: /etc/prometheus/secret/appscode-service-broker-apiserver-cert
```

>**Warning:** Updating deployment will cause restart of your Prometheus server. If you don't use a persistent volume for Prometheus storage, you will lose your previously scrapped data.

### Deploy New Prometheus Server

If you don't have any existing Prometheus server running, you have to deploy one. In this section, we are going to deploy a Prometheus server to collect metrics from AppsCode Service Broker.

**Create ConfigMap:**

At first, create a ConfigMap with the scrapping configuration. Bellow is the YAML of ConfigMap that we are going to create in this tutorial.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: appscode-service-broker-prom-config
  labels:
    app: service-broker
  namespace: monitoring
data:
  prometheus.yml: |-
    global:
      scrape_interval: 30s
      scrape_timeout: 10s
      evaluation_interval: 30s
    scrape_configs:
    - job_name: appscode-service-broker
      kubernetes_sd_configs:
      - role: endpoints
      # we have to provide certificate to establish tls secure connection
      tls_config:
        # public certificate of the AppsCode Service Broker that has been mounted in "/etc/prometheus/secret/<tls secret name>" directory of prometheus server
        ca_file: /etc/prometheus/secret/appscode-service-broker-apiserver-cert/tls.crt
        # dns name for which the certificate is valid
        server_name: appscode-service-broker.kube-system.svc
      # bearer_token_file is required for authorizing prometheus server to AppsCode Service Broker
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      # by default Prometheus server select all kubernetes services as possible target.
      # relabel_config is used to filter only desired endpoints
      relabel_configs:
      # keep only those services that has "prometheus.io/scrape: true" anootation.
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
        regex: true
        action: keep
      # keep only those services that has "app: service-broker" label
      - source_labels: [__meta_kubernetes_service_label_app]
        regex: service-broker
        action: keep
      # keep only those services that has endpoint named "api"
      - source_labels: [__meta_kubernetes_endpoint_port_name]
        regex: api
        action: keep
      # read the metric path from "prometheus.io/path: <path>" annotation
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
        regex: (.+)
        target_label: __metrics_path__
        action: replace
      # read the scrapping scheme from "prometheus.io/scheme: <scheme>" annotation
      - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
        action: replace
        target_label: __scheme__
        regex: (https?)
      # read the port from "prometheus.io/port: <port>" annotation and update scrapping address accordingly
      - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
        action: replace
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
      # add service namespace as label to the scrapped metrics
      - source_labels: [__meta_kubernetes_namespace]
        separator: ;
        regex: (.*)
        target_label: namespace
        replacement: $1
        action: replace
      # add service name as label to the scrapped metrics
      - source_labels: [__meta_kubernetes_service_name]
        separator: ;
        regex: (.*)
        target_label: service
        replacement: $1
        action: replace
```

Let's create the ConfigMap we have shown above,

```console
$ kubectl apply -f docs/examples/monitoring/prom-config.yaml
configmap/appscode-service-broker-prom-config created
```

**Create RBAC:**

If you are using an RBAC enabled cluster, you have to give necessary RBAC permissions for Prometheus. Let's create necessary RBAC stuffs for Prometheus,

```console
$ kubectl apply -f https://raw.githubusercontent.com/appscode/third-party-tools/master/monitoring/prometheus/builtin/artifacts/rbac.yaml
clusterrole.rbac.authorization.k8s.io/prometheus created
serviceaccount/prometheus created
clusterrolebinding.rbac.authorization.k8s.io/prometheus created
```

> YAML for the RBAC resources created above can be found [here](https://github.com/appscode/third-party-tools/blob/master/monitoring/prometheus/builtin/artifacts/rbac.yaml).

**Deploy Prometheus:**

Now, we are ready to deploy Prometheus server. YAML for the deployment that we are going to create for Prometheus is shown below.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      serviceAccountName: prometheus
      containers:
      - name: prometheus
        image: prom/prometheus:v2.4.3
        args:
        - "--config.file=/etc/prometheus/prometheus.yml"
        - "--storage.tsdb.path=/prometheus/"
        ports:
        - containerPort: 9090
        volumeMounts:
        - name: prometheus-config-volume
          mountPath: /etc/prometheus/
        - name: prometheus-storage-volume
          mountPath: /prometheus/
        - name: appscode-service-broker-apiserver-cert # mount the secret volume with public certificate of the AppsCode Service Broker
          mountPath: /etc/prometheus/secret/appscode-service-broker-apiserver-cert
      volumes:
      - name: prometheus-config-volume
        configMap:
          defaultMode: 420
          name: appscode-service-broker-prom-config
      - name: prometheus-storage-volume
        emptyDir: {}
      - name: appscode-service-broker-apiserver-cert
        secret:
          defaultMode: 420
          secretName: appscode-service-broker-apiserver-cert
          items: # avoid mounting private key
          - key: tls.crt
            path: tls.crt
```

Notice that, we have mounted `appscode-service-broker-apiserver-cert` secret as a volume at `/etc/prometheus/secret/appscode-service-broker-apiserver-cert` directory.

> Use a persistent volume instead of `emptyDir` for `prometheus-storage` volume if you don't want to lose collected metrics on Prometheus pod restart.

Now, let's create the deployment,

```console
$ kubectl apply -f docs/examples/monitoring/prom-deploy.yaml
deployment.apps/prometheus created
```

### Verify Monitoring Metrics

Prometheus server is listening to port `9090`. We are going to use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to access Prometheus dashboard.

At first, let's check if the Prometheus pod is in `Running` state.

```console
$ kubectl get pod -n monitoring -l=app=prometheus
NAME                          READY   STATUS    RESTARTS   AGE
prometheus-55577cb994-8f8rp   1/1     Running   0          121m
```

Now, run following command on a separate terminal to forward 9090 port of `prometheus-55577cb994-8f8rp` pod,

```console
$ kubectl port-forward -n monitoring prometheus-55577cb994-8f8rp 9090
Forwarding from 127.0.0.1:9090 -> 9090
Forwarding from [::1]:9090 -> 9090
```

Now, we can access the dashboard at `localhost:9090`. Open [http://localhost:9090](http://localhost:9090) in your browser. You should see `api` endpoint of `appscode-service-broker` service as target.

<p align="center">
  <img alt="Prometheus Target" height="100%" src="/docs/images/monitoring/builtin-prom-target.png" style="padding:10px">
</p>

Check the label marked with red rectangle. This label confirm that the metrics are coming from AppsCode Service Broker through `appscode-service-broker` service.

Now, you can view the collected metrics and create a graph from homepage of this Prometheus dashboard. You can also use this Prometheus server as data source for [Grafana](https://grafana.com/) and create beautiful dashboard with collected metrics.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl delete clusterrole -l=app=prometheus-demo
kubectl delete clusterrolebinding -l=app=prometheus-demo

kubectl delete -n monitoring deployment prometheus
kubectl delete -n monitoring serviceaccount/prometheus
kubectl delete -n monitoring configmap/appscode-service-broker-prom-config
kubectl delete -n monitoring secret appscode-service-broker-apiserver-cert

kubectl delete ns monitoring
```

To uninstall AppsCode Service Broker follow this [guide](/docs/setup/uninstall.md).

## Next Steps

- Learn what metrics AppsCode Service Broker operator exports from [here](/docs/guides/monitoring/overview.md).
- Learn how to monitor AppsCode Service Broker operator using CoreOS Prometheus operator from [here](/docs/guides/monitoring/coreos.md).
