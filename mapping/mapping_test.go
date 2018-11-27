package mapping

import (
	"errors"
	"os"
	"testing"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/consumer"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/AirHelp/rabbit-amazon-forwarder/lambda"
	"github.com/AirHelp/rabbit-amazon-forwarder/rabbitmq"
	"github.com/AirHelp/rabbit-amazon-forwarder/sns"
	"github.com/AirHelp/rabbit-amazon-forwarder/sqs"
)

const (
	rabbitType = "rabbit"
	snsType    = "sns"
)

func TestLoad(t *testing.T) {
	os.Setenv(config.MappingFile, "../tests/rabbit_to_sns.json")
	client := New(MockMappingHelper{})
	var consumerForwarderMapping []ConsumerForwarderMapping
	var err error
	if consumerForwarderMapping, err = client.Load(); err != nil {
		t.Errorf("could not load mapping and start mocked rabbit->sns pair: %s", err.Error())
	}
	if len(consumerForwarderMapping) != 1 {
		t.Errorf("wrong consumerForwarderMapping size, expected 1, got %d", len(consumerForwarderMapping))
	}
}

func TestLoadFile(t *testing.T) {
	os.Setenv(config.MappingFile, "../tests/rabbit_to_sns.json")
	client := New()
	data, err := client.loadFile()
	if err != nil {
		t.Errorf("could not load file: %s", err.Error())
	}
	if len(data) < 1 {
		t.Errorf("could not load file: empty steam found")
	}
}

func TestCreateConsumer(t *testing.T) {
	client := New()
	consumerName := "test-rabbit"
	entry := config.RabbitEntry{Type: "RabbitMQ",
		Name:          consumerName,
		ConnectionURL: "url",
		ExchangeName:  "topic",
		QueueName:     "test-queue",
		RoutingKey:    "#"}
	consumer := client.helper.createConsumer(entry)
	if consumer.Name() != consumerName {
		t.Errorf("wrong consumer name, expected %s, found %s", consumerName, consumer.Name())
	}
}

func TestCreateForwarderSNS(t *testing.T) {
	client := New(MockMappingHelper{})
	forwarderName := "test-sns"
	entry := config.AmazonEntry{Type: "SNS",
		Name:   forwarderName,
		Target: "arn",
	}
	forwarder := client.helper.createForwarder(entry)
	if forwarder.Name() != forwarderName {
		t.Errorf("wrong forwarder name, expected %s, found %s", forwarderName, forwarder.Name())
	}
}

func TestCreateForwarderSQS(t *testing.T) {
	client := New(MockMappingHelper{})
	forwarderName := "test-sqs"
	entry := config.AmazonEntry{Type: "SQS",
		Name:   forwarderName,
		Target: "arn",
	}
	forwarder := client.helper.createForwarder(entry)
	if forwarder.Name() != forwarderName {
		t.Errorf("wrong forwarder name, expected %s, found %s", forwarderName, forwarder.Name())
	}
}

func TestCreateForwarderLambda(t *testing.T) {
	client := New(MockMappingHelper{})
	forwarderName := "test-lambda"
	entry := config.AmazonEntry{Type: "Lambda",
		Name:   forwarderName,
		Target: "function-name",
	}
	forwarder := client.helper.createForwarder(entry)
	if forwarder.Name() != forwarderName {
		t.Errorf("wrong forwarder name, expected %s, found %s", forwarderName, forwarder.Name())
	}
}

// helpers
type MockMappingHelper struct{}

type MockRabbitConsumer struct{}

type MockSNSForwarder struct {
	name string
}

type MockSQSForwarder struct {
	name string
}

type MockLambdaForwarder struct {
	name string
}

type ErrorForwarder struct{}

func (h MockMappingHelper) createConsumer(entry config.RabbitEntry) consumer.Client {
	if entry.Type != rabbitmq.Type {
		return nil
	}
	return MockRabbitConsumer{}
}
func (h MockMappingHelper) createForwarder(entry config.AmazonEntry) forwarder.Client {
	switch entry.Type {
	case sns.Type:
		return MockSNSForwarder{entry.Name}
	case sqs.Type:
		return MockSQSForwarder{entry.Name}
	case lambda.Type:
		return MockLambdaForwarder{entry.Name}
	}
	return ErrorForwarder{}
}

func (c MockRabbitConsumer) Name() string {
	return rabbitType
}

func (c MockRabbitConsumer) Start(client forwarder.Client, check chan bool, stop chan bool) error {
	return nil
}

func (f MockSNSForwarder) Name() string {
	return f.name
}

func (f MockSNSForwarder) Push(message string) error {
	return nil
}

func (f MockSQSForwarder) Name() string {
	return f.name
}

func (f MockLambdaForwarder) Push(message string) error {
	return nil
}

func (f MockLambdaForwarder) Name() string {
	return f.name
}

func (f MockSQSForwarder) Push(message string) error {
	return nil
}

func (f ErrorForwarder) Name() string {
	return "error-forwarder"
}

func (f ErrorForwarder) Push(message string) error {
	return errors.New("Wrong forwader created")
}
