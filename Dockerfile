ARG GOVERSION=alpine
FROM golang:${GOVERSION} AS Builder

ARG CONSUL_ZEROCONF_VERSION=0.5.0

WORKDIR /src

RUN apk add git

COPY . .

RUN go mod download

RUN \
  CGO_ENABLED="0" \
  GO111MODULE=on \
  ./build $CONSUL_ZEROCONF_VERSION

FROM alpine:3.12

ARG UID=100
ARG GID=1000

LABEL org.opencontainers.image.authors="Tyler Andersen <tyler@redserenity.com>"
LABEL org.opencontainers.image.version=$CONSUL_ZEROCONF_VERSION

RUN apk add --no-cache ca-certificates dumb-init

RUN addgroup -g ${GID} consul-zeroconf && \
    adduser -u ${UID} -S -G consul-zeroconf consul-zeroconf

RUN mkdir -p /consul-zeroconf/config && \
    mkdir -p /consul-zeroconf/server && \
    mkdir -p /consul-zeroconf/nodes && \
    chown -R consul-zeroconf:consul-zeroconf /consul-zeroconf

COPY --from=Builder /src/consul-zeroconf /usr/bin/consul-zeroconf

VOLUME /consul-zeroconf/config

USER ${UID}:${GID}

# Run consul-zeroconf by default
CMD ["/bin/consul-zeroconf"]
