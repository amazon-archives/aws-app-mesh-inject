#!/usr/bin/env bash

set -ue

[ -z "$MESH_NAME" ] && { echo "Need to set the environment variable MESH_NAME"; exit 1; }

if ! command -v jq >/dev/null 2>&1; then
    echo "Please install jq before continue"
    exit 1
fi

if ! command -v openssl >/dev/null 2>&1; then
    echo "Please install openssl before continue"
    exit 1
fi

if ! command -v kubectl >/dev/null 2>&1; then
    echo "Please install kubectl before continue"
    exit 1
fi

kubectl apply -f deploy/inject-ns.yaml
./hack/gen-cert.sh
./hack/ca-bundle.sh
kubectl apply -f _output/inject.yaml
echo "waiting for aws-app-mesh-inject to start"
kubectl rollout status deployment/aws-app-mesh-inject -n appmesh-inject

ACTUAL_APPMESH_NAME=$(kubectl get deployment aws-app-mesh-inject -n appmesh-inject -o=jsonpath="{.spec.template.spec.containers[0].env[?(@.name=='APPMESH_NAME')].value}")
if [[ "$ACTUAL_APPMESH_NAME" = "$MESH_NAME" ]]; then
    echo "Mesh name has been set up"
else
    echo "Mesh name is unexpected. Expect:${MESH_NAME}, Actual:${ACTUAL_APPMESH_NAME}"
    exit 1
fi

echo "The injector is ready"