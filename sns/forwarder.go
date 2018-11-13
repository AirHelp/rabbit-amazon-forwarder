package sns

import (
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

const (
	// Type forwarder type
	Type = "SNS"
)

// Config a struct representing the json config
type Config struct {
	Configversion *string `json:"configversion"`
	Topic         string  `json:"topic"`
}

// Forwarder forwarding client
type Forwarder struct {
	name      string
	snsClient snsiface.SNSAPI
	config    Config
}

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.Entry, snsClient ...snsiface.SNSAPI) forwarder.Client {
	//Unmarshal Config
	if entry.Config == nil {
		//we need a config
		return nil
	}

	var config Config
	if err := json.Unmarshal(*entry.Config, &config); err != nil {
		return nil
	}

	if config.Configversion == nil {
		log.Warn("Looks like you're using an old config format version or have forgotten the configversion parameter. We will try and recover")
	}

	var client snsiface.SNSAPI
	if len(snsClient) > 0 {
		client = snsClient[0]
	} else {
		client = sns.New(session.Must(session.NewSession()))
	}
	forwarder := Forwarder{entry.Name, client, config}
	log.WithField("forwarderName", forwarder.Name()).Info("Created forwarder")
	return forwarder
}

// Name forwarder name
func (f Forwarder) Name() string {
	return f.name
}

// Push pushes message to forwarding infrastructure
func (f Forwarder) Push(message string) error {
	if message == "" {
		return errors.New(forwarder.EmptyMessageError)
	}
	params := &sns.PublishInput{
		Message:   aws.String(message),
		TargetArn: aws.String(f.config.Topic),
	}

	resp, err := f.snsClient.Publish(params)
	if err != nil {
		log.WithFields(log.Fields{
			"forwarderName": f.Name(),
			"error":         err.Error()}).Error("Could not forward message")
		return err
	}
	log.WithFields(log.Fields{
		"forwarderName": f.Name(),
		"responseID":    resp.MessageId}).Info("Forward succeeded")
	return nil
}
