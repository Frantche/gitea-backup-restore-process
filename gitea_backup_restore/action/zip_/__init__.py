import zipfile
import pathlib

from gitea_backup_restore.log import mylogger

def extract(settings):

    with zipfile.ZipFile(settings.restore_tmp_filename, mode="r") as archive:
        archive.extractall(settings.restore_tmp_folder)

    mylogger.info('Zip unpack : OK')

def compact(settings):

    directory = pathlib.Path(settings.backup_tmp_folder)

    with zipfile.ZipFile(settings.backup_tmp_filename, mode="w") as archive:
        for file_path in directory.rglob("*"):
            archive.write(
                file_path,
                arcname=file_path.relative_to(directory)
            )

    mylogger.info('Zip compact : OK')