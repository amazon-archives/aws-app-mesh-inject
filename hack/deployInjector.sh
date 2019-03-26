#!/usr/bin/env bash

set -ue

[ -z "$MESH_NAME" ] && { echo "Need to set the environment variable MESH_NAME"; exit 1; }
[ -z "$IMAGE_NAME" ] && { echo "Need to set the environment variable IMAGE_NAME"; exit 1; }

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