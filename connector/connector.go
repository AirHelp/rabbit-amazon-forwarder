package connector

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/symopsio/rabbit-amazon-forwarder/config"

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
	NewCertPoolWithAppendedCa(caCert []byte) *x509.CertPool
}

type X509CertPoolMaker struct {
}

func (x *X509CertPoolMaker) NewCertPoolWithAppendedCa(caCert []byte) *x509.CertPool {
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)
	return certPool
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
	noVerify := os.Getenv(config.NoVerify)
	if noVerify == "1" {
		log.Info("Skipping cert configuration because NoVerify flag was set.")
		c.TlsConfig.MinVersion = tls.VersionTLS12
		c.TlsConfig.InsecureSkipVerify = true
	} else {
		caCertFilePath := os.Getenv(config.CaCertFile)

		if ca, err := c.FileReader.ReadFile(caCertFilePath); err == nil {
			c.TlsConfig.RootCAs = c.CertPoolMaker.NewCertPoolWithAppendedCa(ca)
		} else {
			log.WithFields(log.Fields{
				"error":           err.Error(),
				config.CaCertFile: caCertFilePath}).Info("Error loading CA Cert file")
			return nil, err
		}

		certFilePath := os.Getenv(config.CertFile)
		keyFilePath := os.Getenv(config.KeyFile)
		if cert, err := c.KeyLoader.LoadKeyPair(certFilePath, keyFilePath); err == nil {
			c.TlsConfig.Certificates = append(c.TlsConfig.Certificates, cert)
		} else {
			log.WithFields(log.Fields{
				"error":         err.Error(),
				config.CertFile: certFilePath,
				config.KeyFile:  keyFilePath}).Info("Error loading client certificates")
		}
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
	if strings.HasPrefix(connectionURL, "amqps") {
		return CreateTlsRabbitConnector()
	} else {
		return CreateBasicRabbitConnector()
	}
}
