import logging
from pathlib import Path
from unittest import mock

import pytest
from pydantic import ValidationError
import os

from gitea_backup_restore.config import Settings
from gitea_backup_restore.log import mylogger

mydebbuglogger = logging.getLogger()
mylogger.setLevel(logging.DEBUG)
# init variables definit has already been tested

@mock.patch.dict(os.environ, {}, clear=True)
def test_system_no_env(
):
    with pytest.raises(ValidationError) as exc_info:
        mydebbuglogger.info(Settings().dict())

@mock.patch.dict(os.environ, {
    "BACKUP_ENABLE": "True",
    "BACKUP_METHODE": "s3",
    "BACKUP_FILENAME": "",
    "BACKUP_TARGET_FOLDER": "gitea_backup"
    }, clear=True)
def test_system_s3_no_s3_config():
    with pytest.raises(ValidationError) as exc_info:
        mydebbuglogger.info(Settings().dict())

@mock.patch.dict(os.environ, {
    "BACKUP_ENABLE": "True",
    "BACKUP_METHODE": "s3",
    "BACKUP_FILENAME": "",
    "BACKUP_TARGET_FOLDER": "gitea_backup",
    "ENDPOINT_URL": "gitea_backup",
    "AWS_ACCESS_KEY_ID": "gitea_backup",
    "AWS_SECRET_ACCESS_KEY": "gitea_backup",
    "BUCKET": "gitea_backup"
    }, clear=True)
def test_system_s3():
    mydebbuglogger.info(Settings().dict())
    return Settings()
    
@mock.patch.dict(os.environ, {
    "BACKUP_ENABLE": "True",
    "BACKUP_METHODE": "ftp",
    "BACKUP_FILENAME": "",
    "BACKUP_TARGET_FOLDER": "gitea_backup"
    }, clear=True)
def test_system_ftp_no_ftp_config():
    with pytest.raises(ValidationError) as exc_info:
        mydebbuglogger.info(Settings().dict())


@mock.patch.dict(os.environ, {
    "BACKUP_ENABLE": "True",
    "BACKUP_METHODE": "ftp",
    "BACKUP_FILENAME": "",
    "BACKUP_TARGET_FOLDER": "gitea_backup",
    "BACKUP_FTP_HOST": "gitea_backup",
    "BACKUP_FTP_USER": "gitea_backup",
    "BACKUP_FTP_PASSWORD": "gitea_backup"
    }, clear=True)
def test_system_ftp():
    mydebbuglogger.info(Settings().dict())
    return Settings()