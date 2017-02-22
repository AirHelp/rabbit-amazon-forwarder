package common

const (
	MappingFile = "MAPPING_FILE"
)

type Item struct {
	Type          string `json:"type"`
	Name          string `json:"name"`
	ConnectionURL string `json:"connection"`
	ExchangeName  string `json:"topic"`
	QueueName     string `json:"queue"`
	RoutingKey    string `json:"routing"`
}
