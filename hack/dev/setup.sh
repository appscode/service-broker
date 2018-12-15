#!/usr/bin/env bash

set -eou pipefail

show_help() {
    echo "setup.sh - managing tool for developing service-broker locally"
    echo " "
    echo "setup.sh [args]"
    echo " "
    echo "args:"
    echo "--------"
    echo "-h, --help                    show brief help"
    echo "run                           build and run the service-broker"
    echo "install                       deploy clusterservicebroker, service and an endpoint for "
    echo "-h, --help                    show brief help"
    echo "-h, --help                    show brief help"
    echo "-h, --help                    show brief help"
    echo "-h, --help                    show brief help"
    echo "-n, --namespace=NAMESPACE     specify namespace (default: $SERVICE_BROKER_NAMESPACE)"
    echo "    --docker-registry         docker registry used to pull service-broker image (default: $SERVICE_BROKER_DOCKER_REGISTRY)"
    echo "    --image-pull-secret       name of secret used to pull service-broker image"
    echo "    --port                    port number at which the broker will expose"
    echo "    --catalogPath             the path of catalogs for different service plans"
    echo "    --catalogNames            comma separated names of the catalogs for different service plans"
    echo "    --storage-class           name of the storage-class for database storage"
    echo "    --uninstall               uninstall service-broker"
}
pushd $GOPATH/src/github.com/appscode/service-broker

run() {
    hack/make.py;
    service-broker run \
        --kube-config=/home/shudipta/.kube/config \
        --catalog-names="kubedb" \
        --logtostderr \
        -v 5 \
        --catalog-path=hack/deploy/catalogs
}

install() {
    kubectl apply -f hack/dev/broker_for_locally_run.yaml
    kubectl apply -f hack/dev/service_for_locally_run.yaml
}

uninstall() {
    kubectl delete -f hack/dev/broker_for_locally_run.yaml
    kubectl delete -f hack/dev/service_for_locally_run.yaml
}

install_kubedb() {
    curl -fsSL https://raw.githubusercontent.com/kubedb/cli/0.9.0-rc.2/hack/deploy/kubedb.sh| bash
}

uninstall_kubed() {
    curl -fsSL https://raw.githubusercontent.com/kubedb/cli/0.9.0-rc.2/hack/deploy/kubedb.sh| bash -s -- --uninstall --purge
}

install_catalog() {
    helm init
    kubectl get clusterrolebinding tiller-cluster-admin || kubectl create clusterrolebinding tiller-cluster-admin \
        --clusterrole=cluster-admin \
        --serviceaccount=kube-system:default
    helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com

    # check whether repo is added or not
    helm search service-catalog

    onessl wait-until-ready deployment tiller-deploy --namespace kube-system

    # run following if tiller pod is running
    helm install svc-cat/catalog \
        --name catalog \
        --namespace catalog \
        --set namespacedServiceBrokerEnabled=true
}

uninstall_catalog() {
    kubectl get clusterrolebinding tiller-cluster-admin || kubectl delete clusterrolebinding tiller-cluster-admin
    helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com

    helm del catalog --purge
}

"$@"

popd
