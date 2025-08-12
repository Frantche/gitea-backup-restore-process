package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// CleanTmp cleans temporary directories and creates them fresh
func CleanTmp(settings *config.Settings) error {
	logger.Debug("Cleaning temporary directories")
	
	// Remove and recreate backup tmp folder
	if err := os.RemoveAll(settings.BackupTmpFolder); err != nil {
		return fmt.Errorf("failed to remove backup tmp folder: %w", err)
	}
	if err := os.MkdirAll(settings.BackupTmpFolder, 0755); err != nil {
		return fmt.Errorf("failed to create backup tmp folder: %w", err)
	}
	
	// Remove backup tmp file if exists
	if err := os.Remove(settings.BackupTmpFilename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove backup tmp file: %w", err)
	}
	
	// Remove and recreate restore tmp folder
	if err := os.RemoveAll(settings.RestoreTmpFolder); err != nil {
		return fmt.Errorf("failed to remove restore tmp folder: %w", err)
	}
	
	// Remove restore tmp file if exists
	if err := os.Remove(settings.RestoreTmpFilename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove restore tmp file: %w", err)
	}
	
	logger.Debug("Temporary directories cleaned successfully")
	return nil
}

// BackupFiles backs up Gitea files (repositories, avatars, etc.)
func BackupFiles(settings *config.Settings, giteaConfig *config.GiteaConfig) error {
	logger.Info("Starting file backup")
	
	// Backup repositories
	if giteaConfig.Repository.Root != "" {
		targetDir := filepath.Join(settings.BackupTmpFolder, "repo")
		if err := copyDir(giteaConfig.Repository.Root, targetDir); err != nil {
			logger.Errorf("Failed to backup repositories: %v", err)
		} else {
			logger.Debug("Repositories backed up successfully")
		}
	}
	
	// Backup avatars
	if giteaConfig.Picture.AvatarUploadPath != "" {
		targetDir := filepath.Join(settings.BackupTmpFolder, "avatars")
		if err := copyDir(giteaConfig.Picture.AvatarUploadPath, targetDir); err != nil {
			logger.Errorf("Failed to backup avatars: %v", err)
		} else {
			logger.Debug("Avatars backed up successfully")
		}
	}
	
	// Backup repository avatars
	if giteaConfig.Picture.RepositoryAvatarUploadPath != "" {
		targetDir := filepath.Join(settings.BackupTmpFolder, "repo-avatars")
		if err := copyDir(giteaConfig.Picture.RepositoryAvatarUploadPath, targetDir); err != nil {
			logger.Errorf("Failed to backup repository avatars: %v", err)
		} else {
			logger.Debug("Repository avatars backed up successfully")
		}
	}
	
	logger.Info("File backup completed")
	return nil
}

// RestoreFiles restores Gitea files from backup
func RestoreFiles(settings *config.Settings, giteaConfig *config.GiteaConfig) error {
	logger.Info("Starting file restore")
	
	// Restore repositories
	sourceDir := filepath.Join(settings.RestoreTmpFolder, "repo")
	if giteaConfig.Repository.Root != "" {
		if err := copyDir(sourceDir, giteaConfig.Repository.Root); err != nil {
			logger.Errorf("Failed to restore repositories: %v", err)
		} else {
			logger.Debug("Repositories restored successfully")
		}
	}
	
	// Restore avatars
	sourceDir = filepath.Join(settings.RestoreTmpFolder, "avatars")
	if giteaConfig.Picture.AvatarUploadPath != "" {
		if err := copyDir(sourceDir, giteaConfig.Picture.AvatarUploadPath); err != nil {
			logger.Errorf("Failed to restore avatars: %v", err)
		} else {
			logger.Debug("Avatars restored successfully")
		}
	}
	
	// Restore repository avatars
	sourceDir = filepath.Join(settings.RestoreTmpFolder, "repo-avatars")
	if giteaConfig.Picture.RepositoryAvatarUploadPath != "" {
		if err := copyDir(sourceDir, giteaConfig.Picture.RepositoryAvatarUploadPath); err != nil {
			logger.Errorf("Failed to restore repository avatars: %v", err)
		} else {
			logger.Debug("Repository avatars restored successfully")
		}
	}
	
	logger.Info("File restore completed")
	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Check if source exists
	srcInfo, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Debugf("Source directory does not exist: %s", src)
			return nil // Not an error if source doesn't exist
		}
		return err
	}
	
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}
	
	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	
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