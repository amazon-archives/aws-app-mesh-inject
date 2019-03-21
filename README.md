# AppMesh Inject

The AWS AppMesh Kubernetes sidecar injecting Admission Controller.

## Running
To run this sidecar injector or the demo you need both the [aws cli](https://aws.amazon.com/cli/)
 and [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/).

To deploy the sidecar injector you must have your aws cli configured with the region you will be working in.

To verify please run and verify you are configured for the correct region
```
$ aws configure get region
```
Export the name of your new mesh
```
$ export MESH=my_mesh_name
```
Now you can deploy the appmesh injector

```bash
$ make deployk8s
```

This will bootstrap the required certificates and start the sidecar injector in
your cluster.

To cleanup you can run
```
$ make clean
```

### Enable Sidecar injection

To enable sidecar injection for a namespace, you need to label the namespace with `appmesh.k8s.aws/sidecarInjectorWebhook=enabled`

```
kubectl label namespace appmesh-demo appmesh.k8s.aws/sidecarInjectorWebhook=enabled
```

### Default behavior and how to override

Sidecars will be injected to all pods in the namespace that has enabled sidecar injector webhook. To override, add 
`appmesh.k8s.aws/sidecarInjectorWebhook: disabled` annotation to the pod spec. 

All container ports defined in the pod spec will be passed to sidecars as application ports. 
To override, add `appmesh.k8s.aws/ports: "<ports>"` annotation to the pod spec. 

The name of the controller that eventually creates the pod will be used as virtual node name. For example, if a pod 
is created by a deployment, the virtual node name will be `<deployment name>-<namespace>`. 
To override, add `appmesh.k8s.aws/virtualNode: <virtual node name>` annotation to the pod spec. 

For example:
```yaml
    metadata:
      labels:
        app: my-app
      annotations:
        appmesh.k8s.aws/ports: "8079,8080"
        appmesh.k8s.aws/virtualNode: my-app
        appmesh.k8s.aws/sidecarInjectorWebhook: disabled
```

## Running the Demo

You can run the demo app by running
```
$ make k8sdemo
```

The sidecar injector should have injected sidecars into the deployments, so you should see something like this
Pods may be initing which means injection worked
```
$ kubectl get pods -n appmesh-demo
appmesh-demo     blue-866f865cc7-gbb7z             0/2       Init:0/1   0          3s
appmesh-demo     color-green-6b9db9948-lbbbn       0/2       Init:0/1   0          3s
appmesh-demo     color-orange-f78bfd8ff-snf5q      0/2       Init:0/1   0          3s
appmesh-demo     front-end-54f69dfd7b-zjgss        0/2       Init:0/1   0          4s
```
or init completed
```
$ kubectl get pods -n appmesh-demo
NAME                           READY     STATUS    RESTARTS   AGE
blue-866f865cc7-tkfkv          2/2       Running   0          5s
color-green-6b9db9948-c4qx8    2/2       Running   0          5s
color-orange-f78bfd8ff-chh56   2/2       Running   0          5s
front-end-54f69dfd7b-7qtbh     2/2       Running   0          5s
```

To view the demo webpage run
```
$ kubectl port-forward -n appmesh-demo svc/front-end 8000:80
```
and visit http://localhost:8000/

You should see a lot of red requests

![demo screenshot1](img/screenshot1.png)

The mesh need to be made aware of your pods and how to route them, so you need to run

```
$ make appmeshdemo
```

After a few minutes the demo front-end should switch from all red to around 50% green and 50% blue.

![demo screenshot2](img/screenshot2.png)

This routing is based on demo/appmesh/colors.r.json
```
$ cat demo/appmesh/colors.r.json
{
    "routeName": "colors-route",
    "spec": {
        "httpRoute": {
            "action": {
                "weightedTargets": [
                    {
                        "virtualNode": "orange",
                        "weight": 0
                    },
                    {
                        "virtualNode": "blue",
                        "weight": 5
                    },
                    {
                        "virtualNode": "green",
                        "weight": 5
                    }
                ]
            },
            "match": {
                "prefix": "/"
            }
        }
    },
    "virtualRouterName": "colors"
}
```

You can adjust the weights in this file and then run
```
$ make updatecolors
```

And you should see the traffic distributed evenly across the values you set in the router.

You can clean up the entire demo by running
```
$ make cleandemo
```


## Running Outside of AWS

The AppMesh containers and the AppMesh Inject Admission Controller are hosted
on ECR, but are publically available.  To pull the containers outside of AWS
requires that Kubernetes contain ImagePullSecrets.
The follow steps requores the awscli.
For testing purposes you can add ImagePullSecrets by running:

```
$ make ecrsecrets
```

Then you will need to inject the pull secrets for the sidecars, for each
namespace that will run the AppMesh sidecars.

```
$ NAMESPACE=my-namespace; make nssecrets
```

Replacing "my-namespace" with the namespace you want to add pull secrets to.
