package history

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// Increment adds a backup filename to the history log
func Increment(settings *config.Settings, backupFilename string) error {
	logger.Debugf("Adding backup filename to history: %s", backupFilename)
	
	file, err := os.OpenFile(settings.BackupFileLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open backup log file: %w", err)
	}
	defer file.Close()
	
	_, err = file.WriteString(backupFilename + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to backup log file: %w", err)
	}
	
	logger.Debug("Backup filename added to history successfully")
	return nil
}

// Check verifies if a backup filename already exists in the history log
func Check(settings *config.Settings) (bool, error) {
	if settings.BackupFilename == "" {
		return false, nil
	}
	
	// Check if log file exists
	if _, err := os.Stat(settings.BackupFileLog); os.IsNotExist(err) {
		return false, nil
	}
	
	file, err := os.Open(settings.BackupFileLog)
	if err != nil {
		return false, fmt.Errorf("failed to open backup log file: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == settings.BackupFilename {
			logger.Infof("Backup filename found in history: %s", settings.BackupFilename)
			return true, nil
		}
	}
	
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading backup log file: %w", err)
	}
	
	return false, nil
}