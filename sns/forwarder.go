package sns

import (
	"errors"
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/symopsio/rabbit-amazon-forwarder/config"
	"github.com/symopsio/rabbit-amazon-forwarder/forwarder"
)

const (
	// Type forwarder type
	Type = "SNS"
)

// Forwarder forwarding client
type Forwarder struct {
	name      string
	snsClient snsiface.SNSAPI
	topic     string
}

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.AmazonEntry, snsClient ...snsiface.SNSAPI) forwarder.Client {
	var client snsiface.SNSAPI
	if len(snsClient) > 0 {
		client = snsClient[0]
	} else {
		client = sns.New(session.Must(session.NewSession()))
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
	params := &sns.PublishInput{
		Message:   aws.String(message),
		TargetArn: aws.String(f.topic),
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
