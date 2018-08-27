#!/usr/bin/env bash

set -eoux pipefail

ORG_NAME=kubedb
REPO_NAME=service-broker
APP_LABEL=service-broker #required for `kubectl describe deploy -n kube-system -l app=$APP_LABEL`

export OPERATOR_NAME=service-broker
export APPSCODE_ENV=concourse
export DOCKER_REGISTRY=kubedbci

# get concourse-common
pushd $REPO_NAME
git status # required, otherwise you'll get error `Working tree has modifications.  Cannot add.`. why?
git subtree pull --prefix hack/libbuild https://github.com/appscodelabs/libbuild.git master --squash -m 'concourse'
popd

# create cluster
source $REPO_NAME/hack/libbuild/concourse/init.sh

# install helm
HELM_VERSION=2.10.0
pushd /tmp
curl -Lo helm https://storage.googleapis.com/kubernetes-helm/helm-v$HELM_VERSION-linux-amd64.tar.gz
tar xvzf helm
mv linux-amd64/helm /bin/helm
popd

helm init
kubectl create clusterrolebinding tiller-cluster-admin \
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


pushd "$GOPATH"/src/github.com/$ORG_NAME/$REPO_NAME

./hack/builddeps.sh
./hack/deploy/service-broker.sh build

# deploy kubedb operator
curl -fsSL https://raw.githubusercontent.com/kubedb/cli/0.8.0/hack/deploy/kubedb.sh | bash

# run tests
source "hack/libbuild/common/lib.sh"
detect_tag ''
ginkgo -r -v -progress - trace test/e2e -- --broker-image=kubedbci/service-broker:$TAG --storage-class=$StorageClass

./hack/concourse/uninstall.sh
