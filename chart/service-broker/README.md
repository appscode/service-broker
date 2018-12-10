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

| Parameter           | Description                                                         | Default            |
| --------------------| ------------------------------------------------------------------- | ------------------ |
| `replicaCount`      | Number of Service Broker replicas to create (only 1 is supported) | `1`                |
| `image.registry`    | Docker registry used to pull Service Broker image                 | `appscode`         |
| `image.repository`  | Service Broker container image                                    | `service-broker`   |
| `image.tag`         | Service Broker container image tag                                | `0.1.0`            |
| `image.pullPolicy`  | Service Broker container image pull policy                        | `IfNotPresent`     |
| `imagePullSecrets`  | Specify image pull secrets                                          | `[]` (does not add image pull secrets to deployed pods) |
| `containerPort`     | Specify the container port at which the Service Broker rest server exposes and the container port number | `8080`       |
| `logLevel`          | Log level for operator                                              | `5`                |
| `service.type`      | Specify the type of Service Broker service                        | `ClusterIP`        |
| `service.port`      | Specify the sevice port number                                      | `80`               |
| `resources`         | CPU/Memory resource requests/limits                                 | `{}`               |
| `catalogs.names`    | List of catalogs                                                    | `["kubedb"]`       |
| `catalogs.path`     | The path where catalogs for different database service plans are stored          | `/etc/config/catalogs`       |
| `storageClass`      | StorageClassName for storage                                        | `standard`         |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example:

```console
$ helm install --name appscode-service-broker --set image.pullPolicy=Always appscode/service-broker
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while
installing the chart. For example:

```console
$ helm install --name appscode-service-broker --values values.yaml appscode/service-broker
```
