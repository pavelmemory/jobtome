FROM golang:1.14-alpine3.12 as builder

RUN apk add --no-cache --update \
    gcc \
    musl-dev \
    git \
    make

WORKDIR /jobtome

COPY cmd/ cmd/
COPY internal/ internal/
COPY go.mod go.mod
COPY go.sum go.sum
COPY Makefile Makefile

RUN make build

FROM alpine:3.12

COPY --from=builder /jobtome/build/bin/ /usr/local/bin/

RUN addgroup -g 1000 jobtome && \
    adduser -h /jobtome -D -u 1000 -G jobtome jobtome && \
    chown jobtome:jobtome /jobtome
USER jobtome

ENTRYPOINT jobtome
