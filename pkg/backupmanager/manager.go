package backupmanager

import (
	"encoding/json"
	"errors"
	"log/slog"
	"sync"

	"github.com/sevigo/shugosha/pkg/fsmonitor"
	"github.com/sevigo/shugosha/pkg/model"
)

type BackupResult struct {
	Path   string
	Status string
	Error  string
}

type BackupManager struct {
	db         model.DB
	providers  map[string]model.Provider // this is an interface
	resultChan chan BackupResult
	mu         sync.Mutex // Mutex for thread-safe operations
}

// NewBackupManager initializes a new BackupManager with the given providers and database path.
func NewBackupManager(storage model.DB, monitor *fsmonitor.Monitor, providers map[string]model.Provider) (*BackupManager, error) {
	slog.Debug("BackupManager initialized")

	bm := &BackupManager{
		db:         storage,
		providers:  providers,
		resultChan: make(chan BackupResult, 10),
	}

	// Extract provider names from the providers map
	providerNames := make([]string, 0, len(providers))
	for name, provider := range providers {
		providerNames = append(providerNames, name)
		for _, rootDir := range provider.DirectoryList() {
			bm.updateTotalSize(name, rootDir, 0)
			monitor.Add(rootDir)
		}
	}

	// Store the provider names using SetProviders method
	if err := bm.SetProviders(providerNames); err != nil {
		return nil, err
	}

	monitor.Subscribe(bm)

	return bm, nil
}

func (m *BackupManager) Results() <-chan BackupResult {
	return m.resultChan
}

// HandleEvent handles filesystem events and initiates backups if needed.
func (m *BackupManager) HandleEvent(event model.Event) {
	slog.Debug("[BackupManager] HandleEvent", "type", event.Type, "path", event.Path)

	for _, provider := range m.providers {
		if !isSubscribed(event.Root, provider.DirectoryList()) {
			continue
		}

		// handle event by registered provider
		go m.handleProviderBackup(event, provider)
	}
}

func (m *BackupManager) handleProviderBackup(event model.Event, provider model.Provider) {
	// check if the file is stored
	if m.isBackupNeeded(event.Path, event.Checksum, provider.Name()) {
		slog.Debug("[BackupManager] Backup needed", "path", event.Path, "provider", provider.Name())

		// create backup and communicate the result back
		result := BackupResult{Path: event.Path, Status: "Success"}
		if err := provider.Backup(event); err != nil {
			result.Status = "Failed"
			result.Error = err.Error()
			slog.Error("Backup failed", "error", err, "path", event.Path)
		} else {
			m.updateRecord(provider.Name(), &event)
		}

		m.resultChan <- result
	} else {
		slog.Debug("[BackupManager] No backup needed", "path", event.Path, "provider", provider.Name())
	}
}

func isSubscribed(root string, directoryList []string) bool {
	for _, dir := range directoryList {
		if dir == root {
			return true
		}
	}

	return false
}

// isBackupNeeded checks if a backup is necessary for a given file path and provider.
func (m *BackupManager) isBackupNeeded(path, checksum, providerName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	value, err := m.db.Get(path)
	if err != nil {
		// Handle not found as a need for backup
		if errors.Is(err, model.ErrDBKeyNotFound) {
			return true
		}
		slog.Error("[BackupManager] Error accessing DB", "error", err)
		return true // Assume backup needed on other errors
	}

	var record model.FileRecord
	if err := json.Unmarshal(value, &record); err != nil {
		slog.Error("[BackupManager] Error unmarshaling file record", "error", err)
		return true // Assume backup needed on unmarshal error
	}

	storedChecksum := record.ProviderData[providerName]
	return storedChecksum != checksum
}

func (m *BackupManager) Close() error {
	close(m.resultChan) // Ensure the results channel is closed
	return m.db.Close()
}
