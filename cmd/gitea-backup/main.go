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
	logger.Info("Starting backup process")
	
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
	
	if err := runBackup(settings, giteaConfig); err != nil {
		logger.Errorf("Backup failed: %v", err)
		os.Exit(1)
	}
	
	logger.Info("Backup process completed successfully")
}

func runBackup(settings *config.Settings, giteaConfig *config.GiteaConfig) error {
	// Clean temporary directories
	if err := files.CleanTmp(settings); err != nil {
		return fmt.Errorf("failed to clean temporary directories: %w", err)
	}
	
	// Backup database
	if err := database.BackupDatabase(settings, giteaConfig); err != nil {
		return fmt.Errorf("database backup failed: %w", err)
	}
	
	// Backup files
	if err := files.BackupFiles(settings, giteaConfig); err != nil {
		return fmt.Errorf("file backup failed: %w", err)
	}
	
	// Create zip archive
	if err := compression.CreateZip(settings); err != nil {
		return fmt.Errorf("failed to create zip archive: %w", err)
	}
	
	// Upload to remote storage
	if err := storage.Upload(settings); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	
	// Enforce retention policy
	if err := storage.EnsureMaxRetention(settings); err != nil {
		return fmt.Errorf("retention policy enforcement failed: %w", err)
	}
	
	// Add to history
	if err := history.Increment(settings, settings.BackupTmpRemoteFilename); err != nil {
		return fmt.Errorf("failed to update backup history: %w", err)
	}
	
	return nil
}