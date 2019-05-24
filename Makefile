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
	docker build -t ${REPO}:${IMAGE_TAG} .

push:
	$(eval export IMAGE_ACCOUNT)
	$(eval export IMAGE_REGION)
	$(eval export REPO)
	$(eval export IMAGE_TAG)
	./scripts/pushImage.sh

hashtag:
	$(eval export IMAGE_TAG=${HASH})

buildpushhash: | hashtag build push

buildhash: | hashtag build

pushhash: | hashtag push

ci-test-build:
	go test ./...
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o appmeshinject ./cmd/app-mesh-inject/*.go

#
# Appmesh inject deployment
#
.PHONY: deploydev deploydevhash deploy clean
# Uses the image from developer account
deploydev:
	$(eval export IMAGE_NAME=${REPO}:${IMAGE_TAG})
	$(eval export MESH_NAME)
	./scripts/deployInjector.sh

deploydevhash: | hashtag deploydev

# Uses the official image from EKS account.
deploy:
	$(eval export MESH_NAME)
	./scripts/deployInjector.sh

clean:
	kubectl delete namespace appmesh-inject
	kubectl delete mutatingwebhookconfiguration aws-app-mesh-inject
	kubectl delete clusterrolebindings aws-app-mesh-inject-binding
	kubectl delete clusterrole aws-app-mesh-inject-cr;
	rm -rf ./_output

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
