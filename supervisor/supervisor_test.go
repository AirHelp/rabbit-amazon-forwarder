package supervisor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AirHelp/rabbit-amazon-forwarder/forwarder"
	"github.com/AirHelp/rabbit-amazon-forwarder/mapping"
)

func TestStart(t *testing.T) {
	supervisor := New(prepareConsumers())
	if err := supervisor.Start(); err != nil {
		t.Error("could not start supervised consumer->forwader pairs, error: ", err.Error())
	}
	if len(supervisor.consumers) != 3 {
		t.Errorf("wrong number of consumer-forwarder pairs, expected:%d, got:%d: ", 3, len(supervisor.consumers))
	}
}

func TestRestart(t *testing.T) {
	successJSON := response{Healthy: true, Message: success}
	sucessMessage, err := json.Marshal(successJSON)
	if err != nil {
		t.Error("Could not prepare response. Error: ", err.Error())
	}
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
	if rr.Body.String() != string(sucessMessage) {
		t.Errorf("wrong response body, expected:%s, got:%v", "success", rr.Body.String())
	}
	if rr.Header().Get(contentType) != jsonType {
		t.Errorf("wrong response header, expected:%s, got:%s", jsonType, rr.Header().Get(contentType))
	}
}

func TestCheck(t *testing.T) {
	successJSON := response{Healthy: true, Message: success}
	sucessMessage, err := json.Marshal(successJSON)
	if err != nil {
		t.Error("Could not prepare response. Error: ", err.Error())
	}
	notAccpetedJSON := response{Healthy: false, Message: notSupported}
	notAcceptedMessage, err := json.Marshal(notAccpetedJSON)
	if err != nil {
		t.Error("Could not prepare response. Error: ", err.Error())
	}
	supervisor := New(prepareConsumers())
	if err = supervisor.Start(); err != nil {
		t.Error("could not start supervised consumer->forwader pairs, error: ", err.Error())
	}

	cases := []struct {
		httpCode int
		res      string
		accept   string
	}{
		{200, string(sucessMessage), ""},
		{200, string(sucessMessage), jsonType},
		{200, string(sucessMessage), acceptAll},
		{406, string(notAcceptedMessage), "plain/text"},
	}
	for _, c := range cases {
		req, err := http.NewRequest("GET", "/check", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Accept", c.accept)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(supervisor.Check)

		handler.ServeHTTP(rr, req)

		if rr.Code != c.httpCode {
			t.Errorf("wrong status code, expected:%d, got:%d", rr.Code, c.httpCode)
		}
		if rr.Body.String() != c.res {
			t.Errorf("wrong response body, expected:%s, got:%s", c.res, rr.Body.String())
		}
		if rr.Header().Get(contentType) != jsonType {
			t.Errorf("wrong response header, expected:%s, got:%s", jsonType, rr.Header().Get(contentType))
		}
	}
}

func prepareConsumers() []mapping.ConsumerForwarderMapping {
	var consumers []mapping.ConsumerForwarderMapping
	consumers = append(consumers, mapping.ConsumerForwarderMapping{Consumer: MockRabbitConsumer{"rabbit"}, Forwarder: MockSNSForwarder{"sns"}})
	consumers = append(consumers, mapping.ConsumerForwarderMapping{Consumer: MockRabbitConsumer{"rabbit"}, Forwarder: MockSQSForwarder{"sqs"}})
	consumers = append(consumers, mapping.ConsumerForwarderMapping{Consumer: MockRabbitConsumer{"rabbit"}, Forwarder: MockLambdaForwarder{"lambda"}})
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

type MockLambdaForwarder struct {
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

func (f MockLambdaForwarder) Name() string {
	return f.name
}

func (f MockLambdaForwarder) Push(message string) error {
	return nil
}
