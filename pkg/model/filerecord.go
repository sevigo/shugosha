package model

import "time"

// FileRecord holds information about a backed-up file.
type FileRecord struct {
	Path         string            `json:"path"`
	Timestamp    time.Time         `json:"timestamp"`
	Checksum     string            `json:"checksum"`
	Provider     string            `json:"provider"`
	ProviderData map[string]string `json:"provider_data"` // Map of provider names to checksums
}
