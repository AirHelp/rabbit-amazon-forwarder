package sqs

import (
	"encoding/json"
	"errors"

	log "github.com/sirupsen/logrus"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

const (
	// Type forwarder type
	Type = "SQS"
)

// Config i
type Config struct {
	Queue string `json:"queue"`
}

// Forwarder forwarding client
type Forwarder struct {
	name      string
	sqsClient sqsiface.SQSAPI
	config    Config
}

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.Entry, sqsClient ...sqsiface.SQSAPI) forwarder.Client {
	//Unmarshal Config
	if entry.Config == nil {
		//we need a config
		return nil
	}

	var config Config
	if err := json.Unmarshal(*entry.Config, &config); err != nil {
		return nil
	}

	var client sqsiface.SQSAPI
	if len(sqsClient) > 0 {
		client = sqsClient[0]
	} else {
		client = sqs.New(session.Must(session.NewSession()))
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
	params := &sqs.SendMessageInput{
		MessageBody: aws.String(message),        // Required
		QueueUrl:    aws.String(f.config.Queue), // Required
	}

	resp, err := f.sqsClient.SendMessage(params)

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
