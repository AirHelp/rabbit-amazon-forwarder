FROM golang:1.18-alpine AS golang-build

RUN mkdir -p /go/src/github.com/AirHelp/rabbit-amazon-forwarder
WORKDIR /go/src/github.com/AirHelp/rabbit-amazon-forwarder

RUN apk --no-cache add git

COPY . .
RUN go mod tidy -go=1.18 -compat=1.18

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rabbit-amazon-forwarder .

FROM alpine

RUN mkdir -p /config
RUN mkdir -p /certs
RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY --from=golang-build /go/src/github.com/AirHelp/rabbit-amazon-forwarder/rabbit-amazon-forwarder /
CMD ["/rabbit-amazon-forwarder"]
