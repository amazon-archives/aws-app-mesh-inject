[![CircleCI](https://circleci.com/gh/aws/aws-app-mesh-inject/tree/master.svg?style=svg)](https://circleci.com/gh/aws/aws-app-mesh-inject/tree/master)

# App Mesh Inject

The AWS App Mesh Kubernetes sidecar injecting Admission Controller.

## Security disclosures

If you think you’ve found a potential security issue, please do not post it in the Issues.  Instead, please follow the instructions [here](https://aws.amazon.com/security/vulnerability-reporting/) or [email AWS security directly](mailto:aws-security@amazon.com).

## Installation
Please reference the [install instructions](INSTALL.md).

### Warning
To align our helm repository and this repository we have changed the namespace to appmesh-system and resource names to appmesh-inject. 

## Under the hood
### Enable Sidecar injection

To enable sidecar injection for a namespace, you need to label the namespace with `appmesh.k8s.aws/sidecarInjectorWebhook=enabled`

```
kubectl label namespace appmesh-demo appmesh.k8s.aws/sidecarInjectorWebhook=enabled
```

### Default behavior and how to override

Sidecars will be injected to all new pods in the namespace that has enabled sidecar injector webhook. To disable injecting the sidecar
to particular pods in that namespace, add `appmesh.k8s.aws/sidecarInjectorWebhook: disabled` annotation to the pod spec.

All container ports defined in the pod spec will be passed to sidecars as application ports.
To override, add `appmesh.k8s.aws/ports: "<ports>"` annotation to the pod spec.

By default all egress traffic ports will be routed, except SSH.
To override, add `appmesh.k8s.aws/egressIgnoredPorts: "<ports>"` annotation to the pod spec. ( Comma separated list of ports for which egress traffic will be ignored )

The name of the controller that creates the pod will be used as virtual node name and pass over to the sidecar. For example, if a pod
is created by a deployment, the virtual node name will be `<deployment name>-<namespace>`.
To override, add `appmesh.k8s.aws/virtualNode: <virtual node name>` annotation to the pod spec.

The mesh name provided at install time can be overridden with the `appmesh.k8s.aws/mesh: <mesh name>` annotation at POD spec level.

For example:
```yaml
apiVersion: appsv1
kind: Deployment
metadata:
  labels:
    name: my-cool-deployment
spec:
  template:
    metadata:
      annotations:
        appmesh.k8s.aws/mesh: my-mesh
        appmesh.k8s.aws/ports: "8079,8080"
        appmesh.k8s.aws/egressIgnoredPorts: "22"
        appmesh.k8s.aws/virtualNode: my-app
        appmesh.k8s.aws/sidecarInjectorWebhook: disabled
```

To see an example on how to use this sidecar injector you can visit the [demo page](https://github.com/aws/aws-app-mesh-examples/tree/master/examples/). 

## Troubleshooting

### CA bundle not configured properly

If the CA bundle isn't configured properly, the pod will log the following log message:

```
TLS handshake error from 10.0.0.1:45390: remote error: tls: bad certificate
```

If this happens, set the `CA_BUNDLE` environment variable to the content of the CA bundle. Make sure that this value is base64 encoded (e.g. it shouldn't start with `-----BEGIN CERTIFICATE-----`).
