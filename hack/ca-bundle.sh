#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

ROOT=$(cd $(dirname $0)/../; pwd)

export CA_BUNDLE=$(kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 | tr -d '\n')

if [[ -z $CA_BUNDLE ]]; then
    export CA_BUNDLE=$(kubectl config view --raw -o json --minify | jq -r '.clusters[0].cluster."certificate-authority-data"' | tr -d '"')
fi

mkdir -p _output/
sed \
    -e "s/{{IMAGE_ACCOUNT}}/${IMAGE_ACCOUNT}/g" \
    -e "s/{{IMAGE_REGION}}/${IMAGE_REGION}/g" \
    -e "s/{{IMAGE_TAG}}/${IMAGE_TAG}/g" \
    -e "s/{{MESH_REGION}}/${MESH_REGION}/g" \
    -e "s/{{MESH}}/${MESH_NAME}/g" \
    -e "s/{{CA_BUNDLE}}/${CA_BUNDLE}/g" \
    ${ROOT}/deploy/inject.yaml.template \
    > ${ROOT}/_output/inject.yaml

echo "Created injector manifest at:${ROOT}/_output/inject.yaml"
