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
	if err := handleError(err, "Failed to get provider meta info"); err != nil {
		return nil, err
	}

	if len(value) == 0 {
		return providerMeta, nil
	}

	if err := json.Unmarshal(value, providerMeta); err != nil {
		return nil, handleError(err, "Failed to unmarshal provider meta info")
	}

	return providerMeta, nil
}

func (m *BackupManager) updateTotalSize(providerName, rootDir string, size int64) {
	key := fmt.Sprintf("meta:%s", providerName)
	providerMeta := model.ProviderMetaInfo{}

	value, err := m.db.Get(key)
	if err := handleError(err, "Failed to get provider meta info"); err != nil {
		return
	}

	if len(value) > 0 {
		if err := json.Unmarshal(value, &providerMeta); err != nil {
			handleError(err, "Failed to unmarshal provider meta info")
			return
		}
	} else {
		providerMeta = model.ProviderMetaInfo{Name: providerName, Directories: make(map[string]uint64)}
	}

	providerMeta.Directories[rootDir] += uint64(size)
	updatedValue, err := json.Marshal(providerMeta)
	if err != nil {
		handleError(err, "Failed to marshal provider meta info")
		return
	}

	if err := m.db.Set(key, updatedValue); err != nil {
		handleError(err, "Failed to save provider meta info")
	}
}

func (b *BackupManager) SetProviders(providers []string) error {
	data, err := json.Marshal(providers)
	if err != nil {
		return err
	}
	return b.db.Set(providersKey, data)
}

func (b *BackupManager) GetProviders() ([]string, error) {
	data, err := b.db.Get(providersKey)
	if err := handleError(err, "Failed to get providers"); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []string{}, nil
	}

	var providers []string
	if err := json.Unmarshal(data, &providers); err != nil {
		return nil, handleError(err, "Failed to unmarshal providers")
	}

	return providers, nil
}

// Utility function to handle common error checks
func handleError(err error, message string) error {
	if errors.Is(err, model.ErrDBKeyNotFound) {
		return nil
	}
	if err != nil {
		slog.Error(message, "error", err)
	}
	return err
}
