package database

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
		"--no-owner",
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
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	logger.Debugf("Running PostgreSQL dump command: pg_dump %s", strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failed: %w: %s", err, strings.TrimSpace(stderr.String()))
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

	owner := quotePostgresIdentifier(giteaConfig.Database.User)

	// Reset the public schema instead of only dropping objects owned by the
	// current user. Older dumps can contain objects owned by a different role.
	dropArgs := []string{
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--username=%s", giteaConfig.Database.User),
		"-v",
		"ON_ERROR_STOP=1",
	}

	if port != "" {
		dropArgs = append(dropArgs, fmt.Sprintf("--port=%s", port))
	}

	dropArgs = append(dropArgs,
		"-c",
		fmt.Sprintf("DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public AUTHORIZATION %s; GRANT ALL ON SCHEMA public TO %s; GRANT ALL ON SCHEMA public TO public;", owner, owner),
		giteaConfig.Database.Name,
	)

	dropCmd := exec.Command("psql", dropArgs...)

	logger.Debugf("Running PostgreSQL drop command: psql %s", strings.Join(dropArgs, " "))

	if output, err := dropCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("psql schema reset failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	// Now restore from backup
	restoreArgs := []string{
		fmt.Sprintf("--host=%s", host),
		fmt.Sprintf("--username=%s", giteaConfig.Database.User),
		"-v",
		"ON_ERROR_STOP=1",
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

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		pipeWriter.CloseWithError(rewritePostgresDumpOwners(inFile, pipeWriter, giteaConfig.Database.User))
	}()
	restoreCmd.Stdin = pipeReader

	logger.Debugf("Running PostgreSQL restore command: psql %s", strings.Join(restoreArgs, " "))

	if output, err := restoreCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("psql restore failed: %w: %s", err, strings.TrimSpace(string(output)))
	}

	logger.Info("PostgreSQL database restore completed")
	return nil
}

var postgresOwnerPattern = regexp.MustCompile(`OWNER TO ([^;\s]+);`)

func rewritePostgresDumpOwners(input io.Reader, output io.Writer, owner string) error {
	replacement := "OWNER TO " + quotePostgresIdentifier(owner) + ";"
	reader := bufio.NewReader(input)
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			if _, writeErr := io.WriteString(output, postgresOwnerPattern.ReplaceAllString(line, replacement)); writeErr != nil {
				return writeErr
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func quotePostgresIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
