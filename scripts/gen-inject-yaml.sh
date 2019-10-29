#!/usr/bin/env bash
#
# This script generates the injector's Kubernetes yaml from a template.  The template supports the following
# environment variables:
# 
# IMAGE_NAME (optional): The entire image repository name and tag, formatted like <repo>:<tag>
# VERSION (optional, required if IMAGE_NAME is not set): The version of the official inject container, used as the image tag.  
#   Only used and required if IMAGE_NAME is not provided.
# MESH_REGION (optional): The region that the injected pods will run in and query the App Mesh API.
# MESH_NAME (required): The name of the mesh that injected pods should use.
# ENVOY_LOG_LEVEL (optional): Log level for envoy proxy sidecar logs.
# SIDECAR_IMAGE (optional): Envoy proxy sidecar container image.
# INIT_IMAGE (optional): Init container image.
# INJECT_XRAY_SIDECAR (optional): Boolean inject the sidecar to enable XRay tracing.
# ENABLE_STATS_TAGS (optional): Enables the use of App Mesh defined tags appmesh.mesh and appmesh.virtual_node. For more 
#   information, see config.metrics.v2.TagSpecifier in the Envoy documentation.
# ENABLE_STATSD (optional): Enables DogStatsD stats using 127.0.0.1:8125 as the default daemon endpoint.
# SIDECAR_CPU_REQUESTS (optional): Envoy proxy CPU request.
# SIDECAR_MEMORY_REQUESTS (optional): Envoy proxy memory request.
# CABUNDLE (optional): Base64 encoded cluster CA certificate.
#

set -e
set -o pipefail

ROOT=$(cd $(dirname $0)/../; pwd)

if [[ -z ${CA_BUNDLE:-} ]]; then
    export CA_BUNDLE=$(kubectl config view --raw -o json --minify | jq -r '.clusters[0].cluster."certificate-authority-data"' | tr -d '"')
fi

mkdir -p _output/

echo "processing templates"
eval "cat <<EOF
$(<${ROOT}/deploy/inject.yaml.template)
EOF
" > ${ROOT}/_output/inject.yaml

echo "Created injector manifest at:${ROOT}/_output/inject.yaml"
