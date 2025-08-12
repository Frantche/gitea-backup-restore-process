from configparser import ConfigParser, ExtendedInterpolation, MissingSectionHeaderError

from gitea_backup_restore.log import mylogger
from pathlib import Path


def read_app_init(settings):

    parser =  ConfigParser(interpolation=ExtendedInterpolation())

    txt = Path(settings.app_ini_path).read_text()
    text_add_default = f"[DEFAULT]\n{txt}"

    try:
        parser.read_string(txt)
    except MissingSectionHeaderError:

        parser.read_string(text_add_default)

    return parser
