# Kustomize and Cert-Manager Install
You can automatically provision certificates and configure your mesh with the following commands.

## Prerequisites
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
For kubectl versions less than 1.14
* [kustomize](https://kustomize.io/)


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
- github.com/aws/aws-app-mesh-inject
patches:
- deployment.yaml
```

mymesh.yaml
```yaml

---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: aws-app-mesh-inject
  namespace: appmesh-inject
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

Wait for the sidecar injector to deploy.

```bash
kubectl get pods -n appmesh-inject
NAME                                   READY   STATUS    RESTARTS   AGE
aws-app-mesh-inject-5bb846958c-j5v24   1/1     Running   0          24s
```

Return to [Under the hood](../README.md#under-the-hood) for more information on how to use the sidecar injector.
