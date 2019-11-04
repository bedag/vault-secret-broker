ARG GO_VERSION=1.13
ARG DOCKER_REGISTRY=""

##### BUILD STAGE
FROM ${DOCKER_REGISTRY}golang:${GO_VERSION}-alpine AS builder

ENV GO111MODULE=on
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/vault-secret-broker

##### IMAGE STAGE
FROM ${DOCKER_REGISTRY}alpine:3.9

LABEL maintainer="syx.bbz@bedag.ch"

ENV LANG=en_US.UTF-8
ENV LC_ALL=en_US.UTF-8

RUN apk add --no-cache tini

RUN addgroup -S broker && adduser -S -g broker broker
USER broker

COPY --from=builder build/vault-secret-broker /

ENTRYPOINT ["/sbin/tini", "--", "/vault-secret-broker"]
CMD ["serve"]

EXPOSE 8080/tcp
EXPOSE 8443/tcp