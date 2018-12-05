#!/usr/bin/env bash

pushd $GOPATH/src/github.com/appscode/service-broker/hack/gendocs
go run main.go
popd
