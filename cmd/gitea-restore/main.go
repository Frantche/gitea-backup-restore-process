package main

import (
	"fmt"
	"os"

	"github.com/Frantche/gitea-backup-restore-process/internal/compression"
	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/internal/database"
	"github.com/Frantche/gitea-backup-restore-process/internal/files"
	"github.com/Frantche/gitea-backup-restore-process/internal/history"
	"github.com/Frantche/gitea-backup-restore-process/internal/storage"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

func main() {
	logger.Info("Starting restore process")
	
	// Load configuration
	settings, err := config.NewSettings()
	if err != nil {
		logger.Errorf("Failed to load configuration: %v", err)
		os.Exit(1)
	}
	
	// Read Gitea configuration
	giteaConfig, err := config.ReadGiteaConfig(settings.AppIniPath)
	if err != nil {
		logger.Errorf("Failed to read Gitea configuration: %v", err)
		os.Exit(1)
	}
	
	logger.Debugf("Settings: %+v", settings)
	logger.Debugf("Gitea config: %+v", giteaConfig)
	
	if err := runRestore(settings, giteaConfig); err != nil {
		logger.Errorf("Restore failed: %v", err)
		os.Exit(1)
	}
	
	logger.Info("Restore process completed successfully")
}

func runRestore(settings *config.Settings, giteaConfig *config.GiteaConfig) error {
	// Check if restore has already been performed for this backup
	alreadyRestored, err := history.Check(settings)
	if err != nil {
		return fmt.Errorf("failed to check restore history: %w", err)
	}
	
	if alreadyRestored {
		logger.Info("Gitea has already been restored with this file version")
		return nil
	}
	
	// Download from remote storage
	if err := storage.Download(settings); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	
	// Extract zip archive
	if err := compression.ExtractZip(settings); err != nil {
		return fmt.Errorf("failed to extract zip archive: %w", err)
	}
	
	// Restore files
	if err := files.RestoreFiles(settings, giteaConfig); err != nil {
		return fmt.Errorf("file restore failed: %w", err)
	}
	
	// Restore database
	if err := database.RestoreDatabase(settings, giteaConfig); err != nil {
		return fmt.Errorf("database restore failed: %w", err)
	}
	
	// Add to history
	if err := history.Increment(settings, settings.BackupFilename); err != nil {
		return fmt.Errorf("failed to update restore history: %w", err)
	}
	
	return nil
}