# CHANGELOG

## v0.2.0

### Summary

This version adds Jaeger and datadog tracing support, support for the App Mesh preview channel, and bumps the default envoy image to v1.11.1.1-prod.

### Changes

* Changing branch reference to master (https://github.com/aws/aws-app-mesh-inject/pull/84, @nckturner)
* Deprecated StatsD Exporter (fixes #82) (https://github.com/aws/aws-app-mesh-inject/pull/87, @PaulMaddox)
* Adding Datadog tracer to the Appmesh injector (https://github.com/aws/aws-app-mesh-inject/pull/85, @CharlyF)
* Add CircleCI config.yaml (https://github.com/aws/aws-app-mesh-inject/pull/86, @nckturner)
* Update default App Mesh Envoy sidecar image to latest release (https://github.com/aws/aws-app-mesh-inject/pull/83, @lavignes)
* Add support for Jaeger tracing (https://github.com/aws/aws-app-mesh-inject/pull/81, @stefanprodan)
* Add init container resources requests (https://github.com/aws/aws-app-mesh-inject/pull/61, @stefanprodan)
* Add preview channel option (https://github.com/aws/aws-app-mesh-inject/pull/80, @bcelenza)
* Update default Envoy image to v1.11.1.1-prod (https://github.com/aws/aws-app-mesh-inject/pull/74, @lavignes)

## v0.1.7

### Summary

This version bumps the envoy image to v1.11.2.0-prod.

### Changes

* Update default Envoy image to v1.11.2.0-prod (https://github.com/aws/aws-app-mesh-inject/pull/88, @nckturner)

## v0.1.6

### Summary

This version bumps the default envoy version to 1.11.1 and fixes install script bugs.

### Changes

* Release v0.1.6 (https://github.com/aws/aws-app-mesh-inject/pull/72, @nckturner)
* Read cpu and memory from annotations (https://github.com/aws/aws-app-mesh-inject/pull/67, @kiranmeduri)
* [SECURITY] Update default envoy version to 1.11.1 (https://github.com/aws/aws-app-mesh-inject/pull/70, @bcelenza)
* Only provide a single method of pulling CA_BUNDLE value. (https://github.com/aws/aws-app-mesh-inject/pull/64, @geremyCohen)
* Fix ENABLE_STATSD default value (https://github.com/aws/aws-app-mesh-inject/pull/56, @lethalpaga)
* Fetch CA bundle from requestheader-client-ca-file (https://github.com/aws/aws-app-mesh-inject/pull/57, @lethalpaga)
* Add an option to pass list of ignored ports in egress traffic (https://github.com/aws/aws-app-mesh-inject/pull/58, @midN)

## v0.1.5

### Summary

This version introduces statsD support.

### Changes

* Release v0.1.5 (https://github.com/aws/aws-app-mesh-inject/pull/55, @nckturner)
* Remove --no-cache from docker build (https://github.com/aws/aws-app-mesh-inject/pull/54, @nckturner)
* Add StatsD support (https://github.com/aws/aws-app-mesh-inject/pull/49, @PaulMaddox)

## v0.1.4

### Summary

This version introduces minor docs and install improvements and bumps the envoy image to 1.9.1.0.

### Changes

* Release v0.1.4 (https://github.com/aws/aws-app-mesh-inject/pull/52, @nckturner)
* Rename hack/ to scripts/ (https://github.com/aws/aws-app-mesh-inject/pull/51, @nckturner)
* Add security disclosure (https://github.com/aws/aws-app-mesh-inject/pull/50, @vipulsabhaya)
* Set the proxy route manager to the latest stable version (https://github.com/aws/aws-app-mesh-inject/pull/45, @bcelenza)
* Update all App Mesh Envoy References to 1.9.1.0 (https://github.com/aws/aws-app-mesh-inject/pull/46, @dastbe)

## v0.1.3

### Summary

This version did not introduce any changes.

## v0.1.2

### Summary

This version includes minor docs and install improvements.

### Changes

* Release v0.1.2 (https://github.com/aws/aws-app-mesh-inject/pull/43, @nckturner)
* Removed the sanity check of IMAGE_NAME (https://github.com/aws/aws-app-mesh-inject/pull/42, @jqmichael)
* removed demo (https://github.com/aws/aws-app-mesh-inject/pull/37, @jasonrichardsmith)

## v0.1.1

### Summary

This version introduces X-Ray integration, along with bumping the envoy sidecar version to v1.9.0 and adding support for multiple meshes.

### Changes

* Release v0.1.1 (https://github.com/aws/aws-app-mesh-inject/pull/40, @nckturner)
* Defaulted Envoy log level to "info" (https://github.com/aws/aws-app-mesh-inject/pull/35, @jqmichael)
* Add inject-xray-sidecar and enable-stats-tags to container args in in… (https://github.com/aws/aws-app-mesh-inject/pull/34, @kiranmeduri)
* Injecting X-Ray config and sidecar (https://github.com/aws/aws-app-mesh-inject/pull/30, kiranmeduri)
* Update Envoy to v1.9.0 (https://github.com/aws/aws-app-mesh-inject/pull/28, @stefanprodan)
* Add support for multiple meshes (https://github.com/aws/aws-app-mesh-inject/pull/24, @stefanprodan)
* Cleanup fully (https://github.com/aws/aws-app-mesh-inject/pull/26, @jqmichael)

## v0.1.0

### Summary

This is the initial release.  It implements a webhook that intercepts pod creations and injects an envoy sidecar in annotated namespaces.

### Changes


* Added install script (https://github.com/aws/aws-app-mesh-inject/pull/23, @jqmichael)
* Evalute env varaible in the webhook template instead of using sed (https://github.com/aws/aws-app-mesh-inject/pull/22, @jqmichael)
* Made env variable required within individual targets (https://github.com/aws/aws-app-mesh-inject/pull/20, @jqmichael)
* Rename imports and switch to klog (https://github.com/aws/aws-app-mesh-inject/pull/21, @stefanprodan)
* Refactoring of build (https://github.com/aws/aws-app-mesh-inject/pull/19, @nckturner)
* Demo fixes for updated api (https://github.com/aws/aws-app-mesh-inject/pull/18, @jasonrichardsmith)
* Refactor patch package (https://github.com/aws/aws-app-mesh-inject/pull/15, @stefanprodan)
* Make AppMesh region optional by reading from ec2 metadata service (https://github.com/aws/aws-app-mesh-inject/pull/14, @jqmichael)
* Webhook refactoring [part-1] (https://github.com/aws/aws-app-mesh-inject/pull/10, @stefanprodan)
* Eks integration (https://github.com/aws/aws-app-mesh-inject/pull/7, @jqmichael)
* Add Envoy stats port to patch (https://github.com/aws/aws-app-mesh-inject/pull/9, @stefanprodan)
* Initial webhook implementation (@jasonrichardsmith)
