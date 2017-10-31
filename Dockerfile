FROM golang:1.9.2-alpine3.6 AS golang-build
RUN mkdir -p /go/src/github.com/AirHelp/rabbit-amazon-forwarder
WORKDIR /go/src/github.com/AirHelp/rabbit-amazon-forwarder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rabbit-amazon-forwarder .

FROM alpine:3.6
RUN mkdir -p /config
RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=golang-build /go/src/github.com/AirHelp/rabbit-amazon-forwarder/rabbit-amazon-forwarder /
CMD ["/rabbit-amazon-forwarder"]
