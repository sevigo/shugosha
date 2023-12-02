package config

type BackupConfig struct {
	Providers     []ProviderConfig  `json:"providers"`
	DirectoryList []string          `json:"directoryList"`
	DirectoryMap  map[string]string `json:"directoryMap"`
}

type ProviderConfig struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`     // e.g., "Echo", "AWS"
	Settings map[string]string `json:"settings"` // Provider-specific settings like access keys
}
