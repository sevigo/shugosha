package main

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/sevigo/shugosha/pkg/api"
	"github.com/sevigo/shugosha/pkg/backupmanager"
	"github.com/sevigo/shugosha/pkg/config"
	"github.com/sevigo/shugosha/pkg/db"
	"github.com/sevigo/shugosha/pkg/fsmonitor"
	"github.com/sevigo/shugosha/pkg/logger"
	"github.com/sevigo/shugosha/pkg/model"
	"github.com/sevigo/shugosha/pkg/provider"
)

const version = 0.2

func main() {
	logger.Setup()
	slog.Info("Starting 「Shugosha」 service", "version", version)

	// Initialize database
	storage, err := initializeDatabase()
	if err != nil {
		slog.Error("Database initialization failed", "error", err)
		return
	}
	defer storage.Close()

	// Initialize providers
	configManager, err := config.NewConfigManager(storage)
	if err != nil {
		slog.Error("Config manager initialization failed", "error", err)
		return
	}

	backupConfig, err := configManager.LoadConfig()
	if err != nil {
		slog.Error("Can't load configuration", "error", err)
		return
	}

	providers := provider.InitializeProviders(backupConfig)

	// Initialize backup manager
	backupManager, err := initializeBackupManager(storage, providers)
	if err != nil {
		slog.Error("Backup manager initialization failed", "error", err)
		return
	}
	defer backupManager.Close()

	// Setup and start file system monitor
	monitor, err := setupAndStartMonitor(providers, backupManager)
	if err != nil {
		slog.Error("File system monitoring initialization failed", "error", err)
		return
	}

	// Process backup results (optional)
	go processBackupResults(backupManager)

	if err := startAPIService(configManager); err != nil {
		slog.Error("Can't start API service", "error", err)
		return
	}

	// Wait for user input to stop the monitor
	waitForUserInput()

	// Stop the monitor
	stopMonitor(monitor)
}

// Create and start the API server.
func startAPIService(cm model.ConfigManager) error {
	server := api.NewServer(cm)
	slog.Debug("Starting API server on port 8080...")
	return server.Start(":8080")
}

func initializeDatabase() (*db.BadgerDB, error) {
	storage, err := db.NewBadgerDB(".db/")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return storage, nil
}

func initializeBackupManager(storage *db.BadgerDB, providers map[string]model.Provider) (*backupmanager.BackupManager, error) {
	backupManager, err := backupmanager.NewBackupManager(storage, providers)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup manager: %w", err)
	}
	return backupManager, nil
}

func setupAndStartMonitor(providers map[string]model.Provider, backupManager *backupmanager.BackupManager) (*fsmonitor.Monitor, error) {
	monitor, err := fsmonitor.New(fsmonitor.DefaultConfig())
	if err != nil {
		slog.Error("Failed to setup file system monitor", "error", err)
		return nil, err
	}

	monitor.Subscribe(backupManager)

	for _, provider := range providers {
		for _, path := range provider.DirectoryList() {
			absPath, err := filepath.Abs(path)
			if err != nil {
				slog.Error("Failed to get absolute path", "path", path, "error", err)
				continue
			}

			if err := monitor.Add(absPath); err != nil {
				slog.Error("Failed to add path to monitor", "path", absPath, "error", err)
			} else {
				slog.Info("Start monitoring", "path", absPath)
			}
		}
	}

	monitor.Start()
	return monitor, nil
}

func processBackupResults(backupManager *backupmanager.BackupManager) {
	for result := range backupManager.Results() {
		slog.Info("[MAIN] Backup resultn", "path", result.Path, "status", result.Status)
		if result.Status == "Failed" {
			slog.Error("Backup error", "error", result.Error)
		}
	}
}

func waitForUserInput() {
	fmt.Println("Press ENTER to stop monitoring...")
	fmt.Scanln()
}

func stopMonitor(monitor *fsmonitor.Monitor) {
	if err := monitor.Stop(); err != nil {
		slog.Error("Failed to stop monitor", "error", err)
	}
}
