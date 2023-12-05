package backupmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/sevigo/shugosha/pkg/model"
)

const providersKey = "meta:providers"

func (m *BackupManager) GetMetaInfo(providerName string) (*model.ProviderMetaInfo, error) {
	key := fmt.Sprintf("meta:%s", providerName)
	providerMeta := &model.ProviderMetaInfo{}

	value, err := m.db.Get(key)
	if err != nil {
		if errors.Is(err, model.ErrDBKeyNotFound) {
			// If not found, return an empty struct with no error.
			return providerMeta, nil
		}
		// Return an error if it's something other than KeyNotFound.
		return nil, err
	}

	// Unmarshal the JSON data into the ProviderMetaInfo struct.
	if err := json.Unmarshal(value, providerMeta); err != nil {
		return nil, err
	}

	return providerMeta, nil
}

func (m *BackupManager) updateTotalSize(providerName, rootDir string, size int64) {
	key := fmt.Sprintf("meta:%s", providerName)
	providerMeta := model.ProviderMetaInfo{}

	value, err := m.db.Get(key)
	if err != nil && !errors.Is(err, model.ErrDBKeyNotFound) {
		slog.Error("Failed to get provider meta info", "error", err)
		return
	}

	if err == nil {
		// Unmarshal if existing data was found
		if err := json.Unmarshal(value, &providerMeta); err != nil {
			slog.Error("Failed to unmarshal provider meta info", "error", err)
			return
		}
	} else {
		// Initialize new ProviderMetaInfo if not found
		providerMeta = model.ProviderMetaInfo{
			Name:        providerName,
			Directories: make(map[string]uint64),
		}
	}

	// Update the size for the specific root directory
	providerMeta.Directories[rootDir] += uint64(size)

	// Marshal and save the updated data back to the database
	updatedValue, err := json.Marshal(providerMeta)
	if err != nil {
		slog.Error("Failed to marshal provider meta info", "error", err)
		return
	}

	if err := m.db.Set(key, updatedValue); err != nil {
		slog.Error("Failed to save provider meta info", "error", err)
	}
}

// SetProviders stores the list of provider names in the database.
func (b *BackupManager) SetProviders(providers []string) error {
	data, err := json.Marshal(providers)
	if err != nil {
		return err
	}
	return b.db.Set(providersKey, data)
}

// GetProviders retrieves the list of provider names from the database.
func (b *BackupManager) GetProviders() ([]string, error) {
	data, err := b.db.Get(providersKey)
	if err != nil {
		if errors.Is(err, model.ErrDBKeyNotFound) {
			// If not found, return an empty list with no error.
			return []string{}, nil
		}
		return nil, err
	}

	var providers []string
	if err := json.Unmarshal(data, &providers); err != nil {
		return nil, err
	}

	return providers, nil
}
