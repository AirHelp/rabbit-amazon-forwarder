package rabbitmq

import (
	"fmt"
	"log"

	"github.com/AirHelp/rabbit-amazon-forwarder/common"
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

// CreateConsumer creates conusmer from string map
func CreateConsumer(item common.Item) consumer.Client {
	return Consumer{item.Name, item.ConnectionURL, item.ExchangeName, item.QueueName, item.RoutingKey}
}

// Name consumer name
func (c Consumer) Name() string {
	return c.name
}

// TODO gracefull shotdown
// Consume consumes messages from Rabbit queue
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

	go c.push(forwarder, msgs, check, stop, conn, ch)

	return nil
}

func (c Consumer) push(forwarder forwarder.Client, msgs <-chan amqp.Delivery, check chan bool, stop chan bool, conn *amqp.Connection, ch *amqp.Channel) {
	log.Printf("[%s] Started forwarding messages to %s", c.Name(), forwarder.Name())
	for {
		select {
		case d := <-msgs:
			log.Printf("[%s] Message to forward: %v", c.Name(), d.MessageId)
			err := forwarder.Push(string(d.Body))
			if err != nil {
				log.Printf("[%s] Could not forward message. Error: %s", forwarder.Name(), err.Error())
			} else {
				d.Ack(true)
			}
		case <-check:
			log.Printf("[%s] Checking", forwarder.Name())
		case <-stop:
			log.Printf("[%s] Closing", forwarder.Name())
			ch.Close()
			conn.Close()
			return
		}
	}
}

func failOnError(err error, msg string) error {
	return fmt.Errorf("%s: %s", msg, err)
}
