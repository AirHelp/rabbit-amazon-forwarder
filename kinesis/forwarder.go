package kinesis

import (
	"errors"
	"math/rand"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
)

const (
	// Type forwarder type
	Type = "Kinesis"
)

// Forwarder forwarding client
type Forwarder struct {
	name            string
	kinesisClient   kinesisiface.KinesisAPI
	streamName      string
	outputQ         *[]*kinesis.PutRecordsRequestEntry
	lastSuccessTime *int64
}

// CreateForwarder creates instance of forwarder
func CreateForwarder(entry config.AmazonEntry, kinesisClient ...kinesisiface.KinesisAPI) forwarder.Client {
	log.Info("Test")
	var client kinesisiface.KinesisAPI
	if len(kinesisClient) > 0 {
		client = kinesisClient[0]
	} else {
		client = kinesis.New(session.Must(session.NewSession()))
	}

	outputQ := make([]*kinesis.PutRecordsRequestEntry, 0)
	currentUnixTime := time.Now().UnixNano()
	forwarder := Forwarder{entry.Name, client, entry.Target, &outputQ, &currentUnixTime}
	log.WithField("forwarderName", forwarder.Name()).Info("Created forwarder")

	return forwarder
}

// Name forwarder name
func (f Forwarder) Name() string {
	return f.name
}

func (f Forwarder) flushQueuedMessages() error {

	if len(*f.outputQ) > 0 {

		log.Info("Writing out ", len(*f.outputQ), " records to Kinesis")

		inputRecords := &kinesis.PutRecordsInput{
			StreamName: aws.String(f.streamName),
			Records:    *f.outputQ}

		resp, err := f.kinesisClient.PutRecords(inputRecords)

		*f.outputQ = (*f.outputQ)[:0]
		*f.lastSuccessTime = time.Now().UnixNano()

		if err != nil {
			log.WithFields(log.Fields{
				"forwarderName": f.Name(),
				"error":         err.Error()}).Error("Could not forward message")
			return err
		}

		for _, item := range resp.Records {
			log.WithFields(log.Fields{
				"forwarderName":  f.Name(),
				"SequenceNumber": item.SequenceNumber}).Debug("Forward succeeded")
		}
	}

	return nil
}

// Push pushes message to forwarding infrastructure
func (f Forwarder) Push(message string) error {
	if message == "" {
		return errors.New(forwarder.EmptyMessageError)
	}

	inputRecord := &kinesis.PutRecordsRequestEntry{
		Data:         []byte(message),                            // Required
		PartitionKey: aws.String(strconv.Itoa(rand.Intn(10000)))} //use something random for partition

	*f.outputQ = append(*f.outputQ, inputRecord)

	currentUnixTime := time.Now().UnixNano()

	if (currentUnixTime-*f.lastSuccessTime >= 1000*1000*1000) ||
		(len(*f.outputQ) >= 500) { //maximize for 1 second bursts or 500 items
		f.flushQueuedMessages()
	}

	return nil
}

// Stop stops the forwarder in this case it attempts a flush
func (f Forwarder) Stop() error {
	log.Info("Stopping")
	f.flushQueuedMessages()
	return nil
}
