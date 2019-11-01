FROM alpine:3.9

LABEL maintainer="syx.bbz@bedag.ch"

ENV LANG=en_US.UTF-8
ENV LC_ALL=en_US.UTF-8

RUN apk add --no-cache tini

RUN addgroup -S broker && adduser -S -g broker broker
USER broker

COPY vault-secret-broker /

ENTRYPOINT ["/sbin/tini", "--", "/vault-secret-broker"]
CMD ["serve"]

EXPOSE 8080/tcp
EXPOSE 8443/tcp