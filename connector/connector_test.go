package connector_test

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	"github.com/AirHelp/rabbit-amazon-forwarder/connector"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streadway/amqp"
)

var _ = Describe("Connector", func() {

	Describe("Creating connectors", func() {
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

	Describe("Connecting a basic rabbit connector", func() {

		var (
			rabbitConnector connector.RabbitConnector
			dialer          *MockBasicRabbitDialer
		)

		BeforeEach(func() {
			dialer = &MockBasicRabbitDialer{}
			rabbitConnector = createBasicConnector(dialer)
		})

		Context("With no problems creating the connection", func() {
			It("Should create a connection", func() {
				expectedConnection := createDummyAmqpConnection()
				dialer.ReturnedConnection = expectedConnection

				connection, err := rabbitConnector.CreateConnection("any amqp url")

				Expect(connection).Should(Equal(expectedConnection))
				Expect(err).Should(BeNil())
			})
		})
	})

	Describe("Connecting a tls rabbit connector", func() {

		var (
			rabbitConnector connector.RabbitConnector
			fileReader      *MockFileReader
			dialer          *MockTlsRabbitDialer
			tlsConfig       *tls.Config
			certPoolMaker   *MockCertPoolMaker
		)

		BeforeEach(func() {
			os.Setenv(config.CaCertFile, "CaName")

			dialer = &MockTlsRabbitDialer{}
			fileReader = &MockFileReader{
				Err:       nil,
				DummyFile: []byte("Dummy file"),
			}
			tlsConfig = new(tls.Config)
			certPoolMaker = &MockCertPoolMaker{
				CertPoolToReturn: x509.NewCertPool(),
			}
			rabbitConnector = createTlsConnector(dialer, fileReader, tlsConfig, certPoolMaker)
		})

		Context("With no problems creating the connection", func() {
			It("Should create a connection", func() {
				expectedConnection := createDummyAmqpConnection()
				dialer.ReturnedConnection = expectedConnection

				connection, err := rabbitConnector.CreateConnection("any amqps url")

				// assert that file reader loaded the
				Expect(fileReader.FileNameRead).Should(Equal("CaName"))

				// asert that certs are wired up
				Expect(certPoolMaker.AppendedCaCert).Should(Equal([]byte("Dummy file")))
				Expect(tlsConfig.RootCAs).Should(Equal(certPoolMaker.CertPoolToReturn))

				//assert that connection is created with correct params
				Expect(dialer.ConnectionUrlProvied).Should(Equal("any amqps url"))
				Expect(dialer.TlsConfigProvided).Should(Equal(tlsConfig))

				//assert that the connection is returned
				Expect(connection).Should(Equal(expectedConnection))
				Expect(err).Should(BeNil())
			})
		})

		Context("With an error loading the ca certificate", func() {
			It("Should return an error", func() {
				fileReader.Err = errors.New("Expected")
				fileReader.DummyFile = nil

				connection, err := rabbitConnector.CreateConnection("any amqp url")
				Expect(connection).Should(BeNil())
				Expect(err).Should(Equal(fileReader.Err))
			})
		})

		Context("With no defined CaCertFile config", func() {
			It("Should create a connection without adding the CA", func() {
				os.Unsetenv(config.CaCertFile)
				expectedConnection := createDummyAmqpConnection()
				dialer.ReturnedConnection = expectedConnection

				connection, err := rabbitConnector.CreateConnection("any amqp url")

				Expect(connection).Should(Equal(expectedConnection))
				Expect(err).Should(BeNil())

				Expect(certPoolMaker.Called).Should(BeFalse())
			})
		})

		Context("With a blank CaCertFile config", func() {
			It("Should create a connection without adding the CA", func() {
				os.Setenv(config.CaCertFile, " ")

				expectedConnection := createDummyAmqpConnection()
				dialer.ReturnedConnection = expectedConnection

				connection, err := rabbitConnector.CreateConnection("any amqp url")

				Expect(connection).Should(Equal(expectedConnection))
				Expect(err).Should(BeNil())

				Expect(certPoolMaker.Called).Should(BeFalse())
			})
		})
	})
})

func createBasicConnector(mockDialer connector.RabbitDialer) *connector.BasicRabbitConnector {
	return &connector.BasicRabbitConnector{
		BasicRabbitDialer: mockDialer,
	}
}

func createTlsConnector(
	mockDialer connector.TlsRabbitDialer,
	mockFileReader connector.FileReader,
	tlsConfig *tls.Config,
	certPoolMaker connector.CertPoolMaker) *connector.TlsRabbitConnector {
	return &connector.TlsRabbitConnector{
		TlsConfig:     tlsConfig,
		FileReader:    mockFileReader,
		CertPoolMaker: certPoolMaker,
		KeyLoader:     &MockKeyPairLoader{},
		TlsDialer:     mockDialer,
	}
}

type MockFileReader struct {
	FileNameRead string
	Err          error
	DummyFile    []byte
}

func (i *MockFileReader) ReadFile(filename string) ([]byte, error) {
	i.FileNameRead = filename
	return i.DummyFile, i.Err
}

type MockKeyPairLoader struct {
}

func (x *MockKeyPairLoader) LoadKeyPair(certFile string, keyFile string) (tls.Certificate, error) {
	return tls.Certificate{}, nil
}

type MockTlsRabbitDialer struct {
	ConnectionUrlProvied string
	TlsConfigProvided    *tls.Config
	ReturnedConnection   *amqp.Connection
	Error                error
}

func (s *MockTlsRabbitDialer) DialTLS(connectionURL string, tlsConfig *tls.Config) (*amqp.Connection, error) {
	s.ConnectionUrlProvied = connectionURL
	s.TlsConfigProvided = tlsConfig
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

type MockCertPoolMaker struct {
	Called           bool
	AppendedCaCert   []byte
	CertPoolToReturn *x509.CertPool
}

func (x *MockCertPoolMaker) NewCertPoolWithAppendedCa(caCert []byte) *x509.CertPool {
	x.AppendedCaCert = caCert
	x.Called = true
	return x.CertPoolToReturn
}

func createDummyAmqpConnection() *amqp.Connection {
	return &amqp.Connection{}
}
