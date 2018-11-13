package lambda

import (
	"encoding/json"
	"errors"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	log "github.com/sirupsen/logrus"
)

const (
	// Type forwarder type
	Type = "Lambda"
)

// Config a struct representing the json config
type Config struct {
	Function string `json:"function"`
}

// Forwarder forwarding client
type Forwarder struct {
	lambdaClient lambdaiface.LambdaAPI
	name         string
	config       Config
}

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.Entry, lambdaClient ...lambdaiface.LambdaAPI) forwarder.Client {
	//Unmarshal Config
	if entry.Config == nil {
		//we need a config
		return nil
	}

	var config Config
	if err := json.Unmarshal(*entry.Config, &config); err != nil {
		return nil
	}

	var client lambdaiface.LambdaAPI
	if len(lambdaClient) > 0 {
		client = lambdaClient[0]
	} else {
		client = lambda.New(session.Must(session.NewSession()))
	}
	forwarder := Forwarder{client, entry.Name, config}
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
	params := &lambda.InvokeInput{
		FunctionName: aws.String(f.config.Function),
		Payload:      []byte(message),
	}
	resp, err := f.lambdaClient.Invoke(params)
	if err != nil {
		log.WithFields(log.Fields{
			"forwarderName": f.Name(),
			"error":         err.Error()}).Error("Could not forward message")
		return err
	}
	if resp.FunctionError != nil {
		log.WithFields(log.Fields{
			"forwarderName": f.Name(),
			"functionError": *resp.FunctionError}).Errorf("Could not forward message")
		return errors.New(*resp.FunctionError)
	}
	log.WithFields(log.Fields{
		"forwarderName": f.Name(),
		"statusCode":    resp.StatusCode}).Info("Forward succeeded")
	return nil
}
