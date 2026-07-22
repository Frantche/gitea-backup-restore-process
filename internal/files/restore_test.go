package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceDirRemovesStaleFiles(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "src")
	dst := filepath.Join(root, "dst")

	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatalf("create source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "restored"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatalf("create destination: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dst, "stale"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write stale destination: %v", err)
	}

	if err := replaceDir(src, dst); err != nil {
		t.Fatalf("replaceDir failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "restored")); err != nil {
		t.Fatalf("restored file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "stale")); !os.IsNotExist(err) {
		t.Fatalf("stale file should be removed, got err: %v", err)
	}
}

func TestRestoreOwnerSpec(t *testing.T) {
	if got := restoreOwnerSpec("git"); got != "git:git" {
		t.Fatalf("restoreOwnerSpec(git) = %q, want git:git", got)
	}
	if got := restoreOwnerSpec("1000:1000"); got != "1000:1000" {
		t.Fatalf("restoreOwnerSpec(1000:1000) = %q, want 1000:1000", got)
	}
}
