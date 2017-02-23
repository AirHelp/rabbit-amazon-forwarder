package sqs

import (
	"testing"

	"github.com/AirHelp/rabbit-amazon-forwarder/common"
)

func TestCreateForwarder(t *testing.T) {
	item := common.Item{Type: "SQS",
		Name:          "sqs-test",
		ConnectionEnv: "",
		ExchangeName:  "",
		QueueName:     "test-queue",
		RoutingKey:    "#"}
	forwarder := CreateForwarder(item)
	if forwarder.Name() != item.Name {
		t.Errorf("wrong forwarder name, expected:%s, found: %s", item.Name, forwarder.Name())
	}
}
