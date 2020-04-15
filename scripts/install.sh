#!/usr/bin/env bash
#
# This install script is used by the tutorial in the AWS docs tutorial on using
# App Mesh with Kubernetes:
#
# https://docs.aws.amazon.com/eks/latest/userguide/mesh-k8s-integration.html
#
# Do not modify it without testing the steps from the tutorial.  The default
# branch and version below will be installed by users following the tutorial,
# so the version should be updated with each release.
#


REPO=${REPO:-aws/aws-app-mesh-inject}
BRANCH=${BRANCH:-master}
export MESH_REGION=${MESH_REGION:-us-west-2}
export VERSION=${VERSION:-v0.4.1}

set -e
set -o pipefail

echo "Using AWS region ${MESH_REGION}"
echo "Fetching ${VERSION} from https://github.com/${REPO}/tree/${BRANCH}"

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

curl https://raw.githubusercontent.com/${REPO}/${BRANCH}/deploy/inject-ns.yaml > deploy/inject-ns.yaml
curl https://raw.githubusercontent.com/${REPO}/${BRANCH}/deploy/inject.yaml.template > deploy/inject.yaml.template

mkdir -p scripts
curl https://raw.githubusercontent.com/${REPO}/${BRANCH}/scripts/gen-cert.sh > scripts/gen-cert.sh
curl https://raw.githubusercontent.com/${REPO}/${BRANCH}/scripts/gen-inject-yaml.sh > scripts/gen-inject-yaml.sh

chmod u+x ./scripts/gen-inject-yaml.sh ./scripts/gen-cert.sh

curl https://raw.githubusercontent.com/${REPO}/${BRANCH}/scripts/deployInjector.sh | bash
