package config_test

import (
	"os"
	"testing"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
)

func TestNewSettings_WithDefaults(t *testing.T) {
	// Clear environment variables
	clearEnvVars()
	
	// Set required environment variables
	os.Setenv("BACKUP_ENABLE", "true")
	os.Setenv("BACKUP_METHODE", "s3")
	defer clearEnvVars()
	
	settings, err := config.NewSettings()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if settings.BackupEnable != true {
		t.Errorf("Expected BackupEnable to be true, got %v", settings.BackupEnable)
	}
	
	if settings.BackupMethod != "s3" {
		t.Errorf("Expected BackupMethod to be 's3', got %v", settings.BackupMethod)
	}
	
	if settings.BackupPrefix != "gitea-backup" {
		t.Errorf("Expected BackupPrefix to be 'gitea-backup', got %v", settings.BackupPrefix)
	}
	
	if settings.BackupMaxRetention != 5 {
		t.Errorf("Expected BackupMaxRetention to be 5, got %v", settings.BackupMaxRetention)
	}
}

func TestNewSettings_WithEnvironmentOverrides(t *testing.T) {
	// Clear environment variables
	clearEnvVars()
	
	// Set custom environment variables
	os.Setenv("BACKUP_ENABLE", "false")
	os.Setenv("BACKUP_METHODE", "ftp")
	os.Setenv("BACKUP_PREFIX", "custom-prefix")
	os.Setenv("BACKUP_MAX_RETENTION", "10")
	defer clearEnvVars()
	
	settings, err := config.NewSettings()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if settings.BackupEnable != false {
		t.Errorf("Expected BackupEnable to be false, got %v", settings.BackupEnable)
	}
	
	if settings.BackupMethod != "ftp" {
		t.Errorf("Expected BackupMethod to be 'ftp', got %v", settings.BackupMethod)
	}
	
	if settings.BackupPrefix != "custom-prefix" {
		t.Errorf("Expected BackupPrefix to be 'custom-prefix', got %v", settings.BackupPrefix)
	}
	
	if settings.BackupMaxRetention != 10 {
		t.Errorf("Expected BackupMaxRetention to be 10, got %v", settings.BackupMaxRetention)
	}
}

func TestNewSettings_InvalidBackupMethod(t *testing.T) {
	// Clear environment variables
	clearEnvVars()
	
	// Set invalid backup method
	os.Setenv("BACKUP_ENABLE", "true")
	os.Setenv("BACKUP_METHODE", "invalid")
	defer clearEnvVars()
	
	_, err := config.NewSettings()
	if err == nil {
		t.Fatalf("Expected error for invalid backup method, got nil")
	}
}

func TestNewSettings_TemplateReplacement(t *testing.T) {
	// Clear environment variables
	clearEnvVars()
	
	os.Setenv("BACKUP_ENABLE", "true")
	os.Setenv("BACKUP_METHODE", "s3")
	os.Setenv("BACKUP_PREFIX", "test-backup")
	os.Setenv("BACKUP_TMP_REMOTE_FILENAME", "@prefix-@date.zip")
	defer clearEnvVars()
	
	settings, err := config.NewSettings()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Check that template replacement occurred
	if settings.BackupTmpRemoteFilename == "@prefix-@date.zip" {
		t.Error("Template replacement did not occur")
	}
	
	// Check that prefix was replaced
	if !contains(settings.BackupTmpRemoteFilename, "test-backup") {
		t.Errorf("Expected filename to contain 'test-backup', got %v", settings.BackupTmpRemoteFilename)
	}
	
	// Check that date was replaced (should contain hyphens and numbers)
	if contains(settings.BackupTmpRemoteFilename, "@date") {
		t.Errorf("Expected date template to be replaced, got %v", settings.BackupTmpRemoteFilename)
	}
}

func clearEnvVars() {
	envVars := []string{
		"BACKUP_ENABLE",
		"BACKUP_METHODE",
		"BACKUP_METHOD",
		"BACKUP_FILENAME",
		"BACKUP_PREFIX",
		"BACKUP_MAX_RETENTION",
		"BACKUP_TMP_REMOTE_FILENAME",
	}
	
	for _, env := range envVars {
		os.Unsetenv(env)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr && 
		   (len(s) >= len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}