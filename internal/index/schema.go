package index

import (
	taskflowv1 "github.com/andy-esch/taskflow/contracts/gen/go/contracts/proto/taskflow/v1"
)

// Index represents the complete state of the project planning.
// It is intended to be serialized to `planning-index.json`.
type Index struct {
	// Version of the index schema
	Version string `json:"version"`

	// Timestamp of last update
	LastUpdated string `json:"last_updated"`

	// List of all tasks
	Tasks []*taskflowv1.Task `json:"tasks"`
}
