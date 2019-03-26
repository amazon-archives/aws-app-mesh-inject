SHELL=/bin/bash -eo pipefail
.DEFAULT_GOAL := build
IMAGE_REGION=${shell aws configure get region}
REPO=${IMAGE_ACCOUNT}.dkr.ecr.${IMAGE_REGION}.amazonaws.com/amazon/aws-app-mesh-inject
VERSION=$(shell cat VERSION)
HASH=$(shell git log --pretty=format:'%H' -n 1)
IMAGE_TAG=${VERSION}
IMAGE_ACCOUNT=${shell aws sts get-caller-identity --query "Account" --output text}

#
# Test
#
.PHONY: test goveralls
test:
	go test ./...

goveralls:
	go test -coverprofile=coverage.out ./...
	${GOPATH}/bin/goveralls -coverprofile=coverage.out -service=travis-ci

#
# Build
#
.PHONY: build push hashtag buildpushhash buildhash pushhash
build:
	docker build --no-cache -t ${REPO}:${IMAGE_TAG} .

push:
	$(eval export IMAGE_ACCOUNT)
	$(eval export IMAGE_REGION)
	$(eval export REPO)
	$(eval export IMAGE_TAG)
	./hack/pushImage.sh

hashtag:
	$(eval export IMAGE_TAG=${HASH})

buildpushhash: | hashtag build push

buildhash: | hashtag build

pushhash: | hashtag push

#
# Appmesh inject deployment
#
.PHONY: deploydev deploydevhash deploy clean
# Uses the image from developer account
deploydev:
	$(eval export IMAGE_NAME=${REPO}:${IMAGE_TAG})
	$(eval export MESH_REGION)
	$(eval export MESH_NAME)
	./hack/deployInjector.sh

deploydevhash: | hashtag deploydev

# Uses the official image from EKS account.
deploy:
	$(eval export IMAGE_NAME=602401143452.dkr.ecr.us-west-2.amazonaws.com/amazon/aws-app-mesh-inject:v1.0.0)
	$(eval export MESH_REGION)
	$(eval export MESH_NAME)
	./hack/deployInjector.sh

clean:
	kubectl delete namespace appmesh-inject
	rm -rf ./_output

#
# Demo
#
.PHONY: k8sdemo appmeshdemo updatecolors cleandemo
k8sdemo:
	kubectl apply -f demo/ns.yaml
	kubectl apply -f demo/front-end.yaml
	kubectl apply -f demo/colors.yaml

cleank8sdemo:
	kubectl delete -f demo/ns.yaml

appmeshdemo:
	$(eval export MESH_NAME)
	cd demo/appmesh/ && \
	./deployappmesh.sh

updatecolors:
	cd demo/appmesh/ && \
	./updatecolors.sh

cleandemo:
	$(eval export MESH_NAME)
	kubectl delete -f demo/ns.yaml
	./demo/appmesh/cleanappmesh.sh

#
# ECR pull secrets
#
.PHONY: ecrsecrets nssecrets
ecrsecrets:
	$(eval export TOKEN=$(shell aws ecr get-authorization-token --region ${REGION} \
		--registry-ids 072792469044 \
		--output text --query 'authorizationData[].authorizationToken'| \
		base64 -D | \
		cut -d: -f2))
	kubectl delete secret --ignore-not-found inject-ecr-secret -n aws-app-mesh-inject
	@kubectl create secret docker-registry inject-ecr-secret -n aws-app-mesh-inject\
	 --docker-server=https://${REPO} \
	 --docker-username=AWS \
	 --docker-password="${TOKEN}" \
	 --docker-email="to-be@deprecated.com"
	kubectl patch deployment aws-app-mesh-inject -n aws-app-mesh-inject -p '$(shell cat ecr-secret-patch.json)'

nssecrets:
	$(eval export TOKEN=$(shell aws ecr get-authorization-token --region us-west-2 \
		--registry-ids 111345817488 \
		--output text --query 'authorizationData[].authorizationToken' | \
		base64 -D | \
		cut -d: -f2))
	kubectl delete secret --ignore-not-found appmesh-ecr-secret -n ${NAMESPACE}
	@kubectl create secret docker-registry appmesh-ecr-secret -n ${NAMESPACE} \
	 --docker-server=https://111345817488.dkr.ecr.us-west-2.amazonaws.com \
	 --docker-username=AWS \
	 --docker-password="${TOKEN}" \
	 --docker-email="to-be@deprecated.com"
