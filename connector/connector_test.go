package connector_test

import (
	"github.com/AirHelp/rabbit-amazon-forwarder/connector"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Connector", func() {

	Describe("Creating a connector", func() {
		Context("With a basic rabbit configuration", func() {
			It("should be a BasicRabbitConnector", func() {
				actualConnect := connector.CreateConnector("amqp")
				Expect(actualConnect).Should(BeAssignableToTypeOf(&connector.BasicRabbitConnector{}))
			})
		})
	})

	Describe("Creating a connector", func() {
		Context("With a tls rabbit configuration", func() {
			It("should be a TlsRabbitConnector", func() {
				actualConnect := connector.CreateConnector("amqps")
				Expect(actualConnect).Should(BeAssignableToTypeOf(&connector.TlsRabbitConnector{}))
			})
		})
	})
})
