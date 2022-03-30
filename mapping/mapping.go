package mapping

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/AirHelp/rabbit-amazon-forwarder/connector"
	log "github.com/sirupsen/logrus"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/consumer"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/AirHelp/rabbit-amazon-forwarder/lambda"
	"github.com/AirHelp/rabbit-amazon-forwarder/rabbitmq"
	"github.com/AirHelp/rabbit-amazon-forwarder/sns"
	"github.com/AirHelp/rabbit-amazon-forwarder/sqs"
)

type pairs []pair

type pair struct {
	Source      config.RabbitEntry `json:"source"`
	Destination config.AmazonEntry `json:"destination"`
}

// Client mapping client
type Client struct {
	helper Helper
}

// Helper interface for creating consumers and forwaders
type Helper interface {
	createConsumer(entry config.RabbitEntry) consumer.Client
	createForwarder(entry config.AmazonEntry) forwarder.Client
}

// ConsumerForwarderMapping mapping for consumers and forwarders
type ConsumerForwarderMapping struct {
	Consumer  consumer.Client
	Forwarder forwarder.Client
}

type helperImpl struct{}

// New creates new mapping client
func New(helpers ...Helper) Client {
	var helper Helper
	helper = helperImpl{}
	if len(helpers) > 0 {
		helper = helpers[0]
	}
	return Client{helper}
}

// Load loads mappings
func (c Client) Load() ([]ConsumerForwarderMapping, error) {
	var consumerForwarderMapping []ConsumerForwarderMapping
	runtimeLambda = os.Getenv("RUNTIME_LAMBDA_ARN")
	log.Info("Runtime Lambda ARN: ", runtimeLambda)
	pairsList := pairs{
		{
			Source: config.RabbitEntry{
				Type: rabbitmq.Type,
				Name: "runtime-requests",
				ConnectionURLEnvKey: "CELERY_BROKER_URL",
				Topic: "api.internal_messages",
				ExchangeType: "fanout",
				Queue: "RUNTIME_REQUESTS"
			},
			Destination: config.AmazonEntry{
				Type: lambda.Type,
				Name: "runtime-lambda",
				Target: runtimeLambda,
			},
		}
		{
			Source: config.RabbitEntry{
				Type: rabbitmq.Type,
				Name: "audit-messages",
				ConnectionURLEnvKey: "CELERY_BROKER_URL",
				Topic: "api.audit_messages",
				ExchangeType: "fanout",
				Queue: "AUDIT_MESSAGES"
			},
			Destination: config.AmazonEntry{
				Type: lambda.Type,
				Name: "runtime-lambda",
				Target: runtimeLambda,
			},
		}
	}
	log.Info("Loading consumer - forwarder pairs")
	for _, pair := range pairsList {
		consumer := c.helper.createConsumer(pair.Source)
		forwarder := c.helper.createForwarder(pair.Destination)
		consumerForwarderMapping = append(consumerForwarderMapping, ConsumerForwarderMapping{consumer, forwarder})
	}
	return consumerForwarderMapping, nil
}

func (h helperImpl) createConsumer(entry config.RabbitEntry) consumer.Client {
	log.WithFields(log.Fields{
		"consumerType": entry.Type,
		"consumerName": entry.Name}).Info("Creating consumer")
	switch entry.Type {
	case rabbitmq.Type:
		if entry.ConnectionURLEnvKey != "" {
			entry.ConnectionURL = os.Getenv(entry.ConnectionURLEnvKey)
		}
		if entry.ExchangeType == "" {
			entry.ExchangeType = "topic"
		}
		rabbitConnector := connector.CreateConnector(entry.ConnectionURL)
		return rabbitmq.CreateConsumer(entry, rabbitConnector)
	}
	return nil
}

func (h helperImpl) createForwarder(entry config.AmazonEntry) forwarder.Client {
	log.WithFields(log.Fields{
		"forwarderType": entry.Type,
		"forwarderName": entry.Name}).Info("Creating forwarder")
	switch entry.Type {
	case sns.Type:
		return sns.CreateForwarder(entry)
	case sqs.Type:
		return sqs.CreateForwarder(entry)
	case lambda.Type:
		return lambda.CreateForwarder(entry)
	}
	return nil
}
