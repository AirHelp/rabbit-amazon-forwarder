package common

const (
	// MappingFile path to mapping file
	MappingFile = "MAPPING_FILE"
)

// Item json representation of component
type Item struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	ConnectionEnv string `json:"connection"`
	ExchangeName  string `json:"topic"`
	QueueName     string `json:"queue"`
	RoutingKey    string `json:"routing"`
}
