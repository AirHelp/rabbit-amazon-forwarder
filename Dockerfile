#syntax = docker/dockerfile:1.0-experimental

FROM golang:1.13-alpine AS builder

RUN apk add --no-cache curl git openssh \
 && adduser -D -g '' appuser

COPY . /go/src/github.com/AirHelp/rabbit-amazon-forwarder
WORKDIR /go/src/github.com/AirHelp/rabbit-amazon-forwarder

ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN  go mod tidy \
     && go mod verify \
     && go mod vendor

RUN go build -ldflags="-w -s" -o /go/src/github.com/AirHelp/rabbit-amazon-forwarder/forwarder

FROM alpine

RUN adduser -D -g '' appuser

WORKDIR /app

ENV ENVIRONMENT="dev"

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/AirHelp/rabbit-amazon-forwarder/forwarder /app/forwarder
COPY ./config /app/config

USER appuser

ENTRYPOINT ["/app/forwarder"]