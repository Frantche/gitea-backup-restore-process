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
    
    os.environ["PGPASSWORD"] = db_password

    command = f"pg_dump --host={db_host} --username={db_username} {db_database} > {settings.backup_tmp_folder}/dump.postgres.sql"

    mylogger.debug(f'dump Postgresl db command line: {command}')

    subprocess.run(command, 
        stdout=subprocess.PIPE, 
        shell=True 
    )
    del os.environ['PGPASSWORD']


    mylogger.info('Database is backed up')

def restore(settings, gitea_config):

    db_full_connection = gitea_config['database']['HOST']
    db_host = db_full_connection.split(':')[0]
    db_port = db_full_connection.split(':')[1]
    db_username = gitea_config['database']['USER']
    db_password = gitea_config['database']['PASSWD']
    db_database = gitea_config['database']['NAME']
    
    os.environ["PGPASSWORD"] = db_password

    command_drop_previous_table = f"psql --host={db_host} --username={db_username} -c 'drop owned by {{db_username}}'"
    command_restore_db = f"psql --host={db_host} --username={db_username} {db_database} < {settings.restore_tmp_folder}/dump.postgres.sql"

    subprocess.run(command_drop_previous_table, 
        stdout=subprocess.PIPE, 
        universal_newlines=True,
        shell=True 
    )
    subprocess.run(command_restore_db, 
        stdout=subprocess.PIPE, 
        universal_newlines=True,
        shell=True 
    )

    del os.environ['PGPASSWORD']

    mylogger.info('Database is restored')

postgres = {
    'backup': backup,
    'restore': restore
}
