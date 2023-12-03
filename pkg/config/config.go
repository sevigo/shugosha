package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/sevigo/shugosha/pkg/model"
)

func LoadDefaultConfig() *model.BackupConfig {
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

func LoadConfig(configPath string) (*model.BackupConfig, error) {
	var config model.BackupConfig

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
