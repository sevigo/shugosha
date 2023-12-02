package config

import (
	"encoding/json"
	"fmt"

	"github.com/sevigo/shugosha/pkg/db"
)

const key = "config:backupConfig"

type ConfigManager struct {
	db db.DB
}

func NewConfigManager(db db.DB, keyPrefix string) *ConfigManager {
	return &ConfigManager{
		db: db,
	}
}

func (cm *ConfigManager) SaveConfig(config *BackupConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := cm.db.Set(key, data); err != nil {
		return fmt.Errorf("failed to set config in DB: %w", err)
	}

	return nil
}

func (cm *ConfigManager) LoadConfig() (*BackupConfig, error) {
	data, err := cm.db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get config from DB: %w", err)
	}

	var config BackupConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
