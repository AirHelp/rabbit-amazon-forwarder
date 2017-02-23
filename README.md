# RabbitMQ -> Amazon forwader

Application to forward messages from RabbitMQ to different Amazon services.

# Configuration

Export environment variables:
```bash
export MAPPING_FILE=/config/mapping.json
export AWS_REGION=region
export AWS_ACCESS_KEY_ID=access_key
export AWS_SECRET_ACCESS_KEY=secret_key
```
For every RabbitMQ -> SNS/SQS pair, RabbitMQ should have exported environment variable with connection URL i.e.:
```bash
export RABBIT_URL_ENV=amqp://guest:guest@localhost:5672/
```
and then in mapping file field `Connection` should be set to `RABBIT_URL_ENV`.

# Build

```bash
make release
```

# Run

```bash
docker-compose up
```

# Supervisor

Supervisor is a module which starts the consumer->forwarder pairs.
Exposed endpoints:
- `APP_URL/health` - returns status if all consumers are running
- `APP_URL/restart` - restarts all consumer->forwarder pairs
