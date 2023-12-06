package backupmanager

import (
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/sevigo/shugosha/pkg/model"
)

func (m *BackupManager) updateRecord(providerName string, event *model.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	slog.Debug("[BackupManager] update record", "providerName", providerName, "key", event.Path)

	record, err := m.getOrCreateRecord(event)
	if err != nil {
		slog.Error("Failed to get or create record", "error", err, "key", event.Path)
		return
	}

	if err := m.updateProviderData(record, providerName, event); err != nil {
		slog.Error("Failed to update provider data", "error", err, "key", event.Path)
		return
	}

	if err := m.saveRecord(event.Path, record); err != nil {
		slog.Error("Failed to save updated record to DB", "error", err, "key", event.Path)
	}
}

func (m *BackupManager) getOrCreateRecord(event *model.Event) (*model.FileRecord, error) {
	key := event.Path
	record := &model.FileRecord{}

	value, err := m.db.Get(key)
	if err != nil {
		if errors.Is(err, model.ErrDBKeyNotFound) {
			return &model.FileRecord{ProviderData: make(map[string]string)}, nil
		}
		return record, err
	}

	if err := json.Unmarshal(value, &record); err != nil {
		return record, err
	}

	return record, nil
}

func (m *BackupManager) updateProviderData(record *model.FileRecord, providerName string, event *model.Event) error {
	record.ProviderData[providerName] = event.Checksum
	record.Root = event.Root
	record.Size = event.Size

	m.updateTotalSize(providerName, event.Root, event.Size)

	return nil
}

func (m *BackupManager) saveRecord(path string, record *model.FileRecord) error {
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return m.db.Set(path, recordBytes)
}
