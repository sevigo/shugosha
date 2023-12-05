package model

import "github.com/sevigo/shugosha/pkg/fsmonitor"

// Provider defines the interface for backup providers.
type Provider interface {
	Backup(event fsmonitor.Event) error
	DirectoryList() []string
	Name() string
}

type ProviderMetaInfo struct {
	Name        string            `json:"name"`
	Directories map[string]uint64 `json:"directories"`
}
