package main

import (
	"fmt"
	"log"
	"log/slog"
	"path/filepath"

	"github.com/sevigo/shugosha/pkg/backup"
	"github.com/sevigo/shugosha/pkg/backupmanager"
	"github.com/sevigo/shugosha/pkg/config"
	"github.com/sevigo/shugosha/pkg/db"
	"github.com/sevigo/shugosha/pkg/fsmonitor"
	"github.com/sevigo/shugosha/pkg/logger"
)

const version = 0.2

// Example configuration
var backupConfig = config.BackupConfig{
	Providers: []config.ProviderConfig{
		{
			Name:     "LocalBackup",
			Type:     "Local",
			Settings: map[string]string{"path": "/local/backup"},
		},
		{
			Name:     "CloudBackup",
			Type:     "AWS",
			Settings: map[string]string{
				"accessKey": "key", 
				"secretKey": "secret",
			},
		},
	},
	DirectoryList: []string{
		"/path/to/dir1", 
		"/path/to/dir2",
	},
	DirectoryMap: map[string]string{
		"/path/to/dir1": "LocalBackup",
		"/path/to/dir2": "CloudBackup",
	},
}

func init() {
	logger.Setup()
}

func main() {
	slog.Info("Starting 「Shugosha」 service", "version", version)

	// Initialize backup providers
	dummyProvider := backup.NewDummyProvider() // Example provider
	providers := map[string]backup.Provider{
		"Dummy": dummyProvider,
	}

	// Initialize the database
	db, err := db.NewBadgerDB(".db/")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize the backup manager with Badger DB path and providers
	backupManager, err := backupmanager.NewBackupManager(db, providers)
	if err != nil {
		log.Fatalf("Failed to create backup manager: %v", err)
	}
	defer backupManager.Close()

	monitor, err := fsmonitor.New(fsmonitor.DefaultConfig())
	if err != nil {
		panic(err)
	}
	defer monitor.Stop()

	// Subscribe the backup manager to fsmonitor events
	monitor.Subscribe(backupManager)

	monitor.Start()

	// Add directories to the watch list
	pathsToWatch := []string{`C:\Users\igork\Test`}
	for _, path := range pathsToWatch {
		absPath, err := filepath.Abs(path)
		if err != nil {
			slog.Error("Failed to get absolute path", "path", path, "error", err)
			continue
		}

		err = monitor.Add(absPath)
		if err != nil {
			slog.Error("Failed to add path to monitor", "path", absPath, "error", err)
		} else {
			slog.Info("Start monitoring", "path", absPath)
		}
	}

	// Process backup results (optional)
	go func() {
		for result := range backupManager.Results() {
			// Handle backup results, e.g., log them
			log.Printf("Backup result for %s: %s\n", result.Path, result.Status)
			if result.Status == "Failed" {
				log.Printf("Backup error: %s\n", result.Error)
			}
		}
	}()

	// Wait for user input to stop the monitor (for demonstration purposes)
	fmt.Println("Press ENTER to stop monitoring...")
	fmt.Scanln()

	// Stop the monitor
	if err := monitor.Stop(); err != nil {
		slog.Error("Failed to stop monitor", "error", err)
	}
}
