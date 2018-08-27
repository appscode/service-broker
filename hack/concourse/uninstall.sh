#!/usr/bin/env bash

set -x

curl -fsSL https://raw.githubusercontent.com/kubedb/cli/0.8.0/hack/deploy/kubedb.sh | bash -s -- --uninstall --purge
curl -fsSL https://raw.githubusercontent.com/kubedb/cli/0.8.0/hack/deploy/kubedb.sh | bash -s -- --uninstall --purge

helm del --purge catalog
helm reset
kubectl delete clusterrolebinding tiller-cluster-admin
kubectl delete ns catalog


source "hack/libbuild/common/lib.sh"
detect_tag ''

curl -LO https://raw.githubusercontent.com/appscodelabs/libbuild/master/docker.py
chmod +x docker.py
./docker.py del_tag $DOCKER_REGISTRY $OPERATOR_NAME $TAG
