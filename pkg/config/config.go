package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
)

type BackupConfig struct {
	Providers []ProviderConfig `json:"providers"`
}

type ProviderConfig struct {
	Name          string            `json:"name"`
	Type          string            `json:"type"`     // e.g., "Echo", "AWS"
	Settings      map[string]string `json:"settings"` // Provider-specific settings like access keys
	DirectoryList []string          `json:"directoryList"`
}

func LoadDefaultConfig() *BackupConfig {
	var configPath string
	switch runtime.GOOS {
	case "windows":
		configPath = "config.win.json"
	case "darwin":
		configPath = "config.mac.json"
	case "linux":
		configPath = "config.linux.json"
	default:
		log.Fatalf("Unsupported operating system")
	}

	backupConfig, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	return backupConfig
}

func LoadConfig(configPath string) (*BackupConfig, error) {
	var config BackupConfig

	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
