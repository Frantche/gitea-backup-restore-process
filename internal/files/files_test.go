package files

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
)

func TestCleanTmpRemovesNestedGitDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backup")
	restoreDir := filepath.Join(tmpDir, "restore")
	backupTmpFile := filepath.Join(tmpDir, "backup.zip")
	restoreTmpFile := filepath.Join(tmpDir, "restore.zip")

	nestedGitObjects := filepath.Join(backupDir, "repo", "francois", "seedbox.git", "objects", "pack")
	if err := os.MkdirAll(nestedGitObjects, 0755); err != nil {
		t.Fatalf("failed to create nested backup dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedGitObjects, "pack-file"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to create nested backup file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(restoreDir, "repo"), 0755); err != nil {
		t.Fatalf("failed to create restore dir: %v", err)
	}
	for _, path := range []string{backupTmpFile, restoreTmpFile} {
		if err := os.WriteFile(path, []byte("archive"), 0644); err != nil {
			t.Fatalf("failed to create tmp file %s: %v", path, err)
		}
	}

	settings := &config.Settings{
		BackupTmpFolder:    backupDir,
		BackupTmpFilename:  backupTmpFile,
		RestoreTmpFolder:   restoreDir,
		RestoreTmpFilename: restoreTmpFile,
	}

	if err := CleanTmp(settings); err != nil {
		t.Fatalf("CleanTmp failed: %v", err)
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("backup tmp folder should exist after cleanup: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("backup tmp folder should be empty after cleanup, got %d entries", len(entries))
	}
	if _, err := os.Stat(restoreDir); !os.IsNotExist(err) {
		t.Fatalf("restore tmp folder should be removed, got err: %v", err)
	}
	for _, path := range []string{backupTmpFile, restoreTmpFile} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("tmp file %s should be removed, got err: %v", path, err)
		}
	}
}

func TestRemoveAllWithRetryRetriesTransientErrors(t *testing.T) {
	attempts := 0

	err := removeAllWithRetry("/tmp/backup", func(path string) error {
		attempts++
		if attempts < 3 {
			return &os.PathError{Op: "unlinkat", Path: filepath.Join(path, "objects"), Err: syscall.ENOTEMPTY}
		}
		return nil
	}, 5, 0)
	if err != nil {
		t.Fatalf("expected retry to succeed, got: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRemoveAllWithRetryReturnsLastError(t *testing.T) {
	expectedErr := errors.New("remove failed")
	attempts := 0

	err := removeAllWithRetry("/tmp/backup", func(path string) error {
		attempts++
		return expectedErr
	}, 3, 0)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected final error %v, got %v", expectedErr, err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
