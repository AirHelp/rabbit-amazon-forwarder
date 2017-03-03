package sqs

import (
	"log"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	Type = "SQS"
)

type Forwarder struct {
	name      string
	sqsClient *sqs.SQS
	queue     string
}

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.AmazonEntry) forwarder.Client {
	client := awsClient()
	forwarder := Forwarder{entry.Name, client, entry.Target}
	log.Print("Created forwarder: ", forwarder.Name())
	return forwarder
}

// Name forwarder name
func (f Forwarder) Name() string {
	return f.name
}

// Push pushes message to forwarding infrastructure
func (f Forwarder) Push(message string) error {
	params := &sqs.SendMessageInput{
		MessageBody: aws.String(message), // Required
		QueueUrl:    aws.String(f.queue), // Required
	}
	log.Print(params)
	log.Print(*f.sqsClient.Config.Region)
	resp, err := f.sqsClient.SendMessage(params)

	if err != nil {
		log.Printf("[%s] Could not forward message. Error: %s", f.Name(), err.Error())
		return err
	}
	log.Printf("[%s] Forward succeeded. Response: %s", f.Name(), resp)
	return nil
}

func awsClient() *sqs.SQS {
	sess := session.New()
	return sqs.New(sess)
}
