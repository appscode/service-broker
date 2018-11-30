# AppsCode Service-Broker

This provides an overview of the service broker implemented for installing [AppsCode](https://appscode.com/) cloud services in Kubernetes cluster.

## Overview

A `service broker` is an endpoint that manages a set of software offerings called `services`. Our `Service-Broker` is such a service broker that manages the sevices provided by [AppsCode](https://appscode.com/). It implements the [Open Service Broker API](https://openservicebrokerapi.org/). It provides a simple way to deliver services (those are supported by AppsCode) to applications running on Kubernetes. By creating and managing corresponding resources of different services of AppsCode, this service broker makes it simple for Kubernetes users to use these services. Such as, you can provision `MySQL` service under [Kubedb](https://kubedb.com) by AppsCode in your Kubernetes cluster to configure your applications to use this service.

Using the [Service Catalog](https://kubernetes.io/docs/concepts/extend-kubernetes/service-catalog/), you have the ability to

- List all of the managed services offered by a Service Broker.
- Provision an instance of managed service.
- Bind the instance to your applications and consume any associated credentials.

These listing, provisioning, binding requests are made to the Service Brokers, which are registered with the `Service Catalog`. At present, you can consume the services under `Kubedb` project of AppsCode using our `Service-Broker`. More specifically, now we provides the following services from `Kubedb` through this `Service-Broker`:

- [Elasticsearch](https://kubedb.com/docs/0.9.0-rc.0/guides/elasticsearch/)
- [Memcached](https://kubedb.com/docs/0.9.0-rc.0/guides/memcached/)
- [MongoDB](https://kubedb.com/docs/0.9.0-rc.0/guides/mongodb/)
- [MySQL](https://kubedb.com/docs/0.9.0-rc.0/guides/mysql/)
- [PostgreSQL](https://kubedb.com/docs/0.9.0-rc.0/guides/postgres/)
- [Redis](https://kubedb.com/docs/0.9.0-rc.0/guides/redis/)

## Terminology

The `AppsCode Service Broker` uses some [Open Service Broker terms](https://github.com/openservicebrokerapi/servicebroker/tree/master/spec.md#terminology):

- ***Application***: An entity that might use or bind to a Service Instance.
- ***Platform***: The software that will manage the cloud environment into which Applications are provisioned and Service Brokers are registered. Users will not directly provision Services from Service Brokers, rather they will ask the Platform to manage Services and interact with the Service Brokers for them. For example, the platform for `AppsCode Service-Broker` is the [Kubernetes Service Catalog](https://kubernetes.io/docs/concepts/service-catalog/).
- ***Service***: A managed software offering that can be used by an Application. For example, in our case MySQL, a service under Kubedb, exposes APIs those can be invoked to perform certain actions.
- ***Service Plan***: The representation of different options or tiers (the cost and benefits) for a given service offering.
- ***Service Instance***: An instantiation of a Service Offering and Service Plan.
- ***Service binding***: The ability to use a service instance. This request might refer to an application or some other entity that may wish to use the service instance. In the Platform (Kubernetes Service Catalog), information returned in a binding request is placed in a Kubernetes Secret for the specified namespace.
- ***Service Broker***: Service brokers manage the lifecycle of services. The Platform (Kubernetes Service Catalog) interacts with service brokers to provision and manage service instances and service bindings.

## Behaviour of AppsCode Service-Broker with Service Catalog

The following diagram shows how the Service Broker behaves with Kubernetes Service Broker:

![ref](/docs/images/behaviour.png)
