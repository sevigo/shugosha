package backupmanager

import (
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/dgraph-io/badger/v4"
	"github.com/sevigo/shugosha/pkg/model"
)

func (m *BackupManager) updateRecord(path, checksum, providerName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	slog.Debug("[BackupManager] update record", "providerName", providerName, "key", path)

	record, err := m.getOrCreateRecord(path)
	if err != nil {
		slog.Error("Failed to get or create record", "error", err, "key", path)
		return
	}

	if err := m.updateProviderData(record, providerName, checksum); err != nil {
		slog.Error("Failed to update provider data", "error", err, "key", path)
		return
	}

	if err := m.saveRecord(path, record); err != nil {
		slog.Error("Failed to save updated record to DB", "error", err, "key", path)
	}
}

func (m *BackupManager) getOrCreateRecord(path string) (*model.FileRecord, error) {
	record := &model.FileRecord{}

	value, err := m.db.Get(path)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return &model.FileRecord{ProviderData: make(map[string]string)}, nil
		}
		return record, err
	}

	if err := json.Unmarshal(value, &record); err != nil {
		return record, err
	}

	return record, nil
}

func (m *BackupManager) updateProviderData(record *model.FileRecord, providerName, checksum string) error {
	record.ProviderData[providerName] = checksum
	return nil
}

func (m *BackupManager) saveRecord(path string, record *model.FileRecord) error {
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return m.db.Set(path, recordBytes)
}
