//go:build wireinject
// +build wireinject

package main

import (
	"fmt"

	"github.com/google/wire"

	"github.com/sevigo/shugosha/pkg/api"
	"github.com/sevigo/shugosha/pkg/backupmanager"
	"github.com/sevigo/shugosha/pkg/config"
	"github.com/sevigo/shugosha/pkg/db"
	"github.com/sevigo/shugosha/pkg/fsmonitor"
	"github.com/sevigo/shugosha/pkg/model"
	"github.com/sevigo/shugosha/pkg/provider"
)

// App contains all dependencies of the application.
type App struct {
	ConfigManager model.ConfigManager
	BackupManager *backupmanager.BackupManager
	Monitor       *fsmonitor.Monitor
	Server        *api.Server
}

// NewApp creates a new instance of your application
func NewApp(configManager model.ConfigManager, backupManager *backupmanager.BackupManager, monitor *fsmonitor.Monitor, server *api.Server) *App {
	return &App{
		ConfigManager: configManager,
		BackupManager: backupManager,
		Monitor:       monitor,
		Server:        server,
	}
}

func InitializeApp() (*App, error) {
	wire.Build(
		NewApp,
		configManagerProvider,
		backupProviders,
		backupManagerProvider,
		fsMonitorProvider,
		apiServiceProvider,
		dbProvider,
		backupConfigProvider,
		providerMetaInfoGetterProvider,
	)
	return &App{}, nil
}

func fsMonitorProvider() (*fsmonitor.Monitor, error) {
	monitor, err := fsmonitor.New(fsmonitor.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("Failed to setup file system monitor: %w", err)
	}
	return monitor, nil
}

func backupManagerProvider(storage model.DB, monitor *fsmonitor.Monitor, providers map[string]model.Provider) (*backupmanager.BackupManager, error) {
	backupManager, err := backupmanager.NewBackupManager(storage, monitor, providers)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup manager: %w", err)
	}
	return backupManager, nil
}

func dbProvider() (model.DB, error) {
	storage, err := db.NewBadgerDB(".db/")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	return storage, nil
}

func apiServiceProvider(cm model.ConfigManager, g model.ProviderMetaInfoGetter) *api.Server {
	return api.NewServer(cm, g)
}

func configManagerProvider(storage model.DB) (model.ConfigManager, error) {
	configManager, err := config.NewConfigManager(storage)
	if err != nil {
		return nil, fmt.Errorf("Config manager initialization failed: %w", err)
	}

	return configManager, nil
}

func backupProviders(backupConfig *model.BackupConfig) map[string]model.Provider {
	return provider.InitializeProviders(backupConfig)
}

func backupConfigProvider(configManager model.ConfigManager) (*model.BackupConfig, error) {
	return configManager.LoadConfig()
}

func providerMetaInfoGetterProvider(bm *backupmanager.BackupManager) model.ProviderMetaInfoGetter {
	return bm
}
