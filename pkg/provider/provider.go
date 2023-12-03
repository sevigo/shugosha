package provider

import (
	"fmt"
	"log/slog"

	"github.com/sevigo/shugosha/pkg/model"
	"github.com/sevigo/shugosha/pkg/provider/echo"
)

// NewProvider creates a new provider based on the given config.
func NewProvider(providerConf *model.ProviderConfig) (model.Provider, error) {
	switch providerConf.Type {
	case "Echo":
		return echo.NewEchoProvider(providerConf)

	case "AWS":
		return nil, fmt.Errorf("not implemented now")

	default:
		slog.Info("Unknown provider", "type", providerConf.Type)
		return nil, fmt.Errorf("unknown provider")
	}
}

func InitializeProviders(backupConfig *model.BackupConfig) map[string]model.Provider {
	providers := make(map[string]model.Provider)

	for _, providerConfig := range backupConfig.Providers {
		provider, err := NewProvider(&providerConfig)
		if err != nil {
			slog.Error("Error initializing provider", "error", err, "provider", providerConfig.Name)
			continue
		}
		providers[provider.Name()] = provider
	}

	return providers
}
