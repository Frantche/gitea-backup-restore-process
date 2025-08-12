package database_test

import (
	"testing"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/internal/database"
)

func TestGetAdapter(t *testing.T) {
	tests := []struct {
		dbType      string
		expectError bool
	}{
		{"mysql", false},
		{"postgres", false},
		{"sqlite3", false},
		{"unsupported", true},
		{"", true},
	}
	
	for _, test := range tests {
		t.Run(test.dbType, func(t *testing.T) {
			adapter, err := database.GetAdapter(test.dbType)
			
			if test.expectError {
				if err == nil {
					t.Errorf("Expected error for dbType '%s', got nil", test.dbType)
				}
				if adapter != nil {
					t.Errorf("Expected nil adapter for invalid dbType '%s', got %T", test.dbType, adapter)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for dbType '%s', got %v", test.dbType, err)
				}
				if adapter == nil {
					t.Errorf("Expected valid adapter for dbType '%s', got nil", test.dbType)
				}
			}
		})
	}
}

func TestSQLiteAdapter_BackupRestore(t *testing.T) {
	// This test requires actual file operations, so we'll test the logic without executing
	adapter, err := database.GetAdapter("sqlite3")
	if err != nil {
		t.Fatalf("Failed to get SQLite adapter: %v", err)
	}
	
	if adapter == nil {
		t.Fatal("SQLite adapter should not be nil")
	}
	
	// Test with invalid config (missing database path)
	settings := &config.Settings{
		BackupTmpFolder:  "/tmp/backup",
		RestoreTmpFolder: "/tmp/restore",
	}
	
	giteaConfig := &config.GiteaConfig{
		Database: config.DatabaseConfig{
			DBType: "sqlite3",
			Path:   "", // Empty path should cause error
		},
	}
	
	// Backup should fail with empty path
	err = adapter.Backup(settings, giteaConfig)
	if err == nil {
		t.Error("Expected error for empty SQLite database path, got nil")
	}
	
	// Restore should fail with empty path
	err = adapter.Restore(settings, giteaConfig)
	if err == nil {
		t.Error("Expected error for empty SQLite database path, got nil")
	}
}

func TestMySQLAdapter_Creation(t *testing.T) {
	adapter, err := database.GetAdapter("mysql")
	if err != nil {
		t.Fatalf("Failed to get MySQL adapter: %v", err)
	}
	
	if adapter == nil {
		t.Fatal("MySQL adapter should not be nil")
	}
}

func TestPostgreSQLAdapter_Creation(t *testing.T) {
	adapter, err := database.GetAdapter("postgres")
	if err != nil {
		t.Fatalf("Failed to get PostgreSQL adapter: %v", err)
	}
	
	if adapter == nil {
		t.Fatal("PostgreSQL adapter should not be nil")
	}
}