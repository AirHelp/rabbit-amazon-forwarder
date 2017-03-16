package lambda

import (
	"testing"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
)

func TestCreateForwarder(t *testing.T) {
	entry := config.AmazonEntry{Type: "Lambda",
		Name:   "lambda-test",
		Target: "function1-test",
	}
	forwarder := CreateForwarder(entry)
	if forwarder.Name() != entry.Name {
		t.Errorf("wrong forwarder name, expected:%s, found: %s", entry.Name, forwarder.Name())
	}
}
