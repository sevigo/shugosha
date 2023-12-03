package model

// ConfigManager defines the interface for managing configurations.
type ConfigManager interface {
	SaveConfig(config *BackupConfig) error
	LoadConfig() (*BackupConfig, error)
}

type BackupConfig struct {
	Providers []ProviderConfig `json:"providers"`
}

type ProviderConfig struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`     // e.g., "Echo", "AWS"
	Settings      map[string]string `json:"settings"` // Provider-specific settings like access keys
	DirectoryList []string          `json:"directoryList"`
}
