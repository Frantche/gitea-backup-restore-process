from gitea_backup_restore.action.remote_file_storage import dict_provider
from pydantic import BaseSettings, validator, root_validator
from datetime import datetime
from typing import Optional

from gitea_backup_restore.log import mylogger


class Settings(BaseSettings):
    backup_enable: bool
    backup_methode: str
    backup_filename:Optional[str]
    backup_file_log: str = '/data/backupFileLog.txt'
    backup_tmp_remote_filename: str = f'@prefix-@date.zip'
    backup_prefix: str = 'gitea-backup'
    backup_max_retention: int = 5
    backup_tmp_folder: str = '/tmp/backup'
    backup_tmp_filename: str = '/tmp/backup.zip'
    restore_tmp_folder: str = '/tmp/restore'
    restore_tmp_filename: str = '/tmp/restore.zip'
    app_ini_path: str = '/data/gitea/conf/app.ini'

    @validator('backup_methode')
    def check_backup_methode(cls, v):
        
        if v in dict_provider.keys():
            dict_provider[v]['config']()
        else:
            raise ValueError(f'{v} is not a valid backup methode, please consider using: {list(dict_provider.keys())}')
        return v
    
    @root_validator
    def check_backup_tmp_remote_filename(cls, v):
        
        backup_tmp_remote_filename = v.get('backup_tmp_remote_filename')
        backup_prefix = v.get('backup_prefix')

        backup_tmp_remote_filename = backup_tmp_remote_filename.replace('@prefix', backup_prefix)
        backup_tmp_remote_filename = backup_tmp_remote_filename.replace('@date', f'{datetime.now():%Y-%m-%d-%H-%M-%S}')

        v['backup_tmp_remote_filename'] = backup_tmp_remote_filename
        

        return v