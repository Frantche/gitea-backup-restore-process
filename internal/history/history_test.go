package history_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/internal/history"
)

func TestIncrement(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "backup.log")
	
	settings := &config.Settings{
		BackupFileLog: logFile,
	}
	
	// Test adding first entry
	err := history.Increment(settings, "backup-1.zip")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify file was created and contains the entry
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	expected := "backup-1.zip\n"
	if string(content) != expected {
		t.Errorf("Expected log content '%s', got '%s'", expected, string(content))
	}
	
	// Test adding second entry
	err = history.Increment(settings, "backup-2.zip")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify both entries are in the file
	content, err = os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	expected = "backup-1.zip\nbackup-2.zip\n"
	if string(content) != expected {
		t.Errorf("Expected log content '%s', got '%s'", expected, string(content))
	}
}

func TestCheck_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "nonexistent.log")
	
	settings := &config.Settings{
		BackupFileLog:  logFile,
		BackupFilename: "backup-1.zip",
	}
	
	found, err := history.Check(settings)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if found {
		t.Error("Expected false when log file doesn't exist, got true")
	}
}

func TestCheck_EmptyBackupFilename(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "backup.log")
	
	settings := &config.Settings{
		BackupFileLog:  logFile,
		BackupFilename: "",
	}
	
	found, err := history.Check(settings)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if found {
		t.Error("Expected false when backup filename is empty, got true")
	}
}

func TestCheck_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "backup.log")
	
	// Create log file with some entries
	logContent := "backup-1.zip\nbackup-2.zip\nbackup-3.zip\n"
	err := os.WriteFile(logFile, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}
	
	settings := &config.Settings{
		BackupFileLog:  logFile,
		BackupFilename: "backup-2.zip",
	}
	
	// Test finding existing entry
	found, err := history.Check(settings)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !found {
		t.Error("Expected true when backup filename exists, got false")
	}
	
	// Test not finding non-existent entry
	settings.BackupFilename = "backup-4.zip"
	found, err = history.Check(settings)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if found {
		t.Error("Expected false when backup filename doesn't exist, got true")
	}
}

func TestCheck_WithWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "backup.log")
	
	// Create log file with entries containing whitespace
	logContent := "backup-1.zip\n  backup-2.zip  \nbackup-3.zip\n"
	err := os.WriteFile(logFile, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}
	
	settings := &config.Settings{
		BackupFileLog:  logFile,
		BackupFilename: "backup-2.zip",
	}
	
	// Should find the entry even with surrounding whitespace
	found, err := history.Check(settings)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !found {
		t.Error("Expected true when backup filename exists (with whitespace), got false")
	}
}