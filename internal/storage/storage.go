package storage

import (
	"fmt"
	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// StorageBackend defines the interface for remote storage operations
type StorageBackend interface {
	Upload(settings *config.Settings) error
	Download(settings *config.Settings) error
	EnsureMaxRetention(settings *config.Settings) error
	ValidateConfig() error
}

// GetBackend returns the appropriate storage backend based on the backup method
func GetBackend(method string) (StorageBackend, error) {
	switch method {
	case "s3":
		return &S3Backend{}, nil
	case "ftp":
		return &FTPBackend{}, nil
	default:
		return nil, fmt.Errorf("unsupported storage backend: %s", method)
	}
}

// Upload uploads the backup file to remote storage
func Upload(settings *config.Settings) error {
	logger.Infof("Starting upload to %s", settings.BackupMethod)
	
	backend, err := GetBackend(settings.BackupMethod)
	if err != nil {
		return err
	}
	
	if err := backend.ValidateConfig(); err != nil {
		return fmt.Errorf("storage configuration validation failed: %w", err)
	}
	
	if err := backend.Upload(settings); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	
	logger.Info("Upload completed successfully")
	return nil
}

// Download downloads the backup file from remote storage
func Download(settings *config.Settings) error {
	logger.Infof("Starting download from %s", settings.BackupMethod)
	
	backend, err := GetBackend(settings.BackupMethod)
	if err != nil {
		return err
	}
	
	if err := backend.ValidateConfig(); err != nil {
		return fmt.Errorf("storage configuration validation failed: %w", err)
	}
	
	if err := backend.Download(settings); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	
	logger.Info("Download completed successfully")
	return nil
}

// EnsureMaxRetention ensures the maximum retention policy is enforced
func EnsureMaxRetention(settings *config.Settings) error {
	if settings.BackupMaxRetention <= 0 {
		logger.Debug("Retention policy disabled, skipping cleanup")
		return nil
	}
	
	logger.Infof("Enforcing retention policy (max %d backups)", settings.BackupMaxRetention)
	
	backend, err := GetBackend(settings.BackupMethod)
	if err != nil {
		return err
	}
	
	if err := backend.ValidateConfig(); err != nil {
		return fmt.Errorf("storage configuration validation failed: %w", err)
	}
	
	if err := backend.EnsureMaxRetention(settings); err != nil {
		return fmt.Errorf("retention cleanup failed: %w", err)
	}
	
	logger.Info("Retention policy enforced successfully")
	return nil
}