package connector_test

import (
	"crypto/tls"

	"github.com/AirHelp/rabbit-amazon-forwarder/connector"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streadway/amqp"
)

var _ = Describe("Connector", func() {

	Describe("Creating a connector", func() {
		Context("With a basic rabbit configuration", func() {
			It("should be a BasicRabbitConnector", func() {
				actualConnect := connector.CreateConnector("amqp")
				Expect(actualConnect).Should(BeAssignableToTypeOf(&connector.BasicRabbitConnector{}))
			})
		})

		Context("With a tls rabbit configuration", func() {
			It("should be a TlsRabbitConnector", func() {
				actualConnect := connector.CreateConnector("amqps")
				Expect(actualConnect).Should(BeAssignableToTypeOf(&connector.TlsRabbitConnector{}))
			})
		})
	})

	Describe("Calling connect", func() {
		Context("With a basic rabbit connector", func() {
			It("Should create a connection", func() {
				expectedConnection := createDummyAmqpConnection()
				dialer := &MockBasicRabbitDialer{
					ReturnedConnection: expectedConnection,
					Error:              nil,
				}
				rabbitConnector := createBasicConnector(dialer)

				connection, err := rabbitConnector.CreateConnection("any amqp url")

				Expect(connection).Should(Equal(expectedConnection))
				Expect(err).Should(BeNil())
			})
		})

		Context("With a tls rabbit connect", func() {
			It("Should create a connection", func() {
				expectedConnection := createDummyAmqpConnection()
				dialer := &MockTlsRabbitDialer{
					ReturnedConnection: expectedConnection,
					Error:              nil,
				}
				rabbitConnector := createTlsConnector(dialer)
				connection, err := rabbitConnector.CreateConnection("any amqp url")
				Expect(connection).Should(Equal(expectedConnection))
				Expect(err).Should(BeNil())
			})
		})
	})
})

func createBasicConnector(mockDialer connector.RabbitDialer) *connector.BasicRabbitConnector {
	return &connector.BasicRabbitConnector{
		BasicRabbitDialer: mockDialer,
	}
}

func createTlsConnector(mockDialer connector.TlsRabbitDialer) *connector.TlsRabbitConnector {
	return &connector.TlsRabbitConnector{
		TlsConfig:     new(tls.Config),
		FileReader:    &MockFileReader{},
		CertPoolMaker: &connector.X509CertPoolMaker{},
		KeyLoader:     &MockKeyPairLoader{},
		TlsDialer:     mockDialer,
	}
}

type MockFileReader struct {
}

func (i *MockFileReader) ReadFile(filename string) ([]byte, error) {
	return []byte("Dummy file"), nil
}

type MockKeyPairLoader struct {
}

func (x *MockKeyPairLoader) LoadKeyPair(certFile string, keyFile string) (tls.Certificate, error) {
	return tls.Certificate{}, nil
}

type MockTlsRabbitDialer struct {
	Called             bool
	ReturnedConnection *amqp.Connection
	Error              error
}

func (s *MockTlsRabbitDialer) DialTLS(connectionURL string, tlsConfig *tls.Config) (*amqp.Connection, error) {
	s.Called = true
	return s.ReturnedConnection, s.Error
}

type MockBasicRabbitDialer struct {
	Called             bool
	ReturnedConnection *amqp.Connection
	Error              error
}

func (s *MockBasicRabbitDialer) Dial(connectionURL string) (*amqp.Connection, error) {
	s.Called = true
	return s.ReturnedConnection, s.Error
}

func createDummyAmqpConnection() *amqp.Connection {
	return &amqp.Connection{}
}
