import shutil
import os
from gitea_backup_restore.log import mylogger

import subprocess
from pathlib import Path

def clean_tmp(settings):

    dirpath = Path(settings.backup_tmp_folder)
    if dirpath.exists() and dirpath.is_dir():
        shutil.rmtree(dirpath)
    
    os.makedirs(settings.backup_tmp_folder)

    if os.path.exists(settings.backup_tmp_filename):
        os.remove(settings.backup_tmp_filename)
    
    dirpath = Path(settings.restore_tmp_folder)
    if dirpath.exists() and dirpath.is_dir():
        shutil.rmtree(dirpath)

    if os.path.exists(settings.restore_tmp_filename):
        os.remove(settings.restore_tmp_filename)

def move_folder(source, target, settings):

    if os.path.exists(source):

        shutil.copytree(source, target, dirs_exist_ok=True)
    else:
        print(f'source do not exist : {source}')

def restore(settings, gitea_config):

    move_folder(f"{settings.restore_tmp_folder}/repo/",gitea_config['repository']['ROOT'], settings)
    move_folder(f"{settings.restore_tmp_folder}/avatars/",gitea_config['picture']['AVATAR_UPLOAD_PATH'], settings)
    move_folder(f"{settings.restore_tmp_folder}/repo-avatars/",gitea_config['picture']['REPOSITORY_AVATAR_UPLOAD_PATH'], settings)

    mylogger.info('Files are restore')

def backup(settings, gitea_config):

    move_folder(gitea_config['repository']['ROOT'], f"{settings.backup_tmp_folder}/repo/", settings)
    move_folder(gitea_config['picture']['AVATAR_UPLOAD_PATH'], f"{settings.backup_tmp_folder}/avatars/", settings)
    move_folder(gitea_config['picture']['REPOSITORY_AVATAR_UPLOAD_PATH'], f"{settings.backup_tmp_folder}/repo-avatars/", settings)

    mylogger.info('Files are backup')
