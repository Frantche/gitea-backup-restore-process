package database

import (
	"fmt"
	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// DatabaseAdapter defines the interface for database backup and restore operations
type DatabaseAdapter interface {
	Backup(settings *config.Settings, giteaConfig *config.GiteaConfig) error
	Restore(settings *config.Settings, giteaConfig *config.GiteaConfig) error
}

// GetAdapter returns the appropriate database adapter based on the database type
func GetAdapter(dbType string) (DatabaseAdapter, error) {
	switch dbType {
	case "mysql":
		return &MySQLAdapter{}, nil
	case "postgres":
		return &PostgreSQLAdapter{}, nil
	case "sqlite3":
		return &SQLiteAdapter{}, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// BackupDatabase performs database backup using the appropriate adapter
func BackupDatabase(settings *config.Settings, giteaConfig *config.GiteaConfig) error {
	logger.Infof("Starting database backup for %s", giteaConfig.Database.DBType)
	
	adapter, err := GetAdapter(giteaConfig.Database.DBType)
	if err != nil {
		return err
	}
	
	if err := adapter.Backup(settings, giteaConfig); err != nil {
		return fmt.Errorf("database backup failed: %w", err)
	}
	
	logger.Info("Database backup completed successfully")
	return nil
}

// RestoreDatabase performs database restore using the appropriate adapter
func RestoreDatabase(settings *config.Settings, giteaConfig *config.GiteaConfig) error {
	logger.Infof("Starting database restore for %s", giteaConfig.Database.DBType)
	
	adapter, err := GetAdapter(giteaConfig.Database.DBType)
	if err != nil {
		return err
	}
	
	if err := adapter.Restore(settings, giteaConfig); err != nil {
		return fmt.Errorf("database restore failed: %w", err)
	}
	
	logger.Info("Database restore completed successfully")
	return nil
}