package lambda

import (
	"errors"
	"log"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
)

const (
	// Type forwarder type
	Type = "Lambda"
)

// Forwarder forwarding client
type Forwarder struct {
	name         string
	lambdaClient lambdaiface.LambdaAPI
	function     string
}

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.AmazonEntry, lambdaClient ...lambdaiface.LambdaAPI) forwarder.Client {
	var client lambdaiface.LambdaAPI
	if len(lambdaClient) > 0 {
		client = lambdaClient[0]
	} else {
		client = lambda.New(session.Must(session.NewSession()))
	}
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
	if message == "" {
		return errors.New(forwarder.EmptyMessageError)
	}
	params := &lambda.InvokeInput{
		FunctionName: aws.String(f.function),
		Payload:      []byte(message),
	}
	resp, err := f.lambdaClient.Invoke(params)
	if err != nil {
		log.Printf("[%s] Could not forward message. Error: %s", f.Name(), err.Error())
		return err
	}
	if resp.FunctionError != nil {
		log.Printf("[%s] Could not forward message. Function error: %s", f.Name(), *resp.FunctionError)
		return errors.New(*resp.FunctionError)
	}
	log.Printf("[%s] Forward succeeded. Code:%d, body:%s", f.Name(), resp.StatusCode, string(resp.Payload))
	return nil
}
