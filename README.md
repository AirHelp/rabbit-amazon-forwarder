# RabbitMQ -> Amazon forwader

Application to forward messages from RabbitMQ to different Amazon services.

## Architecture

![Alt text](img/rabbit-amazon-forwarder.png?raw=true "RabbitMQ -> Amazon architecture")

## Configuration

### Environment variables

Export environment variables:
```bash
export MAPPING_FILE=/config/mapping.json
export AWS_REGION=region
export AWS_ACCESS_KEY_ID=access_key
export AWS_SECRET_ACCESS_KEY=secret_key
```

### Mapping file

Definition of forwarder->consumer pairs should be placed inside mapping file.
Sample of RabbitMQ -> SNS mapping file. All fields are required.
```json
[
  {
    "source" : {
      "type" : "RabbitMQ",
      "name" : "test-rabbit",
      "connection" : "amqp://guest:guest@localhost:5672/",
      "topic" : "amq.topic",
      "queue" : "test-queue",
      "routing" : "#"
    },
    "destination" : {
      "type" : "SNS",
      "name" : "test-sns",
      "target" : "arn:aws:sns:eu-west-1:XXXXXXXX:test-forwarder"
    }
  }
]
```
Samples are located in `examples` directory.

## Build docker image

```bash
make release
```

## Run

Using docker:
```bash

```


Using docker-compose:
```bash
docker-compose up
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
