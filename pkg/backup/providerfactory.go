package backup

// Config holds configuration for initializing providers.
type Config struct {
	ProviderType string // e.g., "Dummy", "S3", "FTP", etc.
}

// NewProvider creates a new provider based on the given config.
func NewProvider(config Config) Provider {
	switch config.ProviderType {
	case "Dummy":
		return NewDummyProvider()

	// Add other cases for different provider types
	default:
		panic("Unknown provider type")
	}
}
