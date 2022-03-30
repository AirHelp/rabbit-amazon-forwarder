package sqs

import (
	"errors"
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/symopsio/rabbit-amazon-forwarder/config"
	"github.com/symopsio/rabbit-amazon-forwarder/forwarder"
)

const (
	// Type forwarder type
	Type = "SQS"
)

// Forwarder forwarding client
type Forwarder struct {
	name      string
	sqsClient sqsiface.SQSAPI
	queue     string
}

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.AmazonEntry, sqsClient ...sqsiface.SQSAPI) forwarder.Client {
	var client sqsiface.SQSAPI
	if len(sqsClient) > 0 {
		client = sqsClient[0]
	} else {
		client = sqs.New(session.Must(session.NewSession()))
	}
	forwarder := Forwarder{entry.Name, client, entry.Target}
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
		MessageBody: aws.String(message), // Required
		QueueUrl:    aws.String(f.queue), // Required
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
