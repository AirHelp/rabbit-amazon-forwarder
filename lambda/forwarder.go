package lambda

import (
	"log"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

const (
	Type = "Lambda"
)

type Forwarder struct {
	name         string
	lambdaClient *lambda.Lambda
	function     string
}

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.AmazonEntry) forwarder.Client {
	client := lambda.New(session.New())
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
	params := &lambda.InvokeInput{
		FunctionName: aws.String(f.function),
		Payload:      []byte(message),
	}
	resp, err := f.lambdaClient.Invoke(params)
	if err != nil {
		log.Printf("[%s] Could not forward message. Error: %s", f.Name(), err.Error())
		return err
	}
	log.Printf("[%s] Forward succeeded. Response: %s", f.Name(), resp)
	return nil
}
