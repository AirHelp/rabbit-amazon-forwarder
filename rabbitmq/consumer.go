package rabbitmq

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/consumer"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/streadway/amqp"
)

const (
	// Type consumer type
	Type                      = "RabbitMQ"
	channelClosedMessage      = "Channel closed"
	closedBySupervisorMessage = "Closed by supervisor"
)

// Consumer implementation or RabbitMQ consumer
type Consumer struct {
	name          string
	ConnectionURL string
	ExchangeName  string
	QueueName     string
	RoutingKey    string
}

// parameters for starting consumer
type workerParams struct {
	forwarder forwarder.Client
	msgs      <-chan amqp.Delivery
	check     chan bool
	stop      chan bool
	conn      *amqp.Connection
	ch        *amqp.Channel
}

// CreateConsumer creates conusmer from string map
func CreateConsumer(entry config.RabbitEntry) consumer.Client {
	return Consumer{entry.Name, entry.ConnectionURL, entry.ExchangeName, entry.QueueName, entry.RoutingKey}
}

// Name consumer name
func (c Consumer) Name() string {
	return c.name
}

// Start start consuming messages from Rabbit queue
func (c Consumer) Start(forwarder forwarder.Client, check chan bool, stop chan bool) error {
	log.Print("Starting consumer with params: ", c)
	for {
		delivery, conn, ch, err := c.connect()
		if err != nil {
			log.Print(err)
			time.Sleep(1 * time.Second)
			continue
		}
		params := workerParams{forwarder, delivery, check, stop, conn, ch}
		if err := c.startForwarding(&params); err.Error() == closedBySupervisorMessage {
			break
		}
	}
	return nil
}

func (c Consumer) connect() (<-chan amqp.Delivery, *amqp.Connection, *amqp.Channel, error) {
	deadLetterExchangeName := c.ExchangeName + "-dead-letter"
	deadLetterQueueName := c.QueueName + "-dead-letter"

	conn, err := amqp.Dial(c.ConnectionURL)
	if err != nil {
		return failOnError(err, "Failed to connect to RabbitMQ")
	}
	ch, err := conn.Channel()
	if err != nil {
		return failOnError(err, "Failed to open a channel")
	}

	// regural exchange
	if err = ch.ExchangeDeclare(c.ExchangeName, "topic", true, false, false, false, nil); err != nil {
		return failOnError(err, "Failed to declare an exchange:"+c.ExchangeName)
	}
	// dead-letter-exchange
	if err = ch.ExchangeDeclare(deadLetterExchangeName, "fanout", true, false, false, false, nil); err != nil {
		return failOnError(err, "Failed to declare an exchange:"+deadLetterExchangeName)
	}
	// dead-letter-queue
	if _, err = ch.QueueDeclare(deadLetterQueueName, true, false, false, false, nil); err != nil {
		return failOnError(err, "Failed to declare a queue:"+deadLetterQueueName)
	}
	if err = ch.QueueBind(deadLetterQueueName, "#", deadLetterExchangeName, false, nil); err != nil {
		return failOnError(err, "Failed to bind a queue:"+deadLetterQueueName)
	}
	// regular queue
	if _, err = ch.QueueDeclare(c.QueueName, true, false, false, false,
		amqp.Table{
			"x-dead-letter-exchange": deadLetterExchangeName,
		}); err != nil {
		return failOnError(err, "Failed to declare a queue:"+c.QueueName)
	}
	if err = ch.QueueBind(c.QueueName, c.RoutingKey, c.ExchangeName, false, nil); err != nil {
		return failOnError(err, "Failed to bind a queue:"+c.QueueName)
	}

	msgs, err := ch.Consume(c.QueueName, c.Name(), false, false, false, false, nil)
	if err != nil {
		return failOnError(err, "Failed to register a consumer")
	}
	return msgs, conn, ch, nil
}

func (c Consumer) startForwarding(params *workerParams) error {
	forwarderName := params.forwarder.Name()
	log.Printf("[%s] Started forwarding messages to %s", c.Name(), forwarderName)
	for {
		select {
		case d, ok := <-params.msgs:
			if !ok { // channel already closed
				params.ch.Close()
				params.conn.Close()
				return errors.New(channelClosedMessage)
			}
			log.Printf("[%s] Message to forward: %v", c.Name(), d.MessageId)
			err := params.forwarder.Push(string(d.Body))
			if err != nil {
				log.Printf("[%s] Could not forward message. Error: %v", forwarderName, err)
				if err = d.Reject(false); err != nil {
					log.Printf("[%s] Could not reject message. Error: %v", forwarderName, err)
				}

			} else {
				if err := d.Ack(true); err != nil {
					log.Println("Could not ack message with id:", d.MessageId)
				}
			}
		case <-params.check:
			log.Printf("[%s] Checking", forwarderName)
		case <-params.stop:
			log.Printf("[%s] Closing", forwarderName)
			params.ch.Close()
			params.conn.Close()
			return errors.New(closedBySupervisorMessage)
		}
	}
}

func failOnError(err error, msg string) (<-chan amqp.Delivery, *amqp.Connection, *amqp.Channel, error) {
	return nil, nil, nil, fmt.Errorf("%s: %s", msg, err)
}
