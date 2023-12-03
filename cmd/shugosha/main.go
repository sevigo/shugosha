package main

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/sevigo/shugosha/pkg/backupmanager"
	"github.com/sevigo/shugosha/pkg/logger"
)

const version = 0.2

func main() {
	logger.Setup()
	slog.Info("Starting 「Shugosha」 service", "version", version)

	app, err := InitializeApp()
	if err != nil {
		slog.Error("Failed to initialize applicationv", "error", err)
		return
	}

	// Start the file system monitor
	go func() {
		if err := app.Monitor.Start(); err != nil {
			slog.Error("Failed to start file system monitor", "error", err)
		}
	}()

	// Process backup results
	go processBackupResults(app.BackupManager)

	// Start the API server
	go func() {
		log.Println("Starting API server on port 8080...")
		if err := app.Server.Start(":8080"); err != nil {
			slog.Error("Failed to start API server", "error", err)
			return
		}
	}()

	// Wait for user input to stop the application
	fmt.Println("Press ENTER to stop...")
	fmt.Scanln()

	// Stop the monitor and other services before exiting
	if err := app.Monitor.Stop(); err != nil {
		slog.Error("Failed to stop monitor", "error", err)
	}

	slog.Info("「Shugosha」 service stopped")
}

func processBackupResults(manager *backupmanager.BackupManager) {
	for result := range manager.Results() {
		if result.Status == "Failed" {
			log.Printf("Backup failed for %s: %v", result.Path, result.Error)
		} else {
			log.Printf("Backup successful for %s", result.Path)
		}
	}
}
