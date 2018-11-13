package config

import (
	"encoding/json"
)

const (
	// MappingFile mapping file environment variable
	MappingFile = "MAPPING_FILE"
)

// Entry Generic config entry
type Entry struct {
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Config *json.RawMessage       `json:"config"`
}
