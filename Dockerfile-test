FROM golang:1.11.0-alpine3.8

WORKDIR /go/src/github.com/AirHelp/rabbit-amazon-forwarder

RUN apk --no-cache add git gcc musl-dev && go get -u github.com/golang/dep/cmd/dep

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v -vendor-only

COPY . .
