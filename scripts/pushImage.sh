#!/usr/bin/env bash

[ -z "$IMAGE_ACCOUNT" ] && { echo "Need to set the environment variable IMAGE_ACCOUNT"; exit 1; }
[ -z "$IMAGE_REGION" ] && { echo "Need to set the environment variable IMAGE_REGION"; exit 1; }
[ -z "$REPO" ] && { echo "Need to set the environment variable REPO"; exit 1; }
[ -z "$IMAGE_TAG" ] && { echo "Need to set the environment variable IMAGE_TAG"; exit 1; }

aws ecr get-login --region ${IMAGE_REGION} --no-include-email | bash
docker push ${REPO}:${IMAGE_TAG}