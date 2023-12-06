package backupmanager

import (
	"encoding/json"
	"log/slog"

	"github.com/sevigo/shugosha/pkg/model"
)

func (m *BackupManager) updateRecord(providerName string, event model.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := providerName + ":" + event.Path
	slog.Debug("[BackupManager] update record in db", "providerName", providerName, "key", key)

	if err := m.saveEvent(key, event); err != nil {
		slog.Error("Failed to save event to DB", "error", err, "key", key)
	}

	m.updateTotalSize(providerName, event.Root, event.Size)
}

func (m *BackupManager) saveEvent(key string, event model.Event) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return m.db.Set(key, eventBytes)
}
