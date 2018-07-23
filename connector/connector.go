package connector

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"strings"

	"github.com/AirHelp/rabbit-amazon-forwarder/config"
	log "github.com/sirupsen/logrus"

	"github.com/streadway/amqp"
)

type FileReader interface {
	ReadFile(filename string) ([]byte, error)
}

type IOFileReader struct {
}

func (i *IOFileReader) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(os.Getenv(config.CaCertFile))
}

type CertPoolMaker interface {
	NewCertPool() *x509.CertPool
}

type X509CertPoolMaker struct {
}

func (x *X509CertPoolMaker) NewCertPool() *x509.CertPool {
	return x509.NewCertPool()
}

type KeyLoader interface {
	LoadKeyPair(certFile, keyFile string) (tls.Certificate, error)
}

type X509KeyPairLoader struct {
}

func (x *X509KeyPairLoader) LoadKeyPair(certFile, keyFile string) (tls.Certificate, error) {
	return tls.LoadX509KeyPair(os.Getenv(config.CertFile), os.Getenv(config.KeyFile))
}

type RabbitDialer interface {
	Dial(connectionURL string) (*amqp.Connection, error)
}

type BasicRabbitDialer struct {
}

func (s *BasicRabbitDialer) Dial(connectionURL string) (*amqp.Connection, error) {
	log.Info("Dialing in")
	return amqp.Dial(connectionURL)
}

type TlsRabbitDialer interface {
	DialTLS(connectionURL string, tlsConfig *tls.Config) (*amqp.Connection, error)
}

type X509TlsDialer struct {
}

func (s *X509TlsDialer) DialTLS(connectionURL string, tlsConfig *tls.Config) (*amqp.Connection, error) {
	return amqp.DialTLS(connectionURL, tlsConfig)
}

type RabbitConnector interface {
	CreateConnection(connectionURL string) (*amqp.Connection, error)
}

type BasicRabbitConnector struct {
	basicRabbitDialer *BasicRabbitDialer
}

func (c *BasicRabbitConnector) CreateConnection(connectionURL string) (*amqp.Connection, error) {
	log.Info("Dialing in")
	return c.basicRabbitDialer.Dial(connectionURL)
}

type TlsRabbitConnector struct {
	tlsConfig     *tls.Config
	fileReader    FileReader
	certPoolMaker CertPoolMaker
	keyLoader     KeyLoader
	tlsDialer     TlsRabbitDialer
}

func (c *TlsRabbitConnector) CreateConnection(connectionURL string) (*amqp.Connection, error) {
	log.Info("Dialing in via TLS")
	c.tlsConfig.RootCAs = c.certPoolMaker.NewCertPool()
	if ca, err := c.fileReader.ReadFile(os.Getenv(config.CaCertFile)); err == nil {
		c.tlsConfig.RootCAs.AppendCertsFromPEM(ca)
	} else {
		log.WithField("error", err.Error()).Error("File not found")
	}
	if cert, err := c.keyLoader.LoadKeyPair(os.Getenv(config.CertFile), os.Getenv(config.KeyFile)); err == nil {
		c.tlsConfig.Certificates = append(c.tlsConfig.Certificates, cert)
	} else {
		log.WithField("error", err.Error()).Error("File not found")
	}
	return c.tlsDialer.DialTLS(connectionURL, c.tlsConfig)
}

func CreateBasicRabbitConnector() *BasicRabbitConnector {
	return &BasicRabbitConnector{
		basicRabbitDialer: &BasicRabbitDialer{},
	}
}

func CreateTlsRabbitConnector() *TlsRabbitConnector {
	return &TlsRabbitConnector{
		tlsConfig:     new(tls.Config),
		fileReader:    &IOFileReader{},
		certPoolMaker: &X509CertPoolMaker{},
		keyLoader:     &X509KeyPairLoader{},
		tlsDialer:     &X509TlsDialer{},
	}
}

func CreateConnector(connectionURL string) RabbitConnector {
	if strings.Contains(connectionURL, "amqps") {
		return CreateTlsRabbitConnector()
	} else {
		return CreateBasicRabbitConnector()
	}
}
