package model

// Provider defines the interface for backup providers.
type Provider interface {
	Backup(path string) error
	DirectoryList() []string
	Name() string
}
