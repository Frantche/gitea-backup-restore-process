from gitea_backup_restore.config import Settings
from  gitea_backup_restore.action import restore_history, zip_, files, database, gitea_config, remote_file_storage
from gitea_backup_restore.log import mylogger


def main():
    mylogger.info('Start backup process')
    
    config = Settings()
    gitea_conf = gitea_config.read_app_init(config)
    files.clean_tmp(config)

    mylogger.debug(config)

    database.backup(config,gitea_conf)
    files.backup(config,gitea_conf)
    zip_.compact(config)
    remote_file_storage.upload(config)
    remote_file_storage.ensure_max_retention(config)
    restore_history.increment(config, config.backup_tmp_remote_filename)
    
    mylogger.info('End backup process')
       