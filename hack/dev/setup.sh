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
    echo "run                           build and run the service-broker locally"
    echo "install                       deploy clusterservicebroker, service and an endpoint for service-broker that is running locally"
    echo "uninstall                     delete clusterservicebroker, service and an endpoint those are created before"
    echo "install_kubedb                install KubeDB"
    echo "uninstall_kubedb              uninstall KubeDB"
    echo "install_catalog               install Service Catalog"
    echo "uninstall_catalg              uninstall Service Catalog"
}

export ARG=
while test $# -gt 0; do
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        run|install|uninstall|install_kubedb|uninstall_kubedb|install_catalog|uninstall_catalg)
            if [[ ${ARG} == "" ]]; then
                export ARG=$1
            else
                echo "require only one argument"
                exit 1
            fi
            shift
            ;;
        *)
            show_help
            exit 1
            ;;
    esac
done

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
    curl -fsSL https://raw.githubusercontent.com/kubedb/cli/0.9.0/hack/deploy/kubedb.sh| bash
}

uninstall_kubedb() {
    curl -fsSL https://raw.githubusercontent.com/kubedb/cli/0.9.0/hack/deploy/kubedb.sh| bash -s -- --uninstall --purge
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

${ARG}

popd
