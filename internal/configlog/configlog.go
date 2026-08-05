package configlog

import (
	"github.com/Frantche/gitea-backup-restore-process/internal/config"
	"github.com/Frantche/gitea-backup-restore-process/pkg/logger"
)

// Debug writes an explicit allowlist of operational configuration fields.
// Credentials and complete configuration structures must never be logged.
func Debug(settings *config.Settings, giteaConfig *config.GiteaConfig) {
	logger.Debugf(
		"Settings: backup_method=%q backup_prefix=%q backup_max_retention=%d backup_tmp_folder=%q restore_tmp_folder=%q app_ini_path=%q",
		settings.BackupMethod,
		settings.BackupPrefix,
		settings.BackupMaxRetention,
		settings.BackupTmpFolder,
		settings.RestoreTmpFolder,
		settings.AppIniPath,
	)
	logger.Debugf(
		"Gitea config: database_type=%q database_host=%q database_name=%q database_user=%q repository_root=%q avatar_upload_path=%q repository_avatar_upload_path=%q",
		giteaConfig.Database.DBType,
		giteaConfig.Database.Host,
		giteaConfig.Database.Name,
		giteaConfig.Database.User,
		giteaConfig.Repository.Root,
		giteaConfig.Picture.AvatarUploadPath,
		giteaConfig.Picture.RepositoryAvatarUploadPath,
	)
}
