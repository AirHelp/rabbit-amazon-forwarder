package config

const (
	// MappingFile mapping file environment variable
	MappingFile = "MAPPING_FILE"
	CaCertFile  = "CA_CERT_FILE"
	NoVerify    = "NO_VERIFY"
	CertFile    = "CERT_FILE"
	KeyFile     = "KEY_FILE"
)

// RabbitEntry RabbitMQ mapping entry
type RabbitEntry struct {
	Type                string   `json:"type"`
	Name                string   `json:"name"`
	ConnectionURL       string   `json:"connection"`
	ConnectionURLEnvKey string   `json:"connection_env_key"`
	ExchangeName        string   `json:"topic"`
	ExchangeType        string   `json:"exchange_type"`
	QueueName           string   `json:"queue"`
	RoutingKey          string   `json:"routing"`
	RoutingKeys         []string `json:"routingKeys"`
}

// AmazonEntry SQS/SNS mapping entry
type AmazonEntry struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Target string `json:"target"`
}
