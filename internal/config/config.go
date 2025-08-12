package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Settings represents the configuration for gitea backup/restore
type Settings struct {
	BackupEnable             bool   `yaml:"backup_enable"`
	BackupMethod             string `yaml:"backup_method"`
	BackupFilename           string `yaml:"backup_filename,omitempty"`
	BackupFileLog            string `yaml:"backup_file_log"`
	BackupTmpRemoteFilename  string `yaml:"backup_tmp_remote_filename"`
	BackupPrefix             string `yaml:"backup_prefix"`
	BackupMaxRetention       int    `yaml:"backup_max_retention"`
	BackupTmpFolder          string `yaml:"backup_tmp_folder"`
	BackupTmpFilename        string `yaml:"backup_tmp_filename"`
	RestoreTmpFolder         string `yaml:"restore_tmp_folder"`
	RestoreTmpFilename       string `yaml:"restore_tmp_filename"`
	AppIniPath               string `yaml:"app_ini_path"`
}

// NewSettings creates a new Settings instance with default values and environment overrides
func NewSettings() (*Settings, error) {
	settings := &Settings{
		BackupFileLog:           "/data/backupFileLog.txt",
		BackupTmpRemoteFilename: "@prefix-@date.zip",
		BackupPrefix:            "gitea-backup",
		BackupMaxRetention:      5,
		BackupTmpFolder:         "/tmp/backup",
		BackupTmpFilename:       "/tmp/backup.zip",
		RestoreTmpFolder:        "/tmp/restore",
		RestoreTmpFilename:      "/tmp/restore.zip",
		AppIniPath:              "/data/gitea/conf/app.ini",
	}

	// Load from environment variables
	if err := settings.loadFromEnv(); err != nil {
		return nil, err
	}

	// Process template strings
	settings.processTemplates()

	// Validate required fields
	if err := settings.validate(); err != nil {
		return nil, err
	}

	return settings, nil
}

func (s *Settings) loadFromEnv() error {
	if val := os.Getenv("BACKUP_ENABLE"); val != "" {
		enable, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid BACKUP_ENABLE: %w", err)
		}
		s.BackupEnable = enable
	}

	if val := os.Getenv("BACKUP_METHODE"); val != "" { // Keep original typo for compatibility
		s.BackupMethod = val
	}
	if val := os.Getenv("BACKUP_METHOD"); val != "" { // Also support correct spelling
		s.BackupMethod = val
	}

	if val := os.Getenv("BACKUP_FILENAME"); val != "" {
		s.BackupFilename = val
	}

	if val := os.Getenv("BACKUP_FILE_LOG"); val != "" {
		s.BackupFileLog = val
	}

	if val := os.Getenv("BACKUP_TMP_REMOTE_FILENAME"); val != "" {
		s.BackupTmpRemoteFilename = val
	}

	if val := os.Getenv("BACKUP_PREFIX"); val != "" {
		s.BackupPrefix = val
	}

	if val := os.Getenv("BACKUP_MAX_RETENTION"); val != "" {
		retention, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid BACKUP_MAX_RETENTION: %w", err)
		}
		s.BackupMaxRetention = retention
	}

	if val := os.Getenv("BACKUP_TMP_FOLDER"); val != "" {
		s.BackupTmpFolder = val
	}

	if val := os.Getenv("BACKUP_TMP_FILENAME"); val != "" {
		s.BackupTmpFilename = val
	}

	if val := os.Getenv("RESTORE_TMP_FOLDER"); val != "" {
		s.RestoreTmpFolder = val
	}

	if val := os.Getenv("RESTORE_TMP_FILENAME"); val != "" {
		s.RestoreTmpFilename = val
	}

	if val := os.Getenv("APP_INI_PATH"); val != "" {
		s.AppIniPath = val
	}

	return nil
}

func (s *Settings) processTemplates() {
	// Replace template placeholders in backup filename
	now := time.Now()
	dateStr := now.Format("2006-01-02-15-04-05")
	
	s.BackupTmpRemoteFilename = strings.ReplaceAll(s.BackupTmpRemoteFilename, "@prefix", s.BackupPrefix)
	s.BackupTmpRemoteFilename = strings.ReplaceAll(s.BackupTmpRemoteFilename, "@date", dateStr)
}

func (s *Settings) validate() error {
	// Validate backup method
	supportedMethods := []string{"s3", "ftp"}
	validMethod := false
	for _, method := range supportedMethods {
		if s.BackupMethod == method {
			validMethod = true
			break
		}
	}
	if !validMethod {
		return fmt.Errorf("invalid backup method '%s', supported methods: %v", s.BackupMethod, supportedMethods)
	}

	return nil
}

// GiteaConfig represents Gitea's app.ini configuration
type GiteaConfig struct {
	Database   DatabaseConfig   `ini:"database"`
	Repository RepositoryConfig `ini:"repository"`
	Picture    PictureConfig    `ini:"picture"`
}

type DatabaseConfig struct {
	DBType string `ini:"DB_TYPE"`
	Host   string `ini:"HOST"`
	Name   string `ini:"NAME"`
	User   string `ini:"USER"`
	Passwd string `ini:"PASSWD"`
	Path   string `ini:"PATH"` // For SQLite
}

type RepositoryConfig struct {
	Root string `ini:"ROOT"`
}

type PictureConfig struct {
	AvatarUploadPath           string `ini:"AVATAR_UPLOAD_PATH"`
	RepositoryAvatarUploadPath string `ini:"REPOSITORY_AVATAR_UPLOAD_PATH"`
}