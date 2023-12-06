package backupmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/sevigo/shugosha/pkg/model"
)

const providersKey = "meta:providers"

// Ensure BackupManager satisfies the ProviderMetaInfoGetter interface
var _ model.ProviderMetaInfoGetter = (*BackupManager)(nil)

func (m *BackupManager) GetMetaInfo(providerName string) (*model.ProviderMetaInfo, error) {
	key := fmt.Sprintf("meta:%s", providerName)
	providerMeta := &model.ProviderMetaInfo{}

	value, err := m.db.Get(key)
	if err != nil {
		return nil, handleError(err, "Failed to get provider meta info")
	}

	if err := json.Unmarshal(value, providerMeta); err != nil {
		return nil, handleError(err, "Failed to unmarshal provider meta info")
	}

	return providerMeta, nil
}

func (m *BackupManager) updateTotalSize(providerName, rootDir string, size int64) {
	key := fmt.Sprintf("meta:%s", providerName)
	slog.Debug("[BackupManager] update total size", "providerName", providerName, "key", key, "size", size, "root", rootDir)

	providerMeta, err := m.getProviderMeta(key, providerName)
	if err != nil {
		slog.Error("Failed to get or unmarshal provider meta info", "error", err)
		return
	}

	// Update the size for the specified directory
	providerMeta.Directories[rootDir] += uint64(size)
	slog.Debug("[BackupManager] new total size is", "providerName", providerName, "key", key, "size", providerMeta.Directories[rootDir], "root", rootDir)

	// Marshal and save the updated provider meta
	if err := m.saveProviderMeta(key, providerMeta); err != nil {
		slog.Error("Failed to marshal or save provider meta info", "error", err)
	}
}

func (m *BackupManager) getProviderMeta(key, providerName string) (*model.ProviderMetaInfo, error) {
	value, err := m.db.Get(key)
	if err != nil && !errors.Is(err, model.ErrDBKeyNotFound) {
		return nil, err
	}

	providerMeta := &model.ProviderMetaInfo{Name: providerName, Directories: map[string]uint64{}}
	if value != nil {
		if err := json.Unmarshal(value, providerMeta); err != nil {
			return nil, err
		}
	}
	return providerMeta, nil
}

func (m *BackupManager) saveProviderMeta(key string, providerMeta *model.ProviderMetaInfo) error {
	updatedValue, err := json.Marshal(providerMeta)
	if err != nil {
		return err
	}
	return m.db.Set(key, updatedValue)
}

func (b *BackupManager) SetProviders(providers []string) error {
	data, err := json.Marshal(providers)
	if err != nil {
		return handleError(err, "Failed to marshal providers")
	}
	return b.db.Set(providersKey, data)
}

func (b *BackupManager) GetProviders() ([]string, error) {
	data, err := b.db.Get(providersKey)
	if err != nil {
		return nil, handleError(err, "Failed to get providers")
	}

	var providers []string
	if err := json.Unmarshal(data, &providers); err != nil {
		return nil, handleError(err, "Failed to unmarshal providers")
	}

	return providers, nil
}

// Utility function to handle common error checks
func handleError(err error, message string) error {
	if err == nil || errors.Is(err, model.ErrDBKeyNotFound) {
		return nil
	}
	slog.Error(message, "error", err)
	return err
}
