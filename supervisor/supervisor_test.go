package supervisor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AirHelp/rabbit-amazon-forwarder/consumer"
	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
)

func TestStart(t *testing.T) {
	supervisor := New(prepareConsumers())
	if err := supervisor.Start(); err != nil {
		t.Error("could not start supervised consumer->forwader pairs, error: ", err.Error())
	}
	if len(supervisor.consumers) != 2 {
		t.Errorf("wrong number of consumer-forwarder pairs, expected:%d, got:%d: ", 2, len(supervisor.consumers))
	}
}

func TestRestart(t *testing.T) {
	supervisor := New(prepareConsumers())
	req, err := http.NewRequest("GET", "/restart", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(supervisor.Restart)

	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("wrong status code, expected:%d, got:%d", rr.Code, 200)
	}
	if rr.Body.String() != "success" {
		t.Errorf("wrong response body, expected:%s, got:%v", "success", rr.Body.String())
	}
}

func TestCheck(t *testing.T) {
	supervisor := New(prepareConsumers())
	if err := supervisor.Start(); err != nil {
		t.Error("could not start supervised consumer->forwader pairs, error: ", err.Error())
	}
	req, err := http.NewRequest("GET", "/check", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(supervisor.Check)

	handler.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("wrong status code, expected:%d, got:%d", rr.Code, 200)
	}
	if rr.Body.String() != "success" {
		t.Errorf("wrong response body, expected:%s, got:%v", "success", rr.Body.String())
	}
}

func prepareConsumers() map[consumer.Client]forwarder.Client {
	consumers := make(map[consumer.Client]forwarder.Client)
	consumers[MockRabbitConsumer{"rabbit1"}] = MockSNSForwarder{"sns"}
	consumers[MockRabbitConsumer{"rabbit2"}] = MockSQSForwarder{"sqs"}
	return consumers
}

type MockRabbitConsumer struct {
	name string
}

type MockSNSForwarder struct {
	name string
}

type MockSQSForwarder struct {
	name string
}

func (c MockRabbitConsumer) Name() string {
	return c.name
}

func (c MockRabbitConsumer) Start(client forwarder.Client, check chan bool, stop chan bool) error {
	go func() {
		for {
			select {
			case <-check:
				fmt.Print("Checked")
			}
		}
	}()
	return nil
}

func (f MockSNSForwarder) Name() string {
	return f.name
}

func (f MockSNSForwarder) Push(message string) error {
	return nil
}

func (f MockSQSForwarder) Name() string {
	return f.name
}

func (f MockSQSForwarder) Push(message string) error {
	return nil
}
