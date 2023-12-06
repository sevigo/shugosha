package backupmanager

import (
	"context"
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
	providers  map[string]model.Provider
	resultChan chan BackupResult
	mu         sync.Mutex
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewBackupManager(storage model.DB, monitor *fsmonitor.Monitor, providers map[string]model.Provider) (*BackupManager, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	bm := &BackupManager{
		db:         storage,
		providers:  providers,
		resultChan: make(chan BackupResult, 10),
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}

	for _, rootDir := range monitor.RootDirs() {
		for _, provider := range providers {
			if isSubscribed(rootDir, provider) {
				bm.updateTotalSize(provider.Name(), rootDir, 0)
			}
		}
	}

	providerNames := extractProviderNames(providers)
	if err := bm.SetProviders(providerNames); err != nil {
		return nil, err
	}

	monitor.Subscribe(bm)

	return bm, nil
}

func extractProviderNames(providers map[string]model.Provider) []string {
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	return names
}

func (m *BackupManager) Results() <-chan BackupResult {
	return m.resultChan
}

func (m *BackupManager) HandleEvent(event model.Event) {
	slog.Debug("[manager] handle event", "file", event.Path)

	for name, provider := range m.providers {
		if isSubscribed(event.Root, provider) {
			go m.backupIfNeeded(event, name, provider)
		}
	}
}

func isSubscribed(root string, provider model.Provider) bool {
	slog.Debug("[manager] is subscribed", "root", root, "dirs", provider.DirectoryList())

	for _, dir := range provider.DirectoryList() {
		if dir == root {
			return true
		}
	}

	return false
}

func (m *BackupManager) backupIfNeeded(event model.Event, providerName string, provider model.Provider) {
	slog.Debug("[manager] backup if needed", "providerName", providerName, "file", event.Path)

	select {
	case <-m.ctx.Done():
		return
	default:
		if m.isBackupNeeded(event.Path, event.Checksum, providerName) {
			m.processBackup(event, provider)
		}
	}
}

func (m *BackupManager) processBackup(event model.Event, provider model.Provider) {
	slog.Debug("[manager] process backup", "providerName", provider.Name(), "file", event.Path)

	result := BackupResult{Path: event.Path, Status: "Success"}
	if err := provider.Backup(event); err != nil {
		result.Status = "Failed"
		result.Error = err.Error()
		slog.Error("Backup failed", "error", err, "path", event.Path)
	} else {
		m.updateRecord(provider.Name(), event)
	}
	
	m.resultChan <- result
}

func (m *BackupManager) isBackupNeeded(path, checksum, providerName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := providerName + ":" + path
	record, err := m.getRecord(key)
	if err != nil {
		return true
	}

	return record.Checksum != checksum
}

func (m *BackupManager) getRecord(key string) (*model.Event, error) {
	value, err := m.db.Get(key)
	if errors.Is(err, model.ErrDBKeyNotFound) {
		return &model.Event{}, nil
	} else if err != nil {
		slog.Error("[BackupManager] Error accessing DB", "error", err)
		return nil, err
	}

	var record model.Event
	if err := json.Unmarshal(value, &record); err != nil {
		slog.Error("[BackupManager] Error unmarshaling file record", "error", err)
		return nil, err
	}
	return &record, nil
}

func (m *BackupManager) Close() error {
	m.cancelFunc()
	close(m.resultChan)
	return m.db.Close()
}
