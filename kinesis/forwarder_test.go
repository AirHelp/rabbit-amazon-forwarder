package kinesis

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
)

var badRequest = "Bad request"

func getEntry() config.AmazonEntry {
	return config.AmazonEntry{Type: "Kinesis",
		Name:   "kinesis-test",
		Target: "kinesis-test"}
}

func TestCreateForwarder(t *testing.T) {
	entry := getEntry()
	forwarder := CreateForwarder(entry)
	if forwarder.Name() != entry.Name {
		t.Errorf("wrong forwarder name, expected:%s, found: %s", entry.Name, forwarder.Name())
	}
}

func TestEnqueue(t *testing.T) {
	forwarder := CreateForwarder(getEntry())
	forwarder.Push("Test")
	kinesisForwarder := forwarder.(Forwarder)
	if len(*kinesisForwarder.outputQ) != 1 {
		t.Errorf("Message not added to Queue or too many items in queue")
	}
}

func TestPushMaxQueueSize(t *testing.T) {
	forwarder := CreateForwarder(getEntry(), mockAmazonKinesis{resp: kinesis.PutRecordsOutput{FailedRecordCount: aws.Int64(0)}})
	kinesisForwarder := forwarder.(Forwarder)
	for i := 0; i < maxQUEUELENGTH; i++ {
		forwarder.Push("Test")
	}

	//We have pushed in enough to confirm that the buffer is full and should have been triggered
	if len(*kinesisForwarder.outputQ) != 0 {
		t.Errorf("Queue did not get flushed!")
	}
}

func TestErrorAndRequeue(t *testing.T) {

	mockKinesis := mockAmazonKinesis{
		resp: kinesis.PutRecordsOutput{
			FailedRecordCount: aws.Int64(1),
			Records:           make([]*kinesis.PutRecordsResultEntry, 1),
		},
	}

	mockKinesis.resp.Records[0] = new(kinesis.PutRecordsResultEntry)
	mockKinesis.resp.Records[0].ErrorCode = aws.String("Error")
	mockKinesis.resp.Records[0].ErrorMessage = aws.String("Error")

	forwarder := CreateForwarder(getEntry(), mockKinesis)

	kinesisForwarder := forwarder.(Forwarder)
	for i := 0; i < maxQUEUELENGTH; i++ {
		forwarder.Push("Test")
	}

	//We have pushed in enough to confirm that the buffer is full and should have been triggered
	//There should only be the 1 failed message in the queue
	if len(*kinesisForwarder.outputQ) != 1 {
		t.Errorf("Queue did not get flushed!")
	}
}

type mockAmazonKinesis struct {
	kinesisiface.KinesisAPI
	resp    kinesis.PutRecordsOutput
	message string
}

func (m mockAmazonKinesis) PutRecords(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error) {

	return &m.resp, nil
}
