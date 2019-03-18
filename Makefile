SHELL=/bin/bash -eo pipefail
.DEFAULT_GOAL := build
MESH_REGION=$(shell aws configure get region)
IMAGE_REGION=${MESH_REGION}
REPO=${IMAGE_ACCOUNT}.dkr.ecr.${IMAGE_REGION}.amazonaws.com/amazon/aws-app-mesh-inject
VERSION=$(shell cat VERSION)
HASH=$(shell git log --pretty=format:'%H' -n 1)
IMAGE_TAG=${VERSION}

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
ifeq ($(IMAGE_ACCOUNT),)
    $(error IMAGE_ACCOUNT is not set)
endif
	aws ecr get-login --region ${IMAGE_REGION} --no-include-email | bash
	docker push ${REPO}:${IMAGE_TAG}

hashtag:
	$(eval export IMAGE_TAG=${HASH})

buildpushhash: | hashtag build push

buildhash: | hashtag build

pushhash: | hashtag push

#
# Appmesh inject deployment
#
.PHONY: deployk8s deployk8shash clean
deployk8s:
ifeq ($(MESH),)
    $(error MESH is not set)
else ifeq ($(IMAGE_ACCOUNT),)
    $(error IMAGE_ACCOUNT is not set)
endif
	$(eval export IMAGE_ACCOUNT)
	$(eval export IMAGE_TAG)
	$(eval export IMAGE_REGION)
	$(eval export MESH_REGION)
	$(eval export MESH)
	kubectl apply -f deploy/inject-ns.yaml
	./hack/gen-cert.sh
	./hack/ca-bundle.sh
	kubectl apply -f _output/inject.yaml

deployk8shash: | hashtag deployk8s

clean:
	kubectl delete -f _output/inject.yaml
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
	$(eval export MESH)
	cd demo/appmesh/ && \
	./deployappmesh.sh

updatecolors:
	cd demo/appmesh/ && \
	./updatecolors.sh

cleandemo:
	$(eval export MESH)
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
