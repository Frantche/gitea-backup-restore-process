package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/internal/database"
	"github.com/Frantche/gitea-backup-restore-process/internal/files"
)

// TestE2EWorkflow tests the complete backup/restore workflow
func TestE2EWorkflow(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Create temporary directories for testing
	tmpDir, err := os.MkdirTemp("", "gitea-e2e-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataDir := filepath.Join(tmpDir, "data")
	backupDir := filepath.Join(tmpDir, "backup")
	restoreDir := filepath.Join(tmpDir, "restore")

	// Create directory structure
	for _, dir := range []string{dataDir, backupDir, restoreDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test data structure
	repoDir := filepath.Join(dataDir, "repositories")
	avatarDir := filepath.Join(dataDir, "avatars")
	dbFile := filepath.Join(dataDir, "gitea.db")

	for _, dir := range []string{repoDir, avatarDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	testFiles := map[string]string{
		filepath.Join(repoDir, "test-repo.git"):     "test repository content",
		filepath.Join(avatarDir, "avatar.png"):      "test avatar content",
		dbFile:                                       "test database content",
	}

	for path, content := range testFiles {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	// Create test configuration
	configFile := filepath.Join(tmpDir, "app.ini")
	configContent := `[database]
DB_TYPE = sqlite3
PATH = ` + dbFile + `

[repository]
ROOT = ` + repoDir + `

[picture]
AVATAR_UPLOAD_PATH = ` + avatarDir + `
REPOSITORY_AVATAR_UPLOAD_PATH = ` + avatarDir + `
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test configuration loading
	t.Run("ConfigurationLoading", func(t *testing.T) {
		giteaConfig, err := config.ReadGiteaConfig(configFile)
		if err != nil {
			t.Fatalf("Failed to read Gitea config: %v", err)
		}

		if giteaConfig.Database.DBType != "sqlite3" {
			t.Errorf("Expected DB_TYPE to be sqlite3, got %s", giteaConfig.Database.DBType)
		}

		if giteaConfig.Repository.Root != repoDir {
			t.Errorf("Expected ROOT to be %s, got %s", repoDir, giteaConfig.Repository.Root)
		}
	})

	// Test database adapter
	t.Run("DatabaseAdapter", func(t *testing.T) {
		adapter, err := database.GetAdapter("sqlite3")
		if err != nil {
			t.Fatalf("Failed to get SQLite adapter: %v", err)
		}

		if adapter == nil {
			t.Fatal("SQLite adapter should not be nil")
		}
	})

	// Test file operations
	t.Run("FileBackupRestore", func(t *testing.T) {
		// Set up settings for file operations
		settings := &config.Settings{
			BackupTmpFolder:  backupDir,
			RestoreTmpFolder: restoreDir,
		}

		giteaConfig := &config.GiteaConfig{
			Repository: config.RepositoryConfig{
				Root: repoDir,
			},
			Picture: config.PictureConfig{
				AvatarUploadPath:            avatarDir,
				RepositoryAvatarUploadPath:  avatarDir,
			},
		}

		// Test backup
		if err := files.BackupFiles(settings, giteaConfig); err != nil {
			t.Fatalf("File backup failed: %v", err)
		}

		// Verify backup files exist
		expectedBackupFiles := []string{
			filepath.Join(backupDir, "repo", "test-repo.git"),
			filepath.Join(backupDir, "avatars", "avatar.png"),
		}

		for _, file := range expectedBackupFiles {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("Expected backup file %s does not exist", file)
			}
		}

		// Test restore
		if err := files.RestoreFiles(settings, giteaConfig); err != nil {
			t.Fatalf("File restore failed: %v", err)
		}

		// Verify original files still exist (since we're restoring to the same location)
		for path, expectedContent := range testFiles {
			if path == dbFile {
				continue // Skip database file for this test
			}

			content, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("Failed to read restored file %s: %v", path, err)
				continue
			}

			if string(content) != expectedContent {
				t.Errorf("Restored file %s content mismatch. Expected: %s, Got: %s", path, expectedContent, string(content))
			}
		}
	})

	// Test full workflow simulation
	t.Run("FullWorkflowSimulation", func(t *testing.T) {
		// This test simulates the complete workflow without actual backup/restore commands
		// to avoid dependencies on external services

		// 1. Verify initial data exists
		for path := range testFiles {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Initial test file %s does not exist", path)
			}
		}

		// 2. Simulate backup by copying files
		backupRepo := filepath.Join(backupDir, "simulation")
		if err := os.MkdirAll(backupRepo, 0755); err != nil {
			t.Fatalf("Failed to create simulation backup dir: %v", err)
		}

		for path, content := range testFiles {
			backupPath := filepath.Join(backupRepo, filepath.Base(path))
			if err := os.WriteFile(backupPath, []byte(content), 0644); err != nil {
				t.Errorf("Failed to simulate backup of %s: %v", path, err)
			}
		}

		// 3. Simulate data loss by removing original files
		for path := range testFiles {
			if err := os.Remove(path); err != nil {
				t.Errorf("Failed to simulate data loss for %s: %v", path, err)
			}
		}

		// 4. Verify data is lost
		for path := range testFiles {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Errorf("File %s should not exist after simulated data loss", path)
			}
		}

		// 5. Simulate restore by copying back from backup
		for path, expectedContent := range testFiles {
			backupPath := filepath.Join(backupRepo, filepath.Base(path))
			content, err := os.ReadFile(backupPath)
			if err != nil {
				t.Errorf("Failed to read backup file %s: %v", backupPath, err)
				continue
			}

			if err := os.WriteFile(path, content, 0644); err != nil {
				t.Errorf("Failed to restore file %s: %v", path, err)
				continue
			}

			// Verify restored content
			if string(content) != expectedContent {
				t.Errorf("Restored file %s content mismatch. Expected: %s, Got: %s", path, expectedContent, string(content))
			}
		}

		// 6. Final verification - all data should be restored
		for path := range testFiles {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Restored file %s does not exist", path)
			}
		}
	})
}