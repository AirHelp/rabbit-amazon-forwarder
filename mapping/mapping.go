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
	createConsumer(item common.Item) (consumer.Client, error)
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
		consumer, err := c.helper.createConsumer(pair.Source)
		if err != nil {
			return consumerForwarderMap, err
		}
		forwarder := c.helper.createForwarder(pair.Destination)
		consumerForwarderMap[consumer] = forwarder
	}
	return consumerForwarderMap, nil
}

func (c Client) loadFile() ([]byte, error) {
	filePath := os.Getenv(common.MappingFile)
	log.Print("Loading mapping file: ", filePath)
	return ioutil.ReadFile(filePath)
}

func (h helperImpl) createConsumer(item common.Item) (consumer.Client, error) {
	log.Printf("Creating consumer: [%s, %s]", item.Type, item.Name)
	switch item.Type {
	case rabbitmq.Type:
		return rabbitmq.CreateConsumer(item)
	}
	return nil, nil
}

func (h helperImpl) createForwarder(item common.Item) forwarder.Client {
	log.Printf("Creating forwarder: [%s, %s]", item.Type, item.Name)
	switch item.Type {
	case sns.Type:
		return sns.CreateForwarder(item)
	case sqs.Type:
		return sqs.CreateForwarder(item)
	}
	return nil
}
