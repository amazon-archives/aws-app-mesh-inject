#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

ROOT=$(cd $(dirname $0)/../../; pwd)

export CA_BUNDLE=$(kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 | tr -d '\n')

if [[ -z $CA_BUNDLE ]]; then
	cc=$(kubectl config view --raw --flatten -o json | jq -r '.contexts[] | select(.name == "'$(kubectl config current-context)'") | .context.cluster')
	export CA_BUNDLE=$(kubectl config view --raw --flatten -o json | jq -r '.clusters[] | select(.name == "'${cc}'") | .cluster."certificate-authority-data"')
fi

cat manifest.yaml | envsubst > manifest-ca.yaml
