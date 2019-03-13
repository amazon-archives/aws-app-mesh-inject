#!/bin/bash
set -x
aws appmesh update-route --mesh-name ${MESH} --cli-input-json file://colors.r.2.json
