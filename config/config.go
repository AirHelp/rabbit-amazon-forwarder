package config

const (
	// MappingFile mapping file environment variable
	MappingFile = "MAPPING_FILE"
)

// RabbitEntry RabbitMQ mapping entry
type RabbitEntry struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	ConnectionURL string `json:"connection"`
	ExchangeName  string `json:"topic"`
	QueueName     string `json:"queue"`
	RoutingKey    string `json:"routing"`
}

// AmazonEntry SQS/SNS mapping entry
type AmazonEntry struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Target string `json:"target"`
}
