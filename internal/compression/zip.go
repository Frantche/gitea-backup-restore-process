package compression

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// CreateZip creates a zip archive from the backup tmp folder
func CreateZip(settings *config.Settings) error {
	logger.Info("Creating zip archive")
	
	zipFile, err := os.Create(settings.BackupTmpFilename)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()
	
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	
	err = filepath.Walk(settings.BackupTmpFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip the root folder itself
		if path == settings.BackupTmpFolder {
			return nil
		}
		
		// Get relative path
		relPath, err := filepath.Rel(settings.BackupTmpFolder, path)
		if err != nil {
			return err
		}
		
		// Normalize path separators for zip
		relPath = strings.ReplaceAll(relPath, "\\", "/")
		
		if info.IsDir() {
			// Add directory to zip (with trailing slash)
			_, err := zipWriter.Create(relPath + "/")
			return err
		}
		
		// Add file to zip
		fileWriter, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}
		
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		
		_, err = io.Copy(fileWriter, file)
		return err
	})
	
	if err != nil {
		return fmt.Errorf("failed to create zip archive: %w", err)
	}
	
	logger.Info("Zip archive created successfully")
	return nil
}

// ExtractZip extracts a zip archive to the restore tmp folder
func ExtractZip(settings *config.Settings) error {
	logger.Info("Extracting zip archive")
	
	// Create restore tmp folder
	if err := os.MkdirAll(settings.RestoreTmpFolder, 0755); err != nil {
		return fmt.Errorf("failed to create restore tmp folder: %w", err)
	}
	
	zipReader, err := zip.OpenReader(settings.RestoreTmpFilename)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zipReader.Close()
	
	// Extract files
	for _, file := range zipReader.File {
		err := extractFile(file, settings.RestoreTmpFolder)
		if err != nil {
			return fmt.Errorf("failed to extract file %s: %w", file.Name, err)
		}
	}
	
	logger.Info("Zip archive extracted successfully")
	return nil
}

// extractFile extracts a single file from the zip archive
func extractFile(file *zip.File, destDir string) error {
	// Normalize path and prevent directory traversal
	destPath := filepath.Join(destDir, file.Name)
	if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", file.Name)
	}
	
	if file.FileInfo().IsDir() {
		return os.MkdirAll(destPath, file.FileInfo().Mode())
	}
	
	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	
	// Open file in zip
	zipFile, err := file.Open()
	if err != nil {
		return err
	}
	defer zipFile.Close()
	
	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	// Copy content
	_, err = io.Copy(destFile, zipFile)
	if err != nil {
		return err
	}
	
	// Set file permissions
	return os.Chmod(destPath, file.FileInfo().Mode())
}