package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sevigo/shugosha/pkg/backupmanager"
	"github.com/sevigo/shugosha/pkg/logger"
)

const version = 0.3

func main() {
	logger.Setup()
	slog.Info("Starting 「Shugosha」 service", "version", version)

	app, err := InitializeApp()
	if err != nil {
		slog.Error("Failed to initialize application", "error", err)
		return
	}

	// Create a context that is cancelled on program termination
	ctx, cancel := context.WithCancel(context.Background())

	// Setting up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the file system monitor with context
	go func() {
		if err := app.Monitor.Start(ctx); err != nil {
			slog.Error("Failed to start file system monitor", "error", err)
		}
	}()

	// Process backup results
	go processBackupResults(ctx, app.BackupManager)

	// Start the API server with context
	go func() {
		log.Println("Starting API server on port 8080...")
		if err := app.Server.Start(ctx, ":8080"); err != nil {
			slog.Error("Failed to start API server", "error", err)
			return
		}
	}()

	// Wait for termination signal
	<-sigChan
	cancel() // Cancels the context

	// Wait for a moment to allow goroutines to finish gracefully
	// ...

	slog.Info("「Shugosha」 service stopped")
}

func processBackupResults(ctx context.Context, manager *backupmanager.BackupManager) {
	for {
		select {
		case result := <-manager.Results():
			if result.Status == "Failed" {
				log.Printf("Backup failed for %s: %v", result.Path, result.Error)
			} else {
				log.Printf("Backup successful for %s", result.Path)
			}
		case <-ctx.Done():
			log.Println("Shutting down backup results processing...")
			return
		}
	}
}
