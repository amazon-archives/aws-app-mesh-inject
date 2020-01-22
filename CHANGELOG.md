# CHANGELOG

## v0.4.0

### Summary
  This version supports EKS-Fargate pods and bumps the default envoy version to v1.12.2

### Changes:
* Copy attribution doc to image ([#118](https://github.com/aws/aws-app-mesh-inject/pull/118), @nckturner)
* Update envoy image version to v1.12.2.1-prod ([#115](https://github.com/aws/aws-app-mesh-inject/pull/115), @abaptiste)
* Revert "Update Envoy image to v1.12.2.0-prod" ([#113](https://github.com/aws/aws-app-mesh-inject/pull/113), @nckturner)
* Update Envoy image to v1.12.2.0-prod ([#112](https://github.com/aws/aws-app-mesh-inject/pull/112), abaptiste)
* Support for Amazon EKS on AWS Fargate ([#111](https://github.com/aws/aws-app-mesh-inject/pull/111), @kiranmeduri)
* Rename config dir to envoy-tracing-config ([#108](https://github.com/aws/aws-app-mesh-inject/pull/108), @nckturner)
* Update default app mesh envoy sidecar to v1.12.1.1 ([#107](https://github.com/aws/aws-app-mesh-inject/pull/107), @lavignes)
* Add AWS_REGION env-var to xray-daemon ([#106](https://github.com/aws/aws-app-mesh-inject/pull/106), @kiranmeduri)
* Propogate additional settings via Pod to CNI in container runtime ([#105](https://github.com/aws/aws-app-mesh-inject/pull/105), @kiranmeduri)
* Enable CNI to take over netfilter rules setup from init-container ([#104](https://github.com/aws/aws-app-mesh-inject/pull/104), @kiranmeduri)

## v0.3.1

### Summary

Patch release fixing a bug with the config dir location for injected tracing config for envoy.

### Changes

* Cherry pick of b69d6da: Rename config dir to envoy-tracing-config ([#108](https://github.com/aws/aws-app-mesh-inject/pull/108), @nckturner)

## v0.3.0

### Summary

This version bumps the default envoy version to 1.12.1, allows injection to be opt-in on a per pod basis.

WARNING: it changes the namespace used by the installation to be appmesh-system (same as the controller).

### Changes

* Update default app mesh envoy sidecar to v1.12.1 ([#101](https://github.com/aws/aws-app-mesh-inject/pull/101), @lavignes)
* Attribution document ([#102](https://github.com/aws/aws-app-mesh-inject/pull/102), @nckturner)
* Rename Envoy tracing config volume ([#100](https://github.com/aws/aws-app-mesh-inject/pull/100), @stefanprodan)
* Change namespace to appmesh-system ([#92](https://github.com/aws/aws-app-mesh-inject/pull/92), @jasonrichardsmith)
* Add default value for injection ([#75](https://github.com/aws/aws-app-mesh-inject/pull/75), @nilscan)
* Update default App Mesh Envoy sidecar image to latest release ([#99](https://github.com/aws/aws-app-mesh-inject/pull/99), @lavignes)
* Support selection of region for installing using CLI ([#98](https://github.com/aws/aws-app-mesh-inject/pull/98), @jefp)
* update README.md: add GoReport badge ([#95](https://github.com/aws/aws-app-mesh-inject/pull/95), @krlevkirill)
* Added kustomize deployment with cert-manager ([#90](https://github.com/aws/aws-app-mesh-inject/pull/90), @jasonrichardsmith)

## v0.2.0

### Summary

This version adds Jaeger and datadog tracing support, support for the App Mesh preview channel, and bumps the default envoy image to v1.11.1.1-prod.

### Changes

* Changing branch reference to master ([#84](https://github.com/aws/aws-app-mesh-inject/pull/84), @nckturner)
* Deprecated StatsD Exporter (fixes #82) ([#87](https://github.com/aws/aws-app-mesh-inject/pull/87), @PaulMaddox)
* Adding Datadog tracer to the Appmesh injector ([#85](https://github.com/aws/aws-app-mesh-inject/pull/85), @CharlyF)
* Add CircleCI config.yaml ([#86](https://github.com/aws/aws-app-mesh-inject/pull/86), @nckturner)
* Update default App Mesh Envoy sidecar image to latest release ([#83](https://github.com/aws/aws-app-mesh-inject/pull/83), @lavignes)
* Add support for Jaeger tracing ([#81](https://github.com/aws/aws-app-mesh-inject/pull/81), @stefanprodan)
* Add init container resources requests ([#61](https://github.com/aws/aws-app-mesh-inject/pull/61), @stefanprodan)
* Add preview channel option ([#80](https://github.com/aws/aws-app-mesh-inject/pull/80), @bcelenza)
* Update default Envoy image to v1.11.1.1-prod ([#74](https://github.com/aws/aws-app-mesh-inject/pull/74), @lavignes)

## v0.1.7

### Summary

This version bumps the envoy image to v1.11.2.0-prod.

### Changes

* Update default Envoy image to v1.11.2.0-prod ([#88](https://github.com/aws/aws-app-mesh-inject/pull/88), @nckturner)

## v0.1.6

### Summary

This version bumps the default envoy version to 1.11.1 and fixes install script bugs.

### Changes

* Release v0.1.6 ([#72](https://github.com/aws/aws-app-mesh-inject/pull/72), @nckturner)
* Read cpu and memory from annotations ([#67](https://github.com/aws/aws-app-mesh-inject/pull/67), @kiranmeduri)
* [SECURITY] Update default envoy version to 1.11.1 ([#70](https://github.com/aws/aws-app-mesh-inject/pull/70), @bcelenza)
* Only provide a single method of pulling CA_BUNDLE value. ([#64](https://github.com/aws/aws-app-mesh-inject/pull/64), @geremyCohen)
* Fix ENABLE_STATSD default value ([#56](https://github.com/aws/aws-app-mesh-inject/pull/56), @lethalpaga)
* Fetch CA bundle from requestheader-client-ca-file ([#57](https://github.com/aws/aws-app-mesh-inject/pull/57), @lethalpaga)
* Add an option to pass list of ignored ports in egress traffic ([#58](https://github.com/aws/aws-app-mesh-inject/pull/58), @midN)

## v0.1.5

### Summary

This version introduces statsD support.

### Changes

* Release v0.1.5 ([#55](https://github.com/aws/aws-app-mesh-inject/pull/55), @nckturner)
* Remove --no-cache from docker build ([#54](https://github.com/aws/aws-app-mesh-inject/pull/54), @nckturner)
* Add StatsD support ([#49](https://github.com/aws/aws-app-mesh-inject/pull/49), @PaulMaddox)

## v0.1.4

### Summary

This version introduces minor docs and install improvements and bumps the envoy image to 1.9.1.0.

### Changes

* Release v0.1.4 ([#52](https://github.com/aws/aws-app-mesh-inject/pull/52), @nckturner)
* Rename hack/ to scripts/ ([#51](https://github.com/aws/aws-app-mesh-inject/pull/51), @nckturner)
* Add security disclosure ([#50](https://github.com/aws/aws-app-mesh-inject/pull/50), @vipulsabhaya)
* Set the proxy route manager to the latest stable version ([#45](https://github.com/aws/aws-app-mesh-inject/pull/45), @bcelenza)
* Update all App Mesh Envoy References to 1.9.1.0 ([#46](https://github.com/aws/aws-app-mesh-inject/pull/46), @dastbe)

## v0.1.3

### Summary

This version did not introduce any changes.

## v0.1.2

### Summary

This version includes minor docs and install improvements.

### Changes

* Release v0.1.2 ([#43](https://github.com/aws/aws-app-mesh-inject/pull/43), @nckturner)
* Removed the sanity check of IMAGE_NAME ([#42](https://github.com/aws/aws-app-mesh-inject/pull/42), @jqmichael)
* removed demo ([#37](https://github.com/aws/aws-app-mesh-inject/pull/37), @jasonrichardsmith)

## v0.1.1

### Summary

This version introduces X-Ray integration, along with bumping the envoy sidecar version to v1.9.0 and adding support for multiple meshes.

### Changes

* Release v0.1.1 ([#40](https://github.com/aws/aws-app-mesh-inject/pull/40), @nckturner)
* Defaulted Envoy log level to "info" ([#35](https://github.com/aws/aws-app-mesh-inject/pull/35), @jqmichael)
* Add inject-xray-sidecar and enable-stats-tags to container args in in… ([#34](https://github.com/aws/aws-app-mesh-inject/pull/34), @kiranmeduri)
* Injecting X-Ray config and sidecar ([#30](https://github.com/aws/aws-app-mesh-inject/pull/30), kiranmeduri)
* Update Envoy to v1.9.0 ([#28](https://github.com/aws/aws-app-mesh-inject/pull/28), @stefanprodan)
* Add support for multiple meshes ([#24](https://github.com/aws/aws-app-mesh-inject/pull/24), @stefanprodan)
* Cleanup fully ([#26](https://github.com/aws/aws-app-mesh-inject/pull/26), @jqmichael)

## v0.1.0

### Summary

This is the initial release.  It implements a webhook that intercepts pod creations and injects an envoy sidecar in annotated namespaces.

### Changes


* Added install script ([#23](https://github.com/aws/aws-app-mesh-inject/pull/23), @jqmichael)
* Evalute env varaible in the webhook template instead of using sed ([#22](https://github.com/aws/aws-app-mesh-inject/pull/22), @jqmichael)
* Made env variable required within individual targets ([#20](https://github.com/aws/aws-app-mesh-inject/pull/20), @jqmichael)
* Rename imports and switch to klog ([#21](https://github.com/aws/aws-app-mesh-inject/pull/21), @stefanprodan)
* Refactoring of build ([#19](https://github.com/aws/aws-app-mesh-inject/pull/19), @nckturner)
* Demo fixes for updated api ([#18](https://github.com/aws/aws-app-mesh-inject/pull/18), @jasonrichardsmith)
* Refactor patch package ([#15](https://github.com/aws/aws-app-mesh-inject/pull/15), @stefanprodan)
* Make AppMesh region optional by reading from ec2 metadata service ([#14](https://github.com/aws/aws-app-mesh-inject/pull/14), @jqmichael)
* Webhook refactoring [part-1] ([#10](https://github.com/aws/aws-app-mesh-inject/pull/10), @stefanprodan)
* Eks integration ([#7](https://github.com/aws/aws-app-mesh-inject/pull/7), @jqmichael)
* Add Envoy stats port to patch ([#9](https://github.com/aws/aws-app-mesh-inject/pull/9), @stefanprodan)
* Initial webhook implementation (@jasonrichardsmith)
