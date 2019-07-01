#!/usr/bin/env bash
REPO=${REPO:-aws/aws-app-mesh-inject}
VERSION=${VERSION:-$(curl https://raw.githubusercontent.com/$REPO/master/VERSION)}

set -e

[ -z "$MESH_NAME" ] && { echo "Need to set the environment variable MESH_NAME"; exit 1; }

if ! command -v jq >/dev/null 2>&1; then
    echo "Please install jq before continuing"
    exit 1
fi

if ! command -v openssl >/dev/null 2>&1; then
    echo "Please install openssl before continuing"
    exit 1
fi

if ! command -v kubectl >/dev/null 2>&1; then
    echo "Please install kubectl before continuing"
    exit 1
fi

tmpdir=$(mktemp -d)
echo "\nWorking directory at ${tmpdir}\n"

cd $tmpdir

mkdir -p deploy
curl https://raw.githubusercontent.com/${REPO}/${VERSION}/deploy/inject-ns.yaml > deploy/inject-ns.yaml
curl https://raw.githubusercontent.com/${REPO}/${VERSION}/deploy/inject.yaml.template > deploy/inject.yaml.template

mkdir -p scripts
curl https://raw.githubusercontent.com/${REPO}/${VERSION}/scripts/gen-cert.sh > scripts/gen-cert.sh
curl https://raw.githubusercontent.com/${REPO}/${VERSION}/scripts/ca-bundle.sh > scripts/ca-bundle.sh

chmod u+x ./scripts/ca-bundle.sh ./scripts/gen-cert.sh

curl https://raw.githubusercontent.com/${REPO}/${VERSION}/scripts/deployInjector.sh | bash
