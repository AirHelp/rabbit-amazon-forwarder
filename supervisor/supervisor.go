package supervisor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/symopsio/rabbit-amazon-forwarder/mapping"
)

const (
	jsonType     = "application/json"
	success      = "success"
	notSupported = "not supported response format"
	acceptHeader = "Accept"
	contentType  = "Content-Type"
	acceptAll    = "*/*"
)

type response struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message"`
}

type consumerChannel struct {
	name  string
	check chan bool
	stop  chan bool
}

// Client supervisor client
type Client struct {
	mappings  []mapping.ConsumerForwarderMapping
	consumers map[string]*consumerChannel
}

// New client for supervisor
func New(consumerForwarderMapping []mapping.ConsumerForwarderMapping) Client {
	return Client{mappings: consumerForwarderMapping}
}

// Start starts supervisor
func (c *Client) Start() error {
	c.consumers = make(map[string]*consumerChannel)
	for _, mappingEntry := range c.mappings {
		channel := makeConsumerChannel(mappingEntry.Forwarder.Name())
		c.consumers[mappingEntry.Forwarder.Name()] = channel
		go mappingEntry.Consumer.Start(mappingEntry.Forwarder, channel.check, channel.stop)
		log.WithFields(log.Fields{
			"consumerName":  mappingEntry.Consumer.Name(),
			"forwarderName": mappingEntry.Forwarder.Name()}).Info("Started consumer with forwarder")
	}
	return nil
}

// Check checks running consumers
func (c *Client) Check(w http.ResponseWriter, r *http.Request) {
	if accept := r.Header.Get(acceptHeader); accept != "" &&
		!strings.Contains(accept, jsonType) &&
		!strings.Contains(accept, acceptAll) {
		log.WithField("acceptHeader", accept).Warn("Wrong Accept header")
		notAcceptableResponse(w)
		return
	}
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
		message := fmt.Sprintf("Number of failed consumers: %d", stopped)
		errorResponse(w, message)
		return
	}
	successResponse(w)
}

// Restart restarts every consumer
func (c *Client) Restart(w http.ResponseWriter, r *http.Request) {
	c.stop()
	if err := c.Start(); err != nil {
		log.Error(err)
		errorResponse(w, "")
		return
	}
	successResponse(w)
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

func errorResponse(w http.ResponseWriter, message string) {
	w.Header().Set(contentType, jsonType)
	w.WriteHeader(500)
	w.Write([]byte(message))
}

func notAcceptableResponse(w http.ResponseWriter) {
	w.Header().Set(contentType, jsonType)
	w.WriteHeader(406)
	bytes, err := json.Marshal(response{Healthy: false, Message: notSupported})
	if err != nil {
		log.Error(err)
		w.WriteHeader(500)
		return
	}
	w.Write(bytes)
}

func successResponse(w http.ResponseWriter) {
	w.Header().Set(contentType, jsonType)
	w.WriteHeader(200)
	bytes, err := json.Marshal(response{Healthy: true, Message: success})
	if err != nil {
		log.Error(err)
		w.WriteHeader(200)
		return
	}
	w.Write(bytes)
}
