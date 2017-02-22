package supervisor

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AirHelp/rabbit-amazon-forwarder/consumer"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
)

type consumerChannel struct {
	name  string
	check chan bool
	stop  chan bool
}

// Client supervisor client
type Client struct {
	mappings  map[consumer.Client]forwarder.Client
	consumers map[string]*consumerChannel
}

// New client for supervisor
func New(consumerForwarderMap map[consumer.Client]forwarder.Client) Client {
	return Client{mappings: consumerForwarderMap}
}

// Start starts supervisor
func (c *Client) Start() error {
	c.consumers = make(map[string]*consumerChannel)
	for consumer, forwarder := range c.mappings {
		channel := makeConsumerChannel(forwarder.Name())
		c.consumers[forwarder.Name()] = channel
		if err := consumer.Start(forwarder, channel.check, channel.stop); err != nil {
			return err
		}
		log.Printf("Started consumer:%s with forwader:%s", consumer.Name(), forwarder.Name())
	}
	return nil
}

// Check checks running consumers
func (c *Client) Check(w http.ResponseWriter, r *http.Request) {
	stopped := 0
	for _, consumer := range c.consumers {
		if len(consumer.check) > 0 {
			stopped = stopped + 1
			continue
		}
		consumer.check <- true
		time.Sleep(500 * time.Millisecond)
		if len(consumer.check) > 0 {
			stopped = stopped + 1
		}
	}
	if stopped > 0 {
		w.WriteHeader(500)
		message := fmt.Sprintf("Number of failed consumers: %d", stopped)
		w.Write([]byte(message))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("success"))
}

// Restart restarts every consumer
func (c *Client) Restart(w http.ResponseWriter, r *http.Request) {
	c.stop()
	if err := c.Start(); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("success"))
}

func (c *Client) stop() {
	for _, consumer := range c.consumers {
		consumer.stop <- true
	}

}

func makeConsumerChannel(name string) *consumerChannel {
	check := make(chan bool)
	stop := make(chan bool)
	return &consumerChannel{name: name, check: check, stop: stop}
}
