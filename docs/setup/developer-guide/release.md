---
title: Release | AppsCode Service Broker
description: AppsCode Service Broker Release
menu:
  product_service-broker_0.3.1:
    identifier: release
    name: Release
    parent: developer-guide
    weight: 15
product_name: service-broker
menu_name: product_service-broker_0.3.1
section_menu_id: setup
---
# Release Process

The following steps must be done from a Linux x64 bit machine.

- Do a global replacement of tags so that docs point to the next release.
- Push changes to the `release-x` branch and apply new tag.
- Push all the changes to remote repo.
- Build and push service-broker docker image:

```console
$ cd ~/go/src/github.com/appscode/service-broker
./hack/release.sh
```

- Now, update the release notes in Github. See previous release notes to get an idea what to include there.
