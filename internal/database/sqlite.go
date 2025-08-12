package database

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// SQLiteAdapter implements DatabaseAdapter for SQLite
type SQLiteAdapter struct{}

func (s *SQLiteAdapter) Backup(settings *config.Settings, giteaConfig *config.GiteaConfig) error {
	// For SQLite, we simply copy the database file
	sourcePath := giteaConfig.Database.Path
	if sourcePath == "" {
		return fmt.Errorf("SQLite database path not configured")
	}
	
	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("SQLite database file does not exist: %s", sourcePath)
	}
	
	outputFile := filepath.Join(settings.BackupTmpFolder, "dump.sqlite3.db")
	
	logger.Debugf("Copying SQLite database from %s to %s", sourcePath, outputFile)
	
	if err := copyFile(sourcePath, outputFile); err != nil {
		return fmt.Errorf("failed to copy SQLite database: %w", err)
	}
	
	logger.Info("SQLite database backup completed")
	return nil
}

func (s *SQLiteAdapter) Restore(settings *config.Settings, giteaConfig *config.GiteaConfig) error {
	// For SQLite, we copy the database file back
	targetPath := giteaConfig.Database.Path
	if targetPath == "" {
		return fmt.Errorf("SQLite database path not configured")
	}
	
	inputFile := filepath.Join(settings.RestoreTmpFolder, "dump.sqlite3.db")
	
	// Check if backup file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("SQLite backup file does not exist: %s", inputFile)
	}
	
	logger.Debugf("Copying SQLite database from %s to %s", inputFile, targetPath)
	
	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	
	// Remove existing database file if it exists
	if _, err := os.Stat(targetPath); err == nil {
		if err := os.Remove(targetPath); err != nil {
			return fmt.Errorf("failed to remove existing database file: %w", err)
		}
	}
	
	if err := copyFile(inputFile, targetPath); err != nil {
		return fmt.Errorf("failed to restore SQLite database: %w", err)
	}
	
	logger.Info("SQLite database restore completed")
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	
	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	return os.Chmod(dst, sourceInfo.Mode())
}