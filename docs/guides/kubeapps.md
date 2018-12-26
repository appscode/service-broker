# Browse and Provision Services from AppsCode Service Broker using Kubeapps

Kubeapps is a web-based UI for deploying and managing applications in Kubernetes clusters. Using Kubeapps you can browse and provision external services from the [Service Catalog](https://github.com/kubernetes-incubator/service-catalog) and available Service Brokers. Please see [Kubeapps](https://github.com/kubeapps/kubeapps) for more information about Kubeapps.

Here we have shown how you can browse and manage services provided by Kubedb in cluster through the Kubeapps Dashboard.

## Prerequisites

You need to have

- A Kubernetes cluster(v1.9+), and the [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [Minikube](https://kubernetes.io/docs/setup/minikube).

- [Helm](http://helm.sh/)(v2.10.0+).

  - Follow the [Helm install instructions](https://github.com/kubernetes/helm/blob/master/docs/install.md).
  - If you already have an appropriate version of Helm installed, execute **`helm init`** to install Tiller, the server-side component of Helm.

- [Kubeapps](https://github.com/kubeapps/kubeapps/blob/master/docs/user/getting-started.md)

  > Also you need to create a Kubernetes API token to access to the Kubeapps Dashboard and ensure that you can open it. To do these follow [Create a Kubernetes API token](https://github.com/kubeapps/kubeapps/blob/master/docs/user/getting-started.md#step-2-create-a-kubernetes-api-token) and [Start the Kubeapps Dashboard](https://github.com/kubeapps/kubeapps/blob/master/docs/user/getting-started.md#step-3-start-the-kubeapps-dashboard) sections.

- [Service Catelog](https://svc-cat.io/docs/install/).
- [Kubedb](https://kubedb.com/docs/0.9.0/setup/install/)
- [AppsCode Service Broker](/docs/setup/install.md).

## Browse the Services Offered by AppsCode Service Broker

Now you are ready to go forward. After being ready the prerequisites in your cluster, you will be able to access the Kubeapps Dashboard. It looks like following:

![ref](/docs/images/kubeapps-dashboard.png)

If you select `All Namespaces` from the drop-down menu of namespaces, you will see `kubeapps`, `catalog`, `kubedb-operator`, `appscode-service-broker` in the list.

![ref](/docs/images/apps-in-all-ns.png)

You can click on the applications icon to see more information. Now go to the "Service Instances" page by clicking the "Service Instances (alpha)" menu from the menubar. Then click "Deploy Service Instance" button and you will see the page that lists the services available from all brokers. Since you have only AppsCode Service Broker(according to this guide) installed, so this list contains only services from AppsCode Service Broker. You can select any of the services from the list.

![ref](/docs/images/service-classes.png)

Click "Select a plan" button on the icon for postgresql. Then a list of plans under this service will be shown.

![ref](/docs/images/service-plans-postgresql.png)

## Manage the Services Offered by AppsCode Service Broker

After completing this section you will be able to manage different services from AppsCode Service Broker. Here we have shown the guide for two of our available plans. They are:

- **`demo-postgresql`** (Demo Standalone PostgreSQL database)
- **`postgresql`** (PostgreSQL database with custom specification).

### Manage `demo-postgresql` Plan

Using this plan you can provision a demo PostgreSQL database in the cluster.

#### Provision

This guide will show you how to provision `demo-postgresql` plan for demo standalone PostgreSQL database. To do this you have to do a series of tasks:

- Click on the "Provision" button for `demo-postgresql` plan in the list page (the page that contains the list of plans for service named `postgresql`).
- A pop-up will be appeared to input a name for provision. In this guide we used "postgresqldb" as the name for provision. You can input as your wish.

  ![ref](/docs/images/enter-provision-name.png)

- Click "Continue"
- Another pop-up will be appeared to input the parameters for provisioning. AppsCode Service Broker accepts only metadata and [Postgres Spec](https://kubedb.com/docs/0.9.0/concepts/databases/postgres/#postgres-spec) of [Postgres CRD](https://kubedb.com/docs/0.9.0/concepts/databases/postgres) as parameters for the plans of `postgresql` service. The metadata and spec should be provided with key `"metadata"` and `"spec"` respectively. The metadata is optional for all of the plans available. But the spec is required for the custom plan and it must be valid. Here we have used the following `json` formated metadata:

  ```json
  {
    "metadata": {
      "name": "postgresqldb",
      "labels": {
        "app": "my-postgres"
      }
    }
  }
  ```

  You can also modify the data.

  ![ref](/docs/images/enter-provision-params.png)

- Click "Submit"
- After some time refresh the page. You will see **`ProvisionedSuccessfully`** as the reason in that page like following:

  ![ref](/docs/images/instance-for-demo-postgresql.png)

- Go to the "Service Instances" page by clicking the "Service Instances (alpha)" menu from the menubar and select the `All Namespaces` or `default` from the drop-dwon for namespaces and a list of current service instances in the selected namespace will be shown.

  ![ref](/docs/images/service-instances-list.png)

#### Binding

In this section we are going to bind the instance created in the above provisionin section. So, do the followings:

- Now click the "Add Binding" button of details page of the newly created instance named `postgresqldb`
- A pop-up will appear to provide a name for binding. You can provide any name you wish. Here we used `postgresqldb` (same as the name of instance).

  ![ref](/docs/images/enter-binding-name.png)

- Click "Continue"
- Another pop-up will ask you to provide the parameters in `json` for binding the instance. Currently AppsCode Service Broker does not require any parameter for binding. So just keep it empty.
- Click "Submit"
- After sometime refresh the page and you will find the status of the binding named `postgresqldb` as **`InjectedBindResult`** and

  ![ref](/docs/images/binding-successful.png)

- Click on the link labeled as "show" and the secret data will be appeared in a pop-up.

  ![ref](/docs/images/show-binding-secret-data.png)

#### Unbinding

You can unbind the newly created binding named `postgresqldb` by clicking the "Remove" button on that page and reload it and the binding has been disappeared.

#### Deprovisioning

To deprovision the previously created instance named `postgresqldb` for plan **`demo-postgresql`**, just click "Deprovision" button and press "Delete" on the pop-up window. Finally refresh the page. That's it. Now you can check the "Service Instances" page by clicking the "Service Instances (alpha)" menu from the menubar and you will find that the instance has been disappeared.

### Manage `postgresql` Plan

Using this plan you can provide a custom specificatation for [Postgres Spec](https://kubedb.com/docs/0.9.0/concepts/databases/postgres/#postgres-spec).

#### Provision

This guide will show you how to provision `postgresql` plan for PostgreSQL database with custom specification. To do this you have to do a series of tasks:

- Click on the "Provision" button for `postgresql` plan in the list page (the page that contains the list of plans for service named `postgresql`).
- A pop-up will be appeared to input a name for provision. In this guide we used "postgresqldb" as the name for provision. You can input as your wish.

  ![ref](/docs/images/enter-provision-name.png)

- Click "Continue"
- Another pop-up will be appeared to input the parameters for provisioning. AppsCode Service Broker accepts only metadata and [Postgres Spec](https://kubedb.com/docs/0.9.0/concepts/databases/postgres/#postgres-spec) of [Postgres CRD](https://kubedb.com/docs/0.9.0/concepts/databases/postgres) as parameters for the plans of `postgresql` service. The metadata and spec should be provided with key `"metadata"` and `"spec"` respectively. The metadata is optional for all of the plans available. But the spec is required for the custom plan and it must be valid. Here we have used the following `json` formated parametes for this custom `postgresql` plan:

  ```json
  {
    "metadata": {
      "labels": {
        "app": "my-postgres"
      }
    },
    "spec": {
      "storage": {
        "accessModes": [
          "ReadWriteOnce"
        ],
        "resources": {
          "requests": {
            "storage": "50Mi"
          }
        },
        "storageClassName": "standard"
      },
      "terminationPolicy": "WipeOut",
      "version": "10.2-v1"
    }
  }
  ```

  You can also modify the data.

- Click "Submit"
- After some time refresh the page. You will see **`ProvisionedSuccessfully`** as the reason in that page like following:

  ![ref](/docs/images/instance-for-custom-postgresql.png)

- Go to the "Service Instances" page by clicking the "Service Instances (alpha)" menu from the menubar and select the `All Namespaces` or `default` from the drop-dwon for namespaces and a list of current service instances in the selected namespace will be shown.

  ![ref](/docs/images/service-instances-list.png)

#### Binding

In this section we are going to bind the instance created in the above provisionin section. So, do the followings:

- Now click the "Add Binding" button of details page of the newly created instance named `postgresqldb`
- A pop-up will appear to provide a name for binding. You can provide any name you wish. Here we used `postgresqldb` (same as the name of instance).

  ![ref](/docs/images/enter-binding-name.png)

- Click "Continue"
- Another pop-up will ask you to provide the parameters in `json` for binding the instance. Currently AppsCode Service Broker does not require any parameter for binding. So just keep it empty.

  ![ref](/docs/images/enter-binding-parametes.png)

- Click "Submit"
- After sometime refresh the page and you will find the status of the binding named `postgresqldb` as **`InjectedBindResult`** and

  ![ref](/docs/images/binding-successful.png)

- Click on the link labeled as "show" and the secret data will be appeared in a pop-up.

  ![ref](/docs/images/show-binding-secret-data-custom.png)

#### Unbinding

You can unbind the newly created binding named `postgresqldb` by clicking the "Remove" button on that page and reload it and the binding has been disappeared.

#### Deprovisioning

To deprovision the previously created instance named `postgresqldb` for plan **`postgresql`**, just click "Deprovision" button and press "Delete" on the pop-up window. Finally refresh the page. That's it. Now you can check the "Service Instances" page by clicking the "Service Instances (alpha)" menu from the menubar and you will find that the instance has been disappeared.

## Cleanup

- [Uninstall AppsCode Service Broker](/docs/setup/uninstall.md).
- [Uninstall Kubedb](https://kubedb.com/docs/0.9.0/setup/uninstall/)
- Run the following commands to delete Service Catalog

  ```console
  $ helm delete --purge catalog
  $ kubectl delete ns catalog
  ```
- Run the following commands to delete Kubeapps

  ```console
  $ helm delete --purge kubeapps
  $ kubectl delete ns kubeapps
  ```