package model

import "time"

// FileRecord holds information about a backed-up file.
type FileRecord struct {
	Root         string            `json:"root"`
	Path         string            `json:"path"`
	Timestamp    time.Time         `json:"timestamp"`
	Checksum     string            `json:"checksum"`
	Provider     string            `json:"provider"`
	Size         int64             `json:"size"`
	ProviderData map[string]string `json:"provider_data"`
}
