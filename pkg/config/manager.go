package config

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/sevigo/shugosha/pkg/model"
)

const key = "config:backupConfig"

type Manager struct {
	db model.DB
}

func NewConfigManager(storage model.DB) (model.ConfigManager, error) {
	manager := &Manager{
		db: storage,
	}

	_, err := manager.LoadConfig()
	if err != nil {
		slog.Debug("No existing configuration found. Saving default configuration")
		backupConfig := LoadDefaultConfig()
		if err := manager.SaveConfig(backupConfig); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
	}

	slog.Debug("Loaded existing configuration.")
	return manager, nil
}

func (m *Manager) SaveConfig(config *model.BackupConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := m.db.Set(key, data); err != nil {
		return fmt.Errorf("failed to set config in DB: %w", err)
	}

	slog.Debug("Configuration saved successfully.")

	return nil
}

func (m *Manager) LoadConfig() (*model.BackupConfig, error) {
	data, err := m.db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get config from DB: %w", err)
	}

	var config model.BackupConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
