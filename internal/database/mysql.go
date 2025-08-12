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

// MySQLAdapter implements DatabaseAdapter for MySQL
type MySQLAdapter struct{}

func (m *MySQLAdapter) Backup(settings *config.Settings, giteaConfig *config.GiteaConfig) error {
	host, port := parseHostPort(giteaConfig.Database.Host)
	
	// Set password environment variable
	os.Setenv("MYSQL_PWD", giteaConfig.Database.Passwd)
	defer os.Unsetenv("MYSQL_PWD")
	
	outputFile := filepath.Join(settings.BackupTmpFolder, "dump.mysql.sql")
	
	args := []string{
		"--column-statistics=0",
		"--no-tablespaces",
		fmt.Sprintf("--host=%s", host),
	}
	
	if port != "" {
		args = append(args, fmt.Sprintf("--port=%s", port))
	}
	
	args = append(args,
		fmt.Sprintf("--user=%s", giteaConfig.Database.User),
		giteaConfig.Database.Name,
	)
	
	cmd := exec.Command("mysqldump", args...)
	
	// Redirect output to file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()
	
	cmd.Stdout = outFile
	
	logger.Debugf("Running MySQL dump command: mysqldump %s", strings.Join(args, " "))
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysqldump failed: %w", err)
	}
	
	logger.Info("MySQL database backup completed")
	return nil
}

func (m *MySQLAdapter) Restore(settings *config.Settings, giteaConfig *config.GiteaConfig) error {
	host, port := parseHostPort(giteaConfig.Database.Host)
	
	// Set password environment variable
	os.Setenv("MYSQL_PWD", giteaConfig.Database.Passwd)
	defer os.Unsetenv("MYSQL_PWD")
	
	inputFile := filepath.Join(settings.RestoreTmpFolder, "dump.mysql.sql")
	
	args := []string{
		fmt.Sprintf("--host=%s", host),
	}
	
	if port != "" {
		args = append(args, fmt.Sprintf("--port=%s", port))
	}
	
	args = append(args,
		fmt.Sprintf("--user=%s", giteaConfig.Database.User),
		giteaConfig.Database.Name,
	)
	
	cmd := exec.Command("mysql", args...)
	
	// Read input from file
	inFile, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inFile.Close()
	
	cmd.Stdin = inFile
	
	logger.Debugf("Running MySQL restore command: mysql %s", strings.Join(args, " "))
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysql restore failed: %w", err)
	}
	
	logger.Info("MySQL database restore completed")
	return nil
}

// parseHostPort splits host:port into separate components
func parseHostPort(hostPort string) (host, port string) {
	parts := strings.Split(hostPort, ":")
	host = parts[0]
	if len(parts) > 1 {
		port = parts[1]
	}
	return
}