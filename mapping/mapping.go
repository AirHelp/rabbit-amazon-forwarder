package mapping

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/consumer"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
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
func (c Client) Load() (map[consumer.Client]forwarder.Client, error) {
	consumerForwarderMap := make(map[consumer.Client]forwarder.Client)
	data, err := c.loadFile()
	if err != nil {
		return consumerForwarderMap, err
	}
	var pairsList pairs
	if err = json.Unmarshal(data, &pairsList); err != nil {
		return consumerForwarderMap, err
	}
	log.Print("Loading consumer->forwader pairs")
	for _, pair := range pairsList {
		consumer := c.helper.createConsumer(pair.Source)
		forwarder := c.helper.createForwarder(pair.Destination)
		consumerForwarderMap[consumer] = forwarder
	}
	return consumerForwarderMap, nil
}

func (c Client) loadFile() ([]byte, error) {
	filePath := os.Getenv(config.MappingFile)
	log.Print("Loading mapping file: ", filePath)
	return ioutil.ReadFile(filePath)
}

func (h helperImpl) createConsumer(entry config.RabbitEntry) consumer.Client {
	log.Printf("Creating consumer: [%s, %s]", entry.Type, entry.Name)
	switch entry.Type {
	case rabbitmq.Type:
		return rabbitmq.CreateConsumer(entry)
	}
	return nil
}

func (h helperImpl) createForwarder(entry config.AmazonEntry) forwarder.Client {
	log.Printf("Creating forwarder: [%s, %s]", entry.Type, entry.Name)
	switch entry.Type {
	case sns.Type:
		return sns.CreateForwarder(entry)
	case sqs.Type:
		return sqs.CreateForwarder(entry)
	}
	return nil
}
