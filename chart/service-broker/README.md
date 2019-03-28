# AppsCode Service Broker
[AppsCode Service Broker](https://github.com/appscode/service-broker) - Run AppsCode cloud services on Kubernetes via the Open Service Broker API.

## TL;DR;

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm install appscode/service-broker --name appscode-service-broker --namespace kube-system
```

## Introduction

This chart bootstraps a [Service-Broker](https://github.com/appscode/service-broker) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.9+

## Installing the Chart

To install the chart with the release name `appscode-service-broker`:

```console
$ helm install appscode/service-broker --name appscode-service-broker --namespace kube-system
```

The command deploys AppsCode Service Broker on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `appscode-service-broker`:

```console
$ helm delete appscode-service-broker
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the AppsCode Service Broker chart and their default values.

|                   Parameter                   |                                                                                Description                                                                                 |                          Default                          |
| --------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------- |
| `replicaCount`                                | Number of Service Broker replicas to create (only 1 is supported)                                                                                                          | `1`                                                       |
| `broker.registry`                             | Docker registry used to pull service broker image                                                                                                                          | `appscode`                                                |
| `broker.repository`                           | Service broker container image                                                                                                                                             | `service-broker`                                          |
| `broker.tag`                                  | Service broker container image tag                                                                                                                                         | `0.3.0`                                                   |
| `cleaner.registry`                            | Docker registry used to pull service broker cleaner image                                                                                                                  | `appscode`                                                |
| `cleaner.repository`                          | Service broker cleaner container image                                                                                                                                     | `kubectl`                                                 |
| `cleaner.tag`                                 | Service broker cleaner container image tag                                                                                                                                 | `v1.12`                                                   |
| `imagePullSecrets`                            | Specify image pull secrets                                                                                                                                                 | `nil` (does not add image pull secrets to deployed pods)  |
| `imagePullPolicy`                             | Image pull policy                                                                                                                                                          | `IfNotPresent`                                            |
| `criticalAddon`                               | If true, installs service broker as critical addon                                                                                                                         | `false`                                                   |
| `logLevel`                                    | Log level for service broker                                                                                                                                               | `3`                                                       |
| `resources`                                   | CPU/Memory resource requests/limits                                                                                                                                        | `{}`                                                      |
| `affinity`                                    | Affinity rules for pod assignment                                                                                                                                          | `{}`                                                      |
| `nodeSelector`                                | Node labels for pod assignment                                                                                                                                             | `{}`                                                      |
| `tolerations`                                 | Tolerations used pod assignment                                                                                                                                            | `{}`                                                      |
| `serviceAccount.create`                       | If `true`, create a new service account                                                                                                                                    | `true`                                                    |
| `serviceAccount.name`                         | Service account to be used. If not set and `serviceAccount.create` is `true`, a name is generated using the fullname template                                              | ``                                                        |
| `apiserver.useKubeapiserverFqdnForAks`        | If true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522                                                                    | `true`                                                    |
| `apiserver.healthcheck.enabled`               | Enable readiness and liveliness probes                                                                                                                                     | `true`                                                    |
| `enableAnalytics`                             | Send usage events to Google Analytics                                                                                                                                      | `true`                                                    |
| `monitoring.enabled`                          | Specify whether to monitor service broker.                                                                                                                                 | `false`                                                   |
| `monitoring.agent`                            | Specify which monitoring agent to use for monitoring service broker. It accepts either `prometheus.io/builtin` or `prometheus.io/coreos-operator`.                         | `none`                                                    |
| `monitoring.prometheus.namespace`             | Specify the namespace where Prometheus server is running or will be deployed.                                                                                              | Release namespace                                         |
| `monitoring.serviceMonitor.labels`            | Specify the labels for ServiceMonitor. Prometheus crd will select ServiceMonitor using these labels. Only usable when monitoring agent is `prometheus.io/coreos-operator`. | `app: <generated app name>` and `release: <release name>` |
| `catalog.names`                               | List of catalog                                                                                                                                                            | `["kubedb"]`                                              |
| `catalog.path`                                | The path where catalog for different service plans are mounted                                                                                                             | `/etc/config/catalog`                                     |
| `catalog.controller.enabled`                  | Specify `false` if used without service catalog (eg CloudFoundry)                                                                                                          | `true`                                                    |
| `catalog.controller.serviceAccount.namespace` | Namespace of service catalog manager controller service account                                                                                                            | `catalog`                                                 |
| `catalog.controller.serviceAccount.name`      | Name of service catalog controller manager service account                                                                                                                 | `service-catalog-controller-manager`                      |
| `defaultNamespace`                            | The default namespace for brokers when the request doesn't specify                                                                                                         | `default`                                                 |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example:

```console
$ helm install --name appscode-service-broker --set image.pullPolicy=Always appscode/service-broker
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while
installing the chart. For example:

```console
$ helm install --name appscode-service-broker --values values.yaml appscode/service-broker
```
