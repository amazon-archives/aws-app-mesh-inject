# Install

We provide several options to install the sidecar injector.

 - [Helm](#helm)
 - [Kustomize and Cert-Manager Install](#kustomize-and-cert-manager-install)
 - [CLI Install](#cli-install)

[After Install](#after-install) instructions.
[Cleanup](#cleanup) instructions.

## Helm

Please reference the [Helm chart repository](https://github.com/aws/eks-charts) for install instructions.

## Kustomize and Cert-Manager Install
You can automatically provision certificates and configure your mesh with the following commands.

### Prerequisites
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [kustomize](https://kustomize.io/) (for kubectl versions less than 1.14)
### Install
Install [cert-manager](https://docs.cert-manager.io/en/latest/index.html)

```bash
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.10.1/cert-manager.yaml 
```

Wait for the cert-manager to deploy with the webhook.

```bash
kubectl get pods -n cert-manager
NAME                                       READY   STATUS    RESTARTS   AGE
cert-manager-6999fcbd4b-sd7wm              1/1     Running   0          20m
cert-manager-cainjector-6f5fb74459-8sdn4   1/1     Running   0          20m
cert-manager-webhook-54ccd98f74-w2gfh      1/1     Running   2          20m
```

Now create a local folder.
```bash
mkdir myappmeshconfig
```
Create two files in the folder:

kustomization.yaml
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
bases:
- github.com/aws/aws-app-mesh-inject//kustomize
patches:
- mymesh.yaml
```

mymesh.yaml
```yaml
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: appmesh-inject
  namespace: appmesh-system
spec:
  template:
    spec:
      containers:
        - name: webhook
          env:
            - name: APPMESH_REGION
              value: ## PUT YOUR REGION HERE
            - name: APPMESH_NAME
              value: ## PUT YOUR MESH NAME HERE
            - name: APPMESH_LOG_LEVEL ## (Optional, remove name and value for default "info")
              value: ## ENVOY LOG LEVEL
```
Now you can apply the kustomize manifests from within your kustomize folder.
```bash
kubectl apply -k .
```

## CLI Install

### Prerequisites
CLI Install:
* [openssl](https://www.openssl.org/source/)
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [jq](https://stedolan.github.io/jq/download/)

### Install
To deploy the sidecar injector you must export the name of your new mesh

```
$ export MESH_NAME=my_mesh_name
```

(Optional) To enable stats_tags on sidecar (Envoy) use
```
$ export ENABLE_STATS_TAGS=true
```

(Optional) If enabled, Envoy will emit DogStatsD metrics to 127.0.0.1:8125, where it expects to find a statsd receiver. This could be either a Datadog sidecar, or something like [statsd_exporter](https://github.com/prometheus/statsd_exporter) (see below).
```
$ export ENABLE_STATSD=true
```

(Optional) To deploy [statsd_exporter](https://github.com/prometheus/statsd_exporter) as a sidecar, which can recieve statsd metrics and republish them in Prometheus format. It listens for metrics on 127.0.0.1:8125, and exposes a prometheus endpoint at 127.0.0.1:9201. This is useful as statsd provides metrics around request latency (p50, p99 etc), whereas the standard Envoy prometheus endpoint does not.
```
$ export INJECT_STATSD_EXPORTER_SIDECAR=true
```

(Optional) To enable the xray-daemon sidecar injection use
```
$ export INJECT_XRAY_SIDECAR=true
```

(Optional) The appmesh injector needs a CA bundle to trust the webhooks coming from Kubernetes. The installation scripts will make a best-effort attempt at fetching it automatically, but this cannot be done in some cases.
The CA bundle can also be configured manually by setting a `CA_BUNDLE` environment variable to the content of the bundle.

```
$ export CA_BUNDLE=$(cat /path/to/ca-bundle | base64)
```

Now you can deploy the appmesh injector

### Option 1: clone the repository

```bash
$ make deploy
```

This will bootstrap the required certificates and start the sidecar injector in
your cluster.

To cleanup you can run
```
$ make clean
```

### Option 2: download and execute the install script

Specify the region to download sidecar injector. 
Please reference the [AWS Regional Table](https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/) for the supported regions

```
$ export MESH_REGION='us-east-1'
```
Download and execute the install script

```bash
curl https://raw.githubusercontent.com/aws/aws-app-mesh-inject/master/scripts/install.sh | bash
```

-------------------
## After Install

Wait for the sidecar injector to deploy.

```bash
kubectl get pods -n appmesh-system
NAME                                   READY   STATUS    RESTARTS   AGE
appmesh-inject-5bb846958c-j5v24   1/1     Running   0          24s
```

Return to [Under the hood](../README.md#under-the-hood) for more information on how to use the sidecar injector.

## Cleanup

To cleanup you can run
```
kubectl delete namespace appmesh-system; kubectl delete mutatingwebhookconfiguration appmesh-inject;
kubectl delete clusterrolebindings appmesh-inject; kubectl delete clusterrole appmesh-inject;
```

