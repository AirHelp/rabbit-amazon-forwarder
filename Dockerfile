FROM alpine

RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*
RUN mkdir config && touch /config/mapping.json
ADD rabbit-amazon-forwarder /
CMD ["/rabbit-amazon-forwarder"]
