package configlog

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

func TestDebugNeverLogsDatabasePassword(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	var output bytes.Buffer
	previous := logger.DebugLogger
	logger.DebugLogger = log.New(&output, "", 0)
	t.Cleanup(func() { logger.DebugLogger = previous })

	settings := &config.Settings{
		BackupMethod:       "s3",
		BackupPrefix:       "gitea-backup",
		BackupMaxRetention: 7,
		BackupTmpFolder:    "/tmp/backup",
		RestoreTmpFolder:   "/tmp/restore",
		AppIniPath:         "/data/gitea/conf/app.ini",
	}
	giteaConfig := &config.GiteaConfig{
		Database: config.DatabaseConfig{
			DBType: "postgres",
			Host:   "gitea-db:5432",
			Name:   "gitea",
			User:   "gitea",
			Passwd: "sentinel-database-secret",
		},
		Repository: config.RepositoryConfig{Root: "/data/git/repositories"},
	}

	Debug(settings, giteaConfig)
	got := output.String()
	if strings.Contains(got, "sentinel-database-secret") {
		t.Fatalf("debug output contains database password: %q", got)
	}
	for _, expected := range []string{"backup_method=\"s3\"", "database_type=\"postgres\"", "database_user=\"gitea\""} {
		if !strings.Contains(got, expected) {
			t.Fatalf("debug output %q does not contain %q", got, expected)
		}
	}
}
