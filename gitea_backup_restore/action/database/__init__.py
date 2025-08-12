from .mysql import mysql
from .postgres import postgres

db_type = {
    'mysql': mysql,
    'postgres': postgres
}


def backup(settings, gitea_config):

    db_type_import = db_type[gitea_config['database']['DB_TYPE']]
    db_type_import['backup'](settings, gitea_config)

def restore(settings, gitea_config):

    db_type_import = db_type[gitea_config['database']['DB_TYPE']]
    db_type_import['restore'](settings, gitea_config)
