from .ftp import FtpStorage
from .s3 import S3Storage

dict_provider={
    's3': S3Storage,
    'ftp':FtpStorage
}

def download(setting):
    dict_provider[setting.backup_methode]['download'](setting)

def upload(setting):
    dict_provider[setting.backup_methode]['upload'](setting)

def ensure_max_retention(setting):
    dict_provider[setting.backup_methode]['ensure_max_retention'](setting)
