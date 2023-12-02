package config

import (
	"encoding/json"
	"fmt"

	"github.com/sevigo/shugosha/pkg/db"
)

const key = "config:backupConfig"

type Manager struct {
	db db.DB
}

func NewConfigManager(storage db.DB) (*Manager, error) {
	manager := &Manager{
		db: storage,
	}

	_, err := manager.LoadConfig()
	if err != nil {
		// If no existing configuration, save the default configuration
		backupConfig := LoadDefaultConfig()
		if err := manager.SaveConfig(backupConfig); err != nil {
			return nil, err
		}
	}

	return manager, nil
}

func (m *Manager) SaveConfig(config *BackupConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := m.db.Set(key, data); err != nil {
		return fmt.Errorf("failed to set config in DB: %w", err)
	}

	return nil
}

func (m *Manager) LoadConfig() (*BackupConfig, error) {
	data, err := m.db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get config from DB: %w", err)
	}

	var config BackupConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
