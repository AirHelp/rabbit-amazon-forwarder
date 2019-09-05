# RabbitMQ -> Amazon forwarder

[![Go Report Card](https://goreportcard.com/badge/github.com/AirHelp/rabbit-amazon-forwarder)](https://goreportcard.com/report/github.com/AirHelp/rabbit-amazon-forwarder)

Application to forward messages from RabbitMQ to different Amazon services.

Key features:

* forwarding RabbitMQ message to AWS SNS topic
* forwarding RabbitMQ message to AWS SNS queue
* triggering AWS lambda function directly from RabbitMQ message
* automatic RabbitMQ reconnect
* message delivery assurance based on RabbitMQ persistency and AWS error handling
* dedicated dead-letter exchange and queue creation
* http health checks and restart functionality

## Architecture

![Alt text](img/rabbit-amazon-forwarder.png?raw=true "RabbitMQ -> Amazon architecture")

## Configuration

The list of RabbitMQ sources and corresponding AWS target resources are stored in mapping file.

### Mapping file

Sample of RabbitMQ -> SNS mapping file. All fields are required. Samples are located in [examples](https://github.com/AirHelp/rabbit-amazon-forwarder/tree/master/examples) directory.
```json
[
  {
    "source" : {
      "type" : "RabbitMQ",
      "name" : "test-rabbit",
      "connection" : "amqp://guest:guest@localhost:5672/",
      "topic" : "amq.topic",
      "queue" : "test-queue",
      "routingKeys" : ["#"]
    },
    "destination" : {
      "type" : "SNS",
      "name" : "test-sns",
      "target" : "arn:aws:sns:eu-west-1:XXXXXXXX:test-forwarder"
    }
  }
]
```

### Environment variables

Forwarder uses the following environment variables:
```bash
export MAPPING_FILE=/config/mapping.json
export AWS_REGION=region
export AWS_ACCESS_KEY_ID=access_key
export AWS_SECRET_ACCESS_KEY=secret_key
```

#### Using TLS with rabbit

Specify amqps for the rabbit connection ub the mapping file:
```
 "connection" : "amqps://guest:guest@localhost:5671/",
```

Additional environment variables for working with TLS and rabbit:
```
export CA_CERT=/certs/ca_certificate.pem
export CERT_FILE=/certs/client_certificate.pem
export KEY_FILE=/certs/client_key.pem
```

### Amazon configuration

When making subscription to SNS -> SQS/HTTP/HTTPS set `Raw message delivery` to ensure that json messages are not escaped.

## Build docker image

```bash
make build
```

## Run

Using docker:
```bash
docker run \
-e AWS_REGION=$AWS_REGION \
-e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
-e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
-e MAPPING_FILE=/config/mapping.json \
-v $MAPPING_FILE:/config/mapping.json \
-p 8080:8080 \
airhelp/rabbit-amazon-forwarder
```

Using docker-compose:
```bash
docker-compose up
```

## Test
```
docker-compose build --pull
docker-compose run --rm tests
```

# Release

```bash
make push
docker tag airhelp/rabbit-amazon-forwarder airhelp/rabbit-amazon-forwarder:$VERSION
docker push airhelp/rabbit-amazon-forwarder:$VERSION
```

## Supervisor

Supervisor is a module which starts the consumer->forwarder pairs.
Exposed endpoints:
- `APP_URL/health` - returns status if all consumers are running
- `APP_URL/restart` - restarts all consumer->forwarder pairs
