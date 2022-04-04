FROM golang:1.11.0-alpine3.8 AS golang-build

RUN mkdir -p /go/src/github.com/symopsio/rabbit-amazon-forwarder
WORKDIR /go/src/github.com/symopsio/rabbit-amazon-forwarder

RUN apk --no-cache add git && go get -u github.com/golang/dep/cmd/dep

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v -vendor-only

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rabbit-amazon-forwarder .

FROM alpine:3.8

RUN mkdir -p /config
RUN mkdir -p /certs
RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY --from=golang-build /go/src/github.com/symopsio/rabbit-amazon-forwarder/rabbit-amazon-forwarder /
CMD ["/rabbit-amazon-forwarder"]
