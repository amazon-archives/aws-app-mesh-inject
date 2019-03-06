# build stage
FROM golang:1.10-stretch AS build-env
RUN mkdir -p /go/src/github.com/awslabs/aws-app-mesh-inject
WORKDIR /go/src/github.com/awslabs/aws-app-mesh-inject
COPY  . .
RUN useradd -u 10001 webhook
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o appmeshinject

FROM scratch
COPY --from=build-env /go/src/github.com/awslabs/aws-app-mesh-inject/appmeshinject .
COPY --from=build-env /etc/passwd /etc/passwd
USER webhook
ENTRYPOINT ["/appmeshinject"]
