package mapping

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/AirHelp/rabbit-amazon-forwarder/common"
	"github.com/AirHelp/rabbit-amazon-forwarder/consumer"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/AirHelp/rabbit-amazon-forwarder/rabbitmq"
	"github.com/AirHelp/rabbit-amazon-forwarder/sns"
	"github.com/AirHelp/rabbit-amazon-forwarder/sqs"
)

type pairs []pair

type pair struct {
	Source      common.Item `json:"source"`
	Destination common.Item `json:"destination"`
}

// Client mapping client
type Client struct {
	helper Helper
}

// Helper interface for creating consumers and forwaders
type Helper interface {
	createConsumer(item common.Item) consumer.Client
	createForwarder(item common.Item) forwarder.Client
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

// LoadAndStart loads and starts mappings
func (c Client) LoadAndStart() error {
	data, err := c.loadFile()
	if err != nil {
		return err
	}
	var pairsList pairs
	if err = json.Unmarshal(data, &pairsList); err != nil {
		return err
	}
	log.Print("Starting consumer->forwader pairs")
	for _, pair := range pairsList {
		consumer := c.helper.createConsumer(pair.Source)
		forwarder := c.helper.createForwarder(pair.Destination)
		log.Printf("Starting consumer:%s with forwader:%s", consumer.Name(), forwarder.Name())
		if err := consumer.Consume(forwarder); err != nil {
			return err
		}
	}
	return nil
}

func (c Client) loadFile() ([]byte, error) {
	filePath := os.Getenv(common.MappingFile)
	log.Print("Loading mapping file: ", filePath)
	return ioutil.ReadFile(filePath)
}

func (h helperImpl) createConsumer(item common.Item) consumer.Client {
	log.Print("Creating consumer: ", item.Type)
	switch item.Type {
	case rabbitmq.Type:
		return rabbitmq.CreateConsumer(item)
	}
	return nil
}

func (h helperImpl) createForwarder(item common.Item) forwarder.Client {
	log.Print("Creating forwarder: ", item.Type)
	switch item.Type {
	case sns.Type:
		return sns.CreateForwarder(item)
	case sqs.Type:
		return sqs.CreateForwarder(item)
	}
	return nil
}
