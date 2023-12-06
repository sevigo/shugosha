package model

// Provider defines the interface for backup providers.
type Provider interface {
	Backup(event Event) error
	DirectoryList() []string
	Name() string
}

type ProviderMetaInfo struct {
	Name        string            `json:"name"`
	Directories map[string]uint64 `json:"directories"`
}

type ProviderMetaInfoGetter interface {
	GetProviders() ([]string, error)
	GetMetaInfo(providerName string) (*ProviderMetaInfo, error)
}
