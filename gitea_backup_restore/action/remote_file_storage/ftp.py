import ftplib 
from pydantic import BaseSettings

class FtpConfig(BaseSettings):
    # mandatory fields
    backup_ftp_host: str
    backup_ftp_user: str
    backup_ftp_password: str

def download_file(cls):
    with ftplib.FTP(cls.backup_methode['config']['backup_ftp_host'], cls.backup_methode['config']['backup_ftp_user'], cls.backup_methode['config']['backup_ftp_password']) as ftp:
        ftp.cwd(directory)
        with open(filename, 'wb') as f:
            ftp.retrbinary('RETR ' + filename, f.write)

def upload_file(cls):
    with ftplib.FTP(cls.backup_methode['config']['backup_ftp_host'], cls.backup_methode['config']['backup_ftp_user'], cls.backup_methode['config']['backup_ftp_password']) as ftp:
        ftp.cwd(directory)
        with open(filename, 'wb') as f:
            ftp.retrbinary('RETR ' + filename, f.write)

def ensure_max_retention(cls):
    pass

FtpStorage =  {
        'config': FtpConfig,
        'download':download_file,
        'upload':upload_file,
        'ensure_max_retention':ensure_max_retention,
    }