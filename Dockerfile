# build stage
FROM golang:1.12-alpine AS build-env

RUN apk add git

RUN mkdir -p /go/src/github.com/aws/aws-app-mesh-inject
WORKDIR /go/src/github.com/aws/aws-app-mesh-inject

ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY  . .
RUN adduser -D -u 10001 webhook
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' \
    -o appmeshinject ./cmd/app-mesh-inject/*.go

FROM scratch
COPY --from=build-env /go/src/github.com/aws/aws-app-mesh-inject/appmeshinject .
COPY --from=build-env /etc/passwd /etc/passwd
USER webhook
ENTRYPOINT ["/appmeshinject"]
