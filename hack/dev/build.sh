#!/bin/bash
set -eou pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT=$GOPATH/src/github.com/kubedb/service-broker

export DOCKER_REGISTRY=shudipta
export IMG=db-broker
export TAG=try
export ONESSL=

export NAME=my-broker
export NAMESPACE=db-broker
export SERVICE_ACCOUNT="$NAME"
export APP=db-broker
export IMAGE_PULL_POLICY=Always
#export IMAGE_PULL_POLICY=IfNotPresent


build() {
    pushd $REPO_ROOT
        mkdir -p hack/docker
        go build -o hack/docker/db-broker cmd/mysqldb/main.go
        cp hack/dev/kubedb.sh hack/docker/kubedb.sh

        pushd hack/docker
            chmod 755 db-broker
            cat > Dockerfile <<EOL
FROM busybox

COPY db-broker /bin/db-broker/db_broker
RUN mkdir -p /bin/db-broker/hack/dev
COPY kubedb.sh /bin/db-broker/hack/dev/kubedb.sh

EXPOSE 8088
WORKDIR /bin/db-broker/
EOL
            local cmd="docker build -t $DOCKER_REGISTRY/$IMG:$TAG ."
            echo $cmd; $cmd
            cmd="docker push $DOCKER_REGISTRY/$IMG:$TAG"
            echo $cmd; $cmd
        popd

        rm -rf hack/docker
    popd
}

ensure_onessl() {
    if [ -x "$(command -v onessl)" ]; then
        export ONESSL=onessl
    else
        # ref: https://stackoverflow.com/a/27776822/244009
        case "$(uname -s)" in
            Linux)
                curl -fsSL -o onessl https://github.com/kubepack/onessl/releases/download/0.3.0/onessl-linux-amd64
                chmod +x onessl
                export ONESSL=./onessl
                ;;

            *)
                echo 'other OS'
                ;;
        esac
    fi
}

parse_flags() {
    while test $# -gt 0; do
        case "$1" in
            -n)
                shift
                if test $# -gt 0; then
                    export NAMESPACE=$1
                else
                    echo "no namespace specified"
                    exit 1
                fi
                shift
                ;;
            --namespace*)
                export NAMESPACE=`echo $1 | sed -e 's/^[^=]*=//g'`
                shift
                ;;
            --docker-registry*)
                export DOCKER_REGISTRY=`echo $1 | sed -e 's/^[^=]*=//g'`
                shift
                ;;
        esac
    done
}

deploy_db_broker() {
    local found=0
    ns=($(kubectl get ns -o jsonpath='{range .items[*]}{.metadata.name} {end}'))
    for n in "${ns[@]}"; do
        if [ "$n" = "$NAMESPACE" ]; then
            export found=1
        fi
    done
    if [ "$found" -eq 0 ]; then
        kubectl create ns $NAMESPACE
    fi

    cat hack/dev/deployment.yaml | $ONESSL envsubst | kubectl apply -f -
    cat hack/dev/service.yaml | $ONESSL envsubst | kubectl apply -f -
    cat hack/dev/rbac.yaml | $ONESSL envsubst | kubectl create -f -
    cat hack/dev/broker.yaml | $ONESSL envsubst | kubectl apply -f -

    echo
    echo "waiting until db-broker deployment is ready"
    $ONESSL wait-until-ready deployment $NAME --namespace $NAMESPACE || { echo "db-broker deployment failed to be ready"; exit 1; }

    echo
    echo "Successfully installed db-broker in $NAMESPACE namespace!"
}

run() {
    pushd $REPO_ROOT
        ensure_onessl
        parse_flags
        deploy_db_broker
    popd
}

uninstall() {
    pushd $REPO_ROOT
        # delete db-broker
        kubectl delete deployment -l app=$APP --namespace $NAMESPACE
        kubectl delete service -l app=$APP --namespace $NAMESPACE
        # delete RBAC objects, if --rbac flag was used.
        kubectl delete serviceaccount -l app=$APP --namespace $NAMESPACE
        kubectl delete clusterrolebindings -l app=$APP

        echo "waiting for db-broker pod to stop running"
        for (( ; ; )); do
           pods=($(kubectl get pods --all-namespaces -l app=$APP -o jsonpath='{range .items[*]}{.metadata.name} {end}'))
           total=${#pods[*]}
            if [ $total -eq 0 ] ; then
                break
            fi
           sleep 2
        done

        kubectl delete clusterservicebroker -l app=$APP
        kubectl delete ns $NAMESPACE

        echo
        echo "Successfully uninstalled db-broker!"
        exit 0
    popd
}

"$@"