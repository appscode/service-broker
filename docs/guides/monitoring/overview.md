---
title: Monitoring Overview | AppsCode Service Broker
description: Monitoring AppsCode Service Broker
menu:
  product_service-broker_0.3.0:
    identifier: overview-monitoring
    name: Overview
    parent: monitoring-guides
    weight: 10
product_name: service-broker
menu_name: product_service-broker_0.3.0
section_menu_id: guides
---
> New to AppsCode Service Broker? Please start [here](/docs/concepts/README.md).

# Monitoring AppsCode Service Broker

AppsCode Service Broker has native support for monitoring via [Prometheus](https://prometheus.io/). You can use builtin [Prometheus](https://github.com/prometheus/prometheus) scrapper or [CoreOS Prometheus Operator](https://github.com/coreos/prometheus-operator) to monitor AppsCode Service Broker. This tutorial will show you what metrics AppsCode Service Broker exports and how to enable monitoring.

## Overview

AppsCode Service Broker exports Prometheus metrics in `/metrics` path of TLS secured `8443` port. AppsCode Service Broker installation process creates a service with same name as the service broker (i.e. `appscode-service-broker`) in same namespace. Prometheus server can use `api` endpoint of this service to scrape those metrics.

### Exported Metrics

AppsCode Service Broker exports following Prometheus metrics.

**API Server Metrics:**

|                         Metric Name                          |                                                         Uses                                                          |
| ------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------- |
| apiserver_audit_event_total                                  | Counter of audit events generated and sent to the audit backend.                                                      |
| apiserver_client_certificate_expiration_seconds              | Distribution of the remaining lifetime on the certificate used to authenticate a request.                             |
| apiserver_current_inflight_requests                          | Maximal number of currently used inflight request limit of this apiserver per request kind in last second.            |
| apiserver_request_count                                      | Counter of apiserver requests broken out for each verb, API resource, client, and HTTP response contentType and code. |
| apiserver_request_latencies                                  | Response latency distribution in microseconds for each verb, resource and subresource.                                |
| apiserver_request_latencies_summary                          | Response latency summary in microseconds for each verb, resource and subresource.                                     |
| authenticated_user_requests                                  | Counter of authenticated requests broken out by username.                                                             |

**Go Metrics:**

|              Metric Name              |                                Uses                                |
| ------------------------------------- | ------------------------------------------------------------------ |
| go_gc_duration_seconds                | A summary of the GC invocation durations.                          |
| go_goroutines                         | Number of goroutines that currently exist.                         |
| go_memstats_alloc_bytes               | Number of bytes allocated and still in use.                        |
| go_memstats_alloc_bytes_total         | Total number of bytes allocated, even if freed.                    |
| go_memstats_buck_hash_sys_bytes       | Number of bytes used by the profiling bucket hash table.           |
| go_memstats_frees_total               | Total number of frees.                                             |
| go_memstats_gc_sys_bytes              | Number of bytes used for garbage collection system metadata.       |
| go_memstats_heap_alloc_bytes          | Number of heap bytes allocated and still in use.                   |
| go_memstats_heap_idle_bytes           | Number of heap bytes waiting to be used.                           |
| go_memstats_heap_inuse_bytes          | Number of heap bytes that are in use.                              |
| go_memstats_heap_objects              | Number of allocated objects.                                       |
| go_memstats_heap_released_bytes_total | Total number of heap bytes released to OS.                         |
| go_memstats_heap_sys_bytes            | Number of heap bytes obtained from system.                         |
| go_memstats_last_gc_time_seconds      | Number of seconds since 1970 of last garbage collection.           |
| go_memstats_lookups_total             | Total number of pointer lookups.                                   |
| go_memstats_mallocs_total             | Total number of mallocs.                                           |
| go_memstats_mcache_inuse_bytes        | Number of bytes in use by mcache structures.                       |
| go_memstats_mcache_sys_bytes          | Number of bytes used for mcache structures obtained from system.   |
| go_memstats_mspan_inuse_bytes         | Number of bytes in use by mspan structures.                        |
| go_memstats_mspan_sys_bytes           | Number of bytes used for mspan structures obtained from system.    |
| go_memstats_next_gc_bytes             | Number of heap bytes when next garbage collection will take place. |
| go_memstats_other_sys_bytes           | Number of bytes used for other system allocations.                 |
| go_memstats_stack_inuse_bytes         | Number of bytes in use by the stack allocator.                     |
| go_memstats_stack_sys_bytes           | Number of bytes obtained from system for stack allocator.          |
| go_memstats_sys_bytes                 | Number of bytes obtained by system. Sum of all system allocations. |

**HTTP Metrics:**

|              Metrics               |                    Uses                     |
| ---------------------------------- | ------------------------------------------- |
| http_request_duration_microseconds | The HTTP request latencies in microseconds. |
| http_request_size_bytes            | The HTTP request sizes in bytes.            |
| http_requests_total                | Total number of HTTP requests made.         |
| http_response_size_bytes           | The HTTP response sizes in bytes.           |

**Open Service Broker Metrics:**

|    Metric Name    |                                                                    Uses                                                                    |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| osb_actions_total | Total amount of different actions(i.e. `get_catalog`, `provision`, `deprovision`, `bind`, `unbind` etc.) requested to this service broker. |

**Process Metrics:**

|          Metric Name          |                          Uses                          |
| ----------------------------- | ------------------------------------------------------ |
| process_cpu_seconds_total     | Total user and system CPU time spent in seconds.       |
| process_max_fds               | Maximum number of open file descriptors.               |
| process_open_fds              | Number of open file descriptors.                       |
| process_resident_memory_bytes | Resident memory size in bytes.                         |
| process_start_time_seconds    | Start time of the process since unix epoch in seconds. |
| process_virtual_memory_bytes  | Virtual memory size in bytes.                          |

## How to Enable Monitoring

You can enable monitoring through setting some values while installing or upgrading or updating AppsCode Service Broker. You can also choose which monitoring agent to use for monitoring. AppsCode Service Broker will configure respective resources accordingly. Here, are the list of available helm values and their uses,

|            Helm Values             |                     Acceptable Values                      |                          Default                           |                                                                                    Uses                                                                                    |
| ---------------------------------- | ---------------------------------------------------------- | ---------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `monitoring.enabled`               | `true` or `false`                                          | `false`                                                    | Specify whether to monitor AppsCode Service Broker.                                                                                                                        |
| `monitoring.agent`                 | `prometheus.io/builtin` or `prometheus.io/coreos-operator` | `none`                                                     | Specify which monitoring agent to use for monitoring AppsCode Service Broker.                                                                                              |
| `monitoring.prometheus.namespace`  | any namespace                                              | same namespace as AppsCode Service Broker                  | Specify the namespace where Prometheus server is running or will be deployed                                                                                               |
| `monitoring.serviceMonitor.labels` | any label                                                  | `app: <generated app name>` and `release: <release name>`. | Specify the labels for ServiceMonitor. Prometheus crd will select ServiceMonitor using these labels. Only usable when monitoring agent is `prometheus.io/coreos-operator`. |

You have to provides these values while installing or upgrading or updating AppsCode Service Broker. Here, an example is given which enable monitoring with `prometheus.io/coreos-operator` Prometheuse server for the service broker.

**Helm:**

```console
$ helm install appscode/service-broker --name appscode-service-broker --namespace kube-system \
  --set monitoring.enabled=true \
  --set monitoring.agent=prometheus.io/coreos-operator \
  --set monitoring.prometheus.namespace=monitoring \
  --set monitoring.serviceMonitor.labels.k8s-app=prometheus
```

## Next Steps

- Learn how to monitor AppsCode Service Broker using built-in Prometheus from [here](/docs/guides/monitoring/builtin.md).
- Learn how to monitor AppsCode Service Broker using CoreOS Prometheus from [here](/docs/guides/monitoring/coreos.md).
