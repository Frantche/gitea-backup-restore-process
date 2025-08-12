package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
)

func TestReadGiteaConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "app.ini")
	
	configContent := `[database]
DB_TYPE = mysql
HOST = localhost:3306
NAME = gitea
USER = root
PASSWD = password

[repository]
ROOT = /data/git/repositories

[picture]
AVATAR_UPLOAD_PATH = /data/gitea/avatars
REPOSITORY_AVATAR_UPLOAD_PATH = /data/gitea/repo-avatars
`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Test reading the config
	giteaConfig, err := config.ReadGiteaConfig(configFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify database config
	if giteaConfig.Database.DBType != "mysql" {
		t.Errorf("Expected DB_TYPE to be 'mysql', got %v", giteaConfig.Database.DBType)
	}
	
	if giteaConfig.Database.Host != "localhost:3306" {
		t.Errorf("Expected HOST to be 'localhost:3306', got %v", giteaConfig.Database.Host)
	}
	
	if giteaConfig.Database.Name != "gitea" {
		t.Errorf("Expected NAME to be 'gitea', got %v", giteaConfig.Database.Name)
	}
	
	if giteaConfig.Database.User != "root" {
		t.Errorf("Expected USER to be 'root', got %v", giteaConfig.Database.User)
	}
	
	if giteaConfig.Database.Passwd != "password" {
		t.Errorf("Expected PASSWD to be 'password', got %v", giteaConfig.Database.Passwd)
	}
	
	// Verify repository config
	if giteaConfig.Repository.Root != "/data/git/repositories" {
		t.Errorf("Expected ROOT to be '/data/git/repositories', got %v", giteaConfig.Repository.Root)
	}
	
	// Verify picture config
	if giteaConfig.Picture.AvatarUploadPath != "/data/gitea/avatars" {
		t.Errorf("Expected AVATAR_UPLOAD_PATH to be '/data/gitea/avatars', got %v", giteaConfig.Picture.AvatarUploadPath)
	}
	
	if giteaConfig.Picture.RepositoryAvatarUploadPath != "/data/gitea/repo-avatars" {
		t.Errorf("Expected REPOSITORY_AVATAR_UPLOAD_PATH to be '/data/gitea/repo-avatars', got %v", giteaConfig.Picture.RepositoryAvatarUploadPath)
	}
}

func TestReadGiteaConfig_SQLite(t *testing.T) {
	// Create a temporary config file for SQLite
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "app.ini")
	
	configContent := `[database]
DB_TYPE = sqlite3
PATH = /data/gitea/gitea.db

[repository]
ROOT = /data/git/repositories
`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Test reading the config
	giteaConfig, err := config.ReadGiteaConfig(configFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify database config
	if giteaConfig.Database.DBType != "sqlite3" {
		t.Errorf("Expected DB_TYPE to be 'sqlite3', got %v", giteaConfig.Database.DBType)
	}
	
	if giteaConfig.Database.Path != "/data/gitea/gitea.db" {
		t.Errorf("Expected PATH to be '/data/gitea/gitea.db', got %v", giteaConfig.Database.Path)
	}
}

func TestReadGiteaConfig_WithComments(t *testing.T) {
	// Create a temporary config file with comments
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "app.ini")
	
	configContent := `; This is a comment
[database]
; Another comment
DB_TYPE = postgres
HOST = postgres:5432
# Hash comment
NAME = gitea
USER = gitea
PASSWD = "quoted password"

; Empty line above
[repository]
ROOT = /data/git/repositories
`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Test reading the config
	giteaConfig, err := config.ReadGiteaConfig(configFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify database config
	if giteaConfig.Database.DBType != "postgres" {
		t.Errorf("Expected DB_TYPE to be 'postgres', got %v", giteaConfig.Database.DBType)
	}
	
	if giteaConfig.Database.Passwd != "quoted password" {
		t.Errorf("Expected PASSWD to be 'quoted password', got %v", giteaConfig.Database.Passwd)
	}
}

func TestReadGiteaConfig_NonExistentFile(t *testing.T) {
	_, err := config.ReadGiteaConfig("/non/existent/file.ini")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
}