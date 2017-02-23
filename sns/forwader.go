package sns

import (
	"log"

	"github.com/AirHelp/rabbit-amazon-forwarder/common"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

const (
	Type = "SNS"
)

type Forwarder struct {
	name      string
	snsClient *sns.SNS
	topic     string
}

// CreateForwarder creates instance of forwarder
func CreateForwarder(item common.Item) forwarder.Client {
	client := awsClient()
	forwarder := Forwarder{item.Name, client, item.ExchangeName}
	log.Print("Created forwarder: ", forwarder.Name())
	return forwarder
}

// Name forwarder name
func (f Forwarder) Name() string {
	return f.name
}

// Push pushes message to forwarding infrastructure
func (f Forwarder) Push(message string) error {
	params := &sns.PublishInput{
		Message:   aws.String(message),
		TargetArn: aws.String(f.topic),
	}

	resp, err := f.snsClient.Publish(params)
	if err != nil {
		log.Printf("[%s] Could not forward message. Error: %s", f.Name(), err.Error())
		return err
	}
	log.Printf("[%s] Forward succeeded. Response: %s", f.Name(), resp)
	return nil
}

func awsClient() *sns.SNS {
	sess := session.New()
	return sns.New(sess)
}
