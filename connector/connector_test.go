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

		Context("With an amqps value somewhere else in the connection url", func() {
			It("should be a BasicRabbitConnector", func() {
				actualConnect := connector.CreateConnector("amqp://guest:guest@rabbbit-amqps:5672")
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

		Context("With an error creating the connection", func() {
			It("Should return an error", func() {
				dialer.Error = errors.New("Expected")
				dialer.ReturnedConnection = nil

				connection, err := rabbitConnector.CreateConnection("any amqp url")

				Expect(connection).Should(BeNil())
				Expect(err).Should(Equal(dialer.Error))
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
			keyLoader       *MockkeyLoader
		)

		BeforeEach(func() {
			os.Setenv(config.CaCertFile, "CaName")
			os.Setenv(config.CertFile, "CertFile")
			os.Setenv(config.KeyFile, "KeyFile")

			dialer = &MockTlsRabbitDialer{}
			fileReader = &MockFileReader{
				Error:     nil,
				DummyFile: []byte("Dummy file"),
			}
			tlsConfig = new(tls.Config)
			certPoolMaker = &MockCertPoolMaker{
				CertPoolToReturn: x509.NewCertPool(),
			}
			keyLoader = &MockkeyLoader{
				ReturnedCertificate: tls.Certificate{},
			}
			rabbitConnector = createTlsConnector(dialer, fileReader, tlsConfig, certPoolMaker, keyLoader)
		})

		Context("With no problems creating the connection", func() {
			It("Should create a connection", func() {
				expectedConnection := createDummyAmqpConnection()
				dialer.ReturnedConnection = expectedConnection

				connection, err := rabbitConnector.CreateConnection("any amqps url")

				// assert that file reader loaded the
				Expect(fileReader.FileNameRead).Should(Equal("CaName"))

				// asert that ca is added to root ca
				Expect(certPoolMaker.AppendedCaCert).Should(Equal([]byte("Dummy file")))
				Expect(tlsConfig.RootCAs).Should(Equal(certPoolMaker.CertPoolToReturn))

				// assert that client certifcate is added
				Expect(keyLoader.CertFileProvided).Should(Equal("CertFile"))
				Expect(keyLoader.KeyFileProvided).Should(Equal("KeyFile"))
				Expect(tlsConfig.Certificates).Should(ContainElement(keyLoader.ReturnedCertificate))

				//assert that connection is created with correct params
				Expect(dialer.ConnectionUrlProvided).Should(Equal("any amqps url"))
				Expect(dialer.TlsConfigProvided).Should(Equal(tlsConfig))

				//assert that the connection is returned
				Expect(connection).Should(Equal(expectedConnection))
				Expect(err).Should(BeNil())
			})
		})

		Context("With an error loading the ca certificate", func() {
			It("Should return an error", func() {
				fileReader.Error = errors.New("Expected")
				fileReader.DummyFile = nil

				connection, err := rabbitConnector.CreateConnection("any amqp url")

				Expect(connection).Should(BeNil())
				Expect(err).Should(Equal(fileReader.Error))
			})
		})

		Context("With an error loading client certificates", func() {
			It("Should proceed with creating the connection", func() {
				// We can leave the error handling to the TLS protocol
				// and log an error indicating that no keys were loaded
				var nilCertificate tls.Certificate
				expectedConnection := createDummyAmqpConnection()
				dialer.ReturnedConnection = expectedConnection
				keyLoader.ReturnedCertificate = nilCertificate
				keyLoader.Error = errors.New("Expected")

				connection, err := rabbitConnector.CreateConnection("any amqps url")

				// assert that client certifcate is added
				Expect(len(tlsConfig.Certificates)).Should(Equal(0))

				//assert that connection is created with correct params
				Expect(dialer.ConnectionUrlProvided).Should(Equal("any amqps url"))
				Expect(dialer.TlsConfigProvided).Should(Equal(tlsConfig))

				//assert that the connection is returned
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

func createTlsConnector(
	mockDialer connector.TlsRabbitDialer,
	mockFileReader connector.FileReader,
	tlsConfig *tls.Config,
	certPoolMaker connector.CertPoolMaker,
	keyLoader connector.KeyLoader) *connector.TlsRabbitConnector {
	return &connector.TlsRabbitConnector{
		TlsConfig:     tlsConfig,
		FileReader:    mockFileReader,
		CertPoolMaker: certPoolMaker,
		KeyLoader:     keyLoader,
		TlsDialer:     mockDialer,
	}
}

type MockFileReader struct {
	FileNameRead string
	Error        error
	DummyFile    []byte
}

func (i *MockFileReader) ReadFile(filename string) ([]byte, error) {
	i.FileNameRead = filename
	return i.DummyFile, i.Error
}

type MockkeyLoader struct {
	CertFileProvided    string
	KeyFileProvided     string
	ReturnedCertificate tls.Certificate
	Error               error
}

func (x *MockkeyLoader) LoadKeyPair(certFile string, keyFile string) (tls.Certificate, error) {
	x.CertFileProvided = certFile
	x.KeyFileProvided = keyFile
	return x.ReturnedCertificate, x.Error
}

type MockTlsRabbitDialer struct {
	ConnectionUrlProvided string
	TlsConfigProvided     *tls.Config
	ReturnedConnection    *amqp.Connection
	Error                 error
}

func (s *MockTlsRabbitDialer) DialTLS(connectionURL string, tlsConfig *tls.Config) (*amqp.Connection, error) {
	s.ConnectionUrlProvided = connectionURL
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
