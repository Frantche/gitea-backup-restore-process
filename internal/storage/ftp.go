package storage

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"

	appconfig "github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// FTPBackend implements StorageBackend for FTP
type FTPBackend struct{}

// FTPConfig holds FTP-specific configuration
type FTPConfig struct {
	Host     string
	User     string
	Password string
	Dir      string
}

// getFTPConfig reads FTP configuration from environment variables
func getFTPConfig() (*FTPConfig, error) {
	return &FTPConfig{
		Host:     os.Getenv("BACKUP_FTP_HOST"),
		User:     os.Getenv("BACKUP_FTP_USER"),
		Password: os.Getenv("BACKUP_FTP_PASSWORD"),
		Dir:      os.Getenv("BACKUP_FTP_DIR"),
	}, nil
}

func (f *FTPBackend) ValidateConfig() error {
	ftpConfig, err := getFTPConfig()
	if err != nil {
		return err
	}
	
	if ftpConfig.Host == "" {
		return fmt.Errorf("BACKUP_FTP_HOST is required for FTP backend")
	}
	if ftpConfig.User == "" {
		return fmt.Errorf("BACKUP_FTP_USER is required for FTP backend")
	}
	if ftpConfig.Password == "" {
		return fmt.Errorf("BACKUP_FTP_PASSWORD is required for FTP backend")
	}
	
	return nil
}

func (f *FTPBackend) Upload(settings *appconfig.Settings) error {
	ftpConfig, err := getFTPConfig()
	if err != nil {
		return err
	}
	
	// Connect to FTP server
	conn, err := ftp.Dial(ftpConfig.Host, ftp.DialWithTimeout(30*time.Second))
	if err != nil {
		return fmt.Errorf("failed to connect to FTP server: %w", err)
	}
	defer conn.Quit()
	
	// Login
	err = conn.Login(ftpConfig.User, ftpConfig.Password)
	if err != nil {
		return fmt.Errorf("failed to login to FTP server: %w", err)
	}
	
	// Change to target directory if specified
	if ftpConfig.Dir != "" {
		err = conn.ChangeDir(ftpConfig.Dir)
		if err != nil {
			return fmt.Errorf("failed to change to directory %s: %w", ftpConfig.Dir, err)
		}
	}
	
	// Open local file
	file, err := os.Open(settings.BackupTmpFilename)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()
	
	// Upload file
	err = conn.Stor(settings.BackupTmpRemoteFilename, file)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	
	logger.Info("Upload to FTP successful")
	return nil
}

func (f *FTPBackend) Download(settings *appconfig.Settings) error {
	ftpConfig, err := getFTPConfig()
	if err != nil {
		return err
	}
	
	// Connect to FTP server
	conn, err := ftp.Dial(ftpConfig.Host, ftp.DialWithTimeout(30*time.Second))
	if err != nil {
		return fmt.Errorf("failed to connect to FTP server: %w", err)
	}
	defer conn.Quit()
	
	// Login
	err = conn.Login(ftpConfig.User, ftpConfig.Password)
	if err != nil {
		return fmt.Errorf("failed to login to FTP server: %w", err)
	}
	
	// Change to target directory if specified
	if ftpConfig.Dir != "" {
		err = conn.ChangeDir(ftpConfig.Dir)
		if err != nil {
			return fmt.Errorf("failed to change to directory %s: %w", ftpConfig.Dir, err)
		}
	}
	
	// Get remote filename from environment variable
	remoteFilename := os.Getenv("BACKUP_FILENAME")
	if remoteFilename == "" {
		return fmt.Errorf("BACKUP_FILENAME is required for FTP download")
	}
	
	// Download file
	resp, err := conn.Retr(remoteFilename)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Close()
	
	// Create local file
	file, err := os.Create(settings.RestoreTmpFilename)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()
	
	// Copy downloaded content to local file
	_, err = io.Copy(file, resp)
	if err != nil {
		return fmt.Errorf("failed to write downloaded content: %w", err)
	}
	
	logger.Info("Download from FTP successful")
	return nil
}

func (f *FTPBackend) EnsureMaxRetention(settings *appconfig.Settings) error {
	if settings.BackupMaxRetention <= 0 {
		return nil
	}
	
	ftpConfig, err := getFTPConfig()
	if err != nil {
		return err
	}
	
	// Connect to FTP server
	conn, err := ftp.Dial(ftpConfig.Host, ftp.DialWithTimeout(30*time.Second))
	if err != nil {
		return fmt.Errorf("failed to connect to FTP server: %w", err)
	}
	defer conn.Quit()
	
	// Login
	err = conn.Login(ftpConfig.User, ftpConfig.Password)
	if err != nil {
		return fmt.Errorf("failed to login to FTP server: %w", err)
	}
	
	// Change to target directory if specified
	if ftpConfig.Dir != "" {
		err = conn.ChangeDir(ftpConfig.Dir)
		if err != nil {
			return fmt.Errorf("failed to change to directory %s: %w", ftpConfig.Dir, err)
		}
	}
	
	// List files
	entries, err := conn.List(".")
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}
	
	// Filter files with backup prefix and sort by time
	var backupFiles []struct {
		name string
		time time.Time
	}
	
	for _, entry := range entries {
		if entry.Type == ftp.EntryTypeFile && strings.HasPrefix(entry.Name, settings.BackupPrefix) {
			backupFiles = append(backupFiles, struct {
				name string
				time time.Time
			}{entry.Name, entry.Time})
		}
	}
	
	// Sort files by time (newest first)
	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].time.After(backupFiles[j].time)
	})
	
	// Delete old files beyond retention limit
	if len(backupFiles) > settings.BackupMaxRetention {
		filesToDelete := backupFiles[settings.BackupMaxRetention:]
		
		for _, fileInfo := range filesToDelete {
			// Don't delete the current backup file if it exists
			backupFilename := os.Getenv("BACKUP_FILENAME")
			if backupFilename != "" && fileInfo.name == backupFilename {
				continue
			}
			
			err = conn.Delete(fileInfo.name)
			if err != nil {
				logger.Errorf("Failed to delete file %s: %v", fileInfo.name, err)
				continue
			}
			
			logger.Infof("Deleted old backup from FTP: %s", fileInfo.name)
		}
	}
	
	return nil
}