package connector_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConnector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Connector Suite")
}
