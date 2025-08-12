import shutil
import os
from gitea_backup_restore.log import mylogger

import subprocess

def backup(settings, gitea_config):

    db_full_connection = gitea_config['database']['HOST']
    db_host = db_full_connection.split(':')[0]
    db_port = db_full_connection.split(':')[1]
    db_username = gitea_config['database']['USER']
    db_password = gitea_config['database']['PASSWD']
    db_database = gitea_config['database']['NAME']
    
    os.environ["MYSQL_PWD"] = db_password

    command = f"mysqldump --column-statistics=0   --no-tablespaces --host={db_host} --user={db_username} {db_database} > {settings.backup_tmp_folder}/dump.mysql.sql"

    mylogger.debug(f'dump MYSQL db command line: {command}')

    subprocess.run(command, 
        stdout=subprocess.PIPE, 
        shell=True 
    )
    del os.environ['MYSQL_PWD']


    mylogger.info('Database is backed up')

def restore(settings, gitea_config):

    db_full_connection = gitea_config['database']['HOST']
    db_host = db_full_connection.split(':')[0]
    db_port = db_full_connection.split(':')[1]
    db_username = gitea_config['database']['USER']
    db_password = gitea_config['database']['PASSWD']
    db_database = gitea_config['database']['NAME']
    
    os.environ["MYSQL_PWD"] = db_password

    command = f"mysql --host={db_host} --user={db_username} {db_database} < {settings.restore_tmp_folder}/dump.mysql.sql"

    subprocess.run(command, 
        stdout=subprocess.PIPE, 
        universal_newlines=True,
        shell=True 
    )

    del os.environ['MYSQL_PWD']

    mylogger.info('Database is restored')

mysql = {
    'backup': backup,
    'restore': restore
}