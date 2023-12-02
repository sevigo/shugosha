package backup

// Provider defines the interface for backup providers.
type Provider interface {
	Backup(path string) error
	Name() string
}
