package backupmanager

import (
	"encoding/json"
	"errors"
	"log/slog"
	"sync"

	"github.com/dgraph-io/badger/v4"

	"github.com/sevigo/shugosha/pkg/db"
	"github.com/sevigo/shugosha/pkg/fsmonitor"
	"github.com/sevigo/shugosha/pkg/model"
)

type BackupResult struct {
	Path   string
	Status string
	Error  string
}

type BackupManager struct {
	db         db.DB
	providers  map[string]model.Provider // this is an interface
	resultChan chan BackupResult
	mu         sync.Mutex // Mutex for thread-safe operations
}

// NewBackupManager initializes a new BackupManager with the given providers and database path.
func NewBackupManager(storage db.DB, providers map[string]model.Provider) (*BackupManager, error) {
	slog.Debug("BackupManager initialized")
	return &BackupManager{
		db:         storage,
		providers:  providers,
		resultChan: make(chan BackupResult, 10),
	}, nil
}

func (m *BackupManager) Results() <-chan BackupResult {
	return m.resultChan
}

// HandleEvent handles filesystem events and initiates backups if needed.
func (m *BackupManager) HandleEvent(event fsmonitor.Event) {
	slog.Debug("[BackupManager] HandleEvent", "type", event.Type, "path", event.Path)

	for _, provider := range m.providers {
		// handle event by registered provider
		go func(provider model.Provider) {

			// check if the file is stored
			if m.isBackupNeeded(event.Path, event.Checksum, provider.Name()) {
				slog.Debug("[BackupManager] Backup needed", "path", event.Path, "provider", provider.Name())

				// create backup and communicate the result back
				result := BackupResult{Path: event.Path, Status: "Success"}
				if err := provider.Backup(event.Path); err != nil {
					result.Status = "Failed"
					result.Error = err.Error()
					slog.Error("Backup failed", "error", err, "path", event.Path)
				} else {
					m.updateRecord(event.Path, event.Checksum, provider.Name())
				}

				m.resultChan <- result
			} else {
				slog.Debug("[BackupManager] No backup needed", "path", event.Path, "provider", provider.Name())
			}
		}(provider)
	}
}

// isBackupNeeded checks if a backup is necessary for a given file path and provider.
func (m *BackupManager) isBackupNeeded(path, checksum, providerName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	value, err := m.db.Get(path)
	if err != nil {
		// Handle not found as a need for backup
		if errors.Is(err, badger.ErrKeyNotFound) {
			return true
		}
		slog.Error("[BackupManager] Error accessing DB", "error", err)
		return true // Assume backup needed on other errors
	}

	var record FileRecord
	if err := json.Unmarshal(value, &record); err != nil {
		slog.Error("[BackupManager] Error unmarshaling file record", "error", err)
		return true // Assume backup needed on unmarshal error
	}

	storedChecksum := record.ProviderData[providerName]
	return storedChecksum != checksum
}

func (m *BackupManager) updateRecord(path, checksum, providerName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Retrieve the current file record or create a new one
	var record FileRecord
	value, err := m.db.Get(path)
	if err != nil {
		// If an error other than not found, log and exit
		if errors.Is(err, badger.ErrKeyNotFound) {
			slog.Error("Failed to retrieve record from DB", "error", err, "path", path)
			return
		}
		// If the record is not found, initialize a new one
		record = FileRecord{
			ProviderData: make(map[string]string),
		}
	} else {
		// Unmarshal the existing record
		if err := json.Unmarshal(value, &record); err != nil {
			slog.Error("Failed to unmarshal file record", "error", err, "path", path)
			return
		}
	}

	// Update the provider data with the new checksum
	record.ProviderData[providerName] = checksum

	// Serialize the updated record
	recordBytes, err := json.Marshal(record)
	if err != nil {
		slog.Error("Failed to marshal file record", "error", err, "path", path)
		return
	}

	// Update the record in the database
	if err := m.db.Set(path, recordBytes); err != nil {
		slog.Error("Failed to update file record in DB", "error", err, "path", path)
	} else {
		slog.Debug("Successfully updated file record in DB", "path", path)
	}
}

func (m *BackupManager) Close() error {
	return m.db.Close()
}
