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
	return ioutil.ReadFile(filename)
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

func (x *X509KeyPairLoader) LoadKeyPair(certFile string, keyFile string) (tls.Certificate, error) {
	return tls.LoadX509KeyPair(certFile, keyFile)
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
	BasicRabbitDialer RabbitDialer
}

func (c *BasicRabbitConnector) CreateConnection(connectionURL string) (*amqp.Connection, error) {
	log.Info("Dialing in")
	return c.BasicRabbitDialer.Dial(connectionURL)
}

type TlsRabbitConnector struct {
	TlsConfig     *tls.Config
	FileReader    FileReader
	CertPoolMaker CertPoolMaker
	KeyLoader     KeyLoader
	TlsDialer     TlsRabbitDialer
}

func (c *TlsRabbitConnector) CreateConnection(connectionURL string) (*amqp.Connection, error) {
	log.Info("Dialing in via TLS")
	c.TlsConfig.RootCAs = c.CertPoolMaker.NewCertPool()
	log.Info("1")
	if ca, err := c.FileReader.ReadFile(os.Getenv(config.CaCertFile)); err == nil {
		c.TlsConfig.RootCAs.AppendCertsFromPEM(ca)
		log.Info("2")
	} else {
		log.WithField("error", err.Error()).Error("File not found")
	}
	if cert, err := c.KeyLoader.LoadKeyPair(os.Getenv(config.CertFile), os.Getenv(config.KeyFile)); err == nil {
		c.TlsConfig.Certificates = append(c.TlsConfig.Certificates, cert)
	} else {
		log.WithField("error", err.Error()).Error("File not found")
	}
	return c.TlsDialer.DialTLS(connectionURL, c.TlsConfig)
}

func CreateBasicRabbitConnector() *BasicRabbitConnector {
	return &BasicRabbitConnector{
		BasicRabbitDialer: &BasicRabbitDialer{},
	}
}

func CreateTlsRabbitConnector() *TlsRabbitConnector {
	return &TlsRabbitConnector{
		TlsConfig:     new(tls.Config),
		FileReader:    &IOFileReader{},
		CertPoolMaker: &X509CertPoolMaker{},
		KeyLoader:     &X509KeyPairLoader{},
		TlsDialer:     &X509TlsDialer{},
	}
}

func CreateConnector(connectionURL string) RabbitConnector {
	if strings.Contains(connectionURL, "amqps") {
		return CreateTlsRabbitConnector()
	} else {
		return CreateBasicRabbitConnector()
	}
}
