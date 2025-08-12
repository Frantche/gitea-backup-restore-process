from gitea_backup_restore.config import Settings
from  gitea_backup_restore.action import restore_history, zip_, files, database, gitea_config, remote_file_storage
from gitea_backup_restore.log import mylogger


def main():
    
    config = Settings()
    gitea_conf = gitea_config.read_app_init(config)

    if not restore_history.check(config):
        mylogger.info('Start restore process')

        remote_file_storage.download(config)
        zip_.extract(config)
        files.restore(config, gitea_conf)
        database.restore(config, gitea_conf)

        restore_history.increment(config, config.backup_filename)
        mylogger.info('End restore process')
    else:
        mylogger.info('Gitea has been already restore with this file version')
    