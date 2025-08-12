package database

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// PostgreSQLAdapter implements DatabaseAdapter for PostgreSQL
type PostgreSQLAdapter struct{}

func (p *PostgreSQLAdapter) Backup(settings *config.Settings, giteaConfig *config.GiteaConfig) error {
	host, port := parseHostPort(giteaConfig.Database.Host)
	
	// Set password environment variable
	os.Setenv("PGPASSWORD", giteaConfig.Database.Passwd)
	defer os.Unsetenv("PGPASSWORD")
	
	outputFile := filepath.Join(settings.BackupTmpFolder, "dump.postgres.sql")
	
	args := []string{
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--username=%s", giteaConfig.Database.User),
	}
	
	if port != "" {
		args = append(args, fmt.Sprintf("--port=%s", port))
	}
	
	args = append(args, giteaConfig.Database.Name)
	
	cmd := exec.Command("pg_dump", args...)
	
	// Redirect output to file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()
	
	cmd.Stdout = outFile
	
	logger.Debugf("Running PostgreSQL dump command: pg_dump %s", strings.Join(args, " "))
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failed: %w", err)
	}
	
	logger.Info("PostgreSQL database backup completed")
	return nil
}

func (p *PostgreSQLAdapter) Restore(settings *config.Settings, giteaConfig *config.GiteaConfig) error {
	host, port := parseHostPort(giteaConfig.Database.Host)
	
	// Set password environment variable
	os.Setenv("PGPASSWORD", giteaConfig.Database.Passwd)
	defer os.Unsetenv("PGPASSWORD")
	
	inputFile := filepath.Join(settings.RestoreTmpFolder, "dump.postgres.sql")
	
	// First, drop existing tables (equivalent to "drop owned by user")
	dropArgs := []string{
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--username=%s", giteaConfig.Database.User),
	}
	
	if port != "" {
		dropArgs = append(dropArgs, fmt.Sprintf("--port=%s", port))
	}
	
	dropArgs = append(dropArgs, 
		"-c", 
		fmt.Sprintf("DROP OWNED BY %s", giteaConfig.Database.User),
		giteaConfig.Database.Name,
	)
	
	dropCmd := exec.Command("psql", dropArgs...)
	
	logger.Debugf("Running PostgreSQL drop command: psql %s", strings.Join(dropArgs, " "))
	
	// Drop command might fail if there are no existing tables, so we don't check error
	dropCmd.Run()
	
	// Now restore from backup
	restoreArgs := []string{
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--username=%s", giteaConfig.Database.User),
	}
	
	if port != "" {
		restoreArgs = append(restoreArgs, fmt.Sprintf("--port=%s", port))
	}
	
	restoreArgs = append(restoreArgs, giteaConfig.Database.Name)
	
	restoreCmd := exec.Command("psql", restoreArgs...)
	
	// Read input from file
	inFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inFile.Close()
	
	restoreCmd.Stdin = inFile
	
	logger.Debugf("Running PostgreSQL restore command: psql %s", strings.Join(restoreArgs, " "))
	
	if err := restoreCmd.Run(); err != nil {
		return fmt.Errorf("psql restore failed: %w", err)
	}
	
	logger.Info("PostgreSQL database restore completed")
	return nil
}