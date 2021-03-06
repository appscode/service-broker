---
title: Changelog | Service Broker
description: Changelog
menu:
  product_service-broker_0.3.1:
    identifier: changelog-service-broker
    name: Changelog
    parent: welcome
    weight: 10
product_name: service-broker
menu_name: product_service-broker_0.3.1
section_menu_id: welcome
url: /products/service-broker/0.3.1/welcome/changelog/
aliases:
  - /products/service-broker/0.3.1/CHANGELOG/
---

# Change Log

## [0.3.1](https://github.com/appscode/service-broker/tree/0.3.1) (2019-03-28)
[Full Changelog](https://github.com/appscode/service-broker/compare/0.3.0...0.3.1)

**Fixed bugs:**

- Convert Credentials into a map respecting json tags [\#57](https://github.com/appscode/service-broker/pull/57) ([tamalsaha](https://github.com/tamalsaha))

## [0.3.0](https://github.com/appscode/service-broker/tree/0.3.0) (2019-03-28)
[Full Changelog](https://github.com/appscode/service-broker/compare/0.2.0...0.3.0)

**Closed issues:**

- Provide best practice plans [\#32](https://github.com/appscode/service-broker/issues/32)

**Merged pull requests:**

- Prepare release 0.3.0 [\#56](https://github.com/appscode/service-broker/pull/56) ([tamalsaha](https://github.com/tamalsaha))
- Don't create ClusterServiceBroker if not used with svc catalog [\#55](https://github.com/appscode/service-broker/pull/55) ([tamalsaha](https://github.com/tamalsaha))
- Detect instance name in clusters wihtout svc-catalog [\#54](https://github.com/appscode/service-broker/pull/54) ([tamalsaha](https://github.com/tamalsaha))
- Update KubeDB dependency to 0.11.0 [\#53](https://github.com/appscode/service-broker/pull/53) ([tamalsaha](https://github.com/tamalsaha))
- Add app labels to CRDs [\#51](https://github.com/appscode/service-broker/pull/51) ([tamalsaha](https://github.com/tamalsaha))

## [0.2.0](https://github.com/appscode/service-broker/tree/0.2.0) (2019-03-14)
[Full Changelog](https://github.com/appscode/service-broker/compare/0.1.0...0.2.0)

**Closed issues:**

- Provisioning is failing in reason of missing configurations [\#38](https://github.com/appscode/service-broker/issues/38)

**Merged pull requests:**

- Prepare docs for 0.2.0 release [\#52](https://github.com/appscode/service-broker/pull/52) ([tamalsaha](https://github.com/tamalsaha))
- Use app-{InstanceID} as instance name when no svcinstance found. [\#50](https://github.com/appscode/service-broker/pull/50) ([tamalsaha](https://github.com/tamalsaha))
- Add a default namespace option [\#49](https://github.com/appscode/service-broker/pull/49) ([tamalsaha](https://github.com/tamalsaha))
- Update Kubernetes client libraries to 1.13.0 [\#47](https://github.com/appscode/service-broker/pull/47) ([tamalsaha](https://github.com/tamalsaha))

## [0.1.0](https://github.com/appscode/service-broker/tree/0.1.0) (2019-02-22)
**Closed issues:**

- Expose /metrics and /heathz [\#34](https://github.com/appscode/service-broker/issues/34)
- Remove storing provision request info into map\[\] [\#22](https://github.com/appscode/service-broker/issues/22)
- Fix installer script to usable from curl | bash [\#18](https://github.com/appscode/service-broker/issues/18)
- Send analytics [\#17](https://github.com/appscode/service-broker/issues/17)
- Explore service broker integration [\#15](https://github.com/appscode/service-broker/issues/15)
- Implement MySQL service broker for Kubernetes [\#1](https://github.com/appscode/service-broker/issues/1)
- Add monitoring to installer [\#35](https://github.com/appscode/service-broker/issues/35)
- Secure service broker using TLS [\#33](https://github.com/appscode/service-broker/issues/33)
- TODO List [\#14](https://github.com/appscode/service-broker/issues/14)

**Merged pull requests:**

- Fix bad links [\#44](https://github.com/appscode/service-broker/pull/44) ([tamalsaha](https://github.com/tamalsaha))
- Add developer guide [\#43](https://github.com/appscode/service-broker/pull/43) ([tamalsaha](https://github.com/tamalsaha))
- Add Hugo frontmatter [\#42](https://github.com/appscode/service-broker/pull/42) ([tamalsaha](https://github.com/tamalsaha))
- Update references to KubeDB 0.10.0 [\#41](https://github.com/appscode/service-broker/pull/41) ([tamalsaha](https://github.com/tamalsaha))
- Pass Annotations to Operator PodTemplate [\#40](https://github.com/appscode/service-broker/pull/40) ([tamalsaha](https://github.com/tamalsaha))
- Update KubeDB client libraries to 0.10.0 [\#39](https://github.com/appscode/service-broker/pull/39) ([tamalsaha](https://github.com/tamalsaha))
- Add monitoring docs [\#37](https://github.com/appscode/service-broker/pull/37) ([hossainemruz](https://github.com/hossainemruz))
- Use k8s.io/apiserver to serve broker handlers [\#36](https://github.com/appscode/service-broker/pull/36) ([tamalsaha](https://github.com/tamalsaha))
- Update Docs and Tests [\#30](https://github.com/appscode/service-broker/pull/30) ([shudipta](https://github.com/shudipta))
- Various fixes [\#29](https://github.com/appscode/service-broker/pull/29) ([tamalsaha](https://github.com/tamalsaha))
- Add release script [\#28](https://github.com/appscode/service-broker/pull/28) ([tamalsaha](https://github.com/tamalsaha))
- Implement Bind using AppBinding [\#27](https://github.com/appscode/service-broker/pull/27) ([tamalsaha](https://github.com/tamalsaha))
- Use provision namespace for instances and change the instance label's key [\#26](https://github.com/appscode/service-broker/pull/26) ([shudipta](https://github.com/shudipta))
- Fix formatting string [\#25](https://github.com/appscode/service-broker/pull/25) ([tamalsaha](https://github.com/tamalsaha))
- Use chart name appscode-service-broker [\#24](https://github.com/appscode/service-broker/pull/24) ([tamalsaha](https://github.com/tamalsaha))
- Remove map\[\] for storing provision request info [\#23](https://github.com/appscode/service-broker/pull/23) ([shudipta](https://github.com/shudipta))
- Update Kubernetes client libraries to 1.12.0 [\#21](https://github.com/appscode/service-broker/pull/21) ([tamalsaha](https://github.com/tamalsaha))
- Fix installer script to usable from curl | bash [\#20](https://github.com/appscode/service-broker/pull/20) ([shudipta](https://github.com/shudipta))
- Update the flags with cobra flags [\#19](https://github.com/appscode/service-broker/pull/19) ([shudipta](https://github.com/shudipta))
- Update AppsCode Service Broker Issue [\#16](https://github.com/appscode/service-broker/pull/16) ([shudipta](https://github.com/shudipta))
- Update Service Broker for Kubedb-0.9.0 [\#13](https://github.com/appscode/service-broker/pull/13) ([shudipta](https://github.com/shudipta))
- Fix coverage.sh [\#12](https://github.com/appscode/service-broker/pull/12) ([tamalsaha](https://github.com/tamalsaha))
- Add travis.yaml [\#11](https://github.com/appscode/service-broker/pull/11) ([tamalsaha](https://github.com/tamalsaha))
- Use --pull flag with docker build \(\#20\) [\#10](https://github.com/appscode/service-broker/pull/10) ([tamalsaha](https://github.com/tamalsaha))
- fix github status [\#9](https://github.com/appscode/service-broker/pull/9) ([tahsinrahman](https://github.com/tahsinrahman))
- update pipeline [\#8](https://github.com/appscode/service-broker/pull/8) ([tahsinrahman](https://github.com/tahsinrahman))
- Add concourse tests [\#7](https://github.com/appscode/service-broker/pull/7) ([tahsinrahman](https://github.com/tahsinrahman))
- Add docs [\#6](https://github.com/appscode/service-broker/pull/6) ([shudipta](https://github.com/shudipta))
- Add default plans for MongoDB, Redis and Memcached [\#5](https://github.com/appscode/service-broker/pull/5) ([shudipta](https://github.com/shudipta))
- Add plan for PostgreSQL and Elasticsearch [\#4](https://github.com/appscode/service-broker/pull/4) ([shudipta](https://github.com/shudipta))
- Mysql broker for kubedb [\#3](https://github.com/appscode/service-broker/pull/3) ([shudipta](https://github.com/shudipta))
- Implement mysql osb-broker api [\#2](https://github.com/appscode/service-broker/pull/2) ([shudipta](https://github.com/shudipta))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*