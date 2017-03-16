package rabbitmq

import (
	"fmt"
	"log"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/consumer"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/streadway/amqp"
)

const (
	Type = "RabbitMQ"
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
	conn, err := amqp.Dial(c.ConnectionURL)
	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
	}

	ch, err := conn.Channel()
	if err != nil {
		failOnError(err, "Failed to open a channel")
	}

	err = ch.ExchangeDeclare(
		c.ExchangeName, // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return failOnError(err, "Failed to declare an exchange")
	}

	queue, err := ch.QueueDeclare(
		c.QueueName, // name
		true,        // durable
		false,       // delete when usused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		failOnError(err, "Failed to declare a queue")
	}
	err = ch.QueueBind(
		queue.Name,     // queue name
		c.RoutingKey,   // routing key
		c.ExchangeName, // exchange
		false,
		nil)
	if err != nil {
		failOnError(err, "Failed to bind a queue")
	}

	msgs, err := ch.Consume(
		queue.Name, // queue
		c.Name(),   // consumer
		false,      // auto ack
		false,      // exclusive
		false,      // no local
		false,      // no wait
		nil,        // args
	)
	if err != nil {
		return failOnError(err, "Failed to register a consumer")
	}
	params := workerParams{forwarder, msgs, check, stop, conn, ch}
	go c.startForwarding(&params)

	return nil
}

func (c Consumer) startForwarding(params *workerParams) {
	forwarderName := params.forwarder.Name()
	log.Printf("[%s] Started forwarding messages to %s", c.Name(), forwarderName)
	for {
		select {
		case d, ok := <-params.msgs:
			if !ok { // channel already closed
				params.ch.Close()
				params.conn.Close()
				go c.Start(params.forwarder, params.check, params.stop)
				return
			}
			log.Printf("[%s] Message to forward: %v", c.Name(), d.MessageId)
			err := params.forwarder.Push(string(d.Body))
			if err != nil {
				log.Printf("[%s] Could not forward message. Error: %s", forwarderName, err.Error())
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
			break
		}
	}
}

func failOnError(err error, msg string) error {
	return fmt.Errorf("%s: %s", msg, err)
}
