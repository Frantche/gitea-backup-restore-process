import boto3
from botocore.exceptions import NoCredentialsError

from pydantic import BaseSettings
from gitea_backup_restore.log import mylogger

class S3Config(BaseSettings):
    # mandatory fields
    endpoint_url: str
    aws_access_key_id: str
    aws_secret_access_key: str
    verify: bool = True
    signature_version: str = 's3v4'
    bucket: str
    backup_filename:str
    prefix:str=""
    
def get_client():

    s3_conf = S3Config()

    return boto3.resource(
        's3', 
        endpoint_url=s3_conf.endpoint_url,
        aws_access_key_id=s3_conf.aws_access_key_id,
        aws_secret_access_key=s3_conf.aws_secret_access_key,
        aws_session_token=None,
        config=boto3.session.Config(signature_version=s3_conf.signature_version),
        verify=s3_conf.verify
        )
 

def ensure_max_retention(setting):

    if setting.backup_max_retention>0:
        s3 = get_client()
        s3_conf = S3Config()

        my_bucket = s3.Bucket(s3_conf.bucket)

        files = my_bucket.objects.filter(Prefix=setting.backup_prefix)
        files = [
                obj.key for obj in sorted(
                    files, 
                    key=lambda x: x.last_modified,
                    reverse=True
                    )
                ]
        
        if len(files) > setting.backup_max_retention:

            try:
                for file in files[setting.backup_max_retention:]:

                    if file != setting.backup_filename: 
                        response = s3.Object(s3_conf.bucket,file).delete()

                        mylogger.info(f'Delete file from S3 : {file}')
                        mylogger.info(response)
            
            except Exception  as e:
                mylogger.exception(e)
        

def download_file(setting):
    s3 = get_client()
    s3_conf = S3Config()

    s3.meta.client.download_file(s3_conf.bucket,  s3_conf.backup_filename, setting.restore_tmp_filename)
    mylogger.info('Download OK')

def check_config():
    return S3Config()

def upload_file(setting):
    s3 = get_client()
    s3_conf = S3Config()
    try:
        s3.Bucket(s3_conf.bucket).upload_file(setting.backup_tmp_filename, setting.backup_tmp_remote_filename)
        mylogger.info("Upload Successful")
        return True
    except FileNotFoundError:
        mylogger.info("The file was not found")
        return False
    except NoCredentialsError:
        mylogger.info("Credentials not available")
        return False

S3Storage =  {
        'config': check_config,
        'download':download_file,
        'upload':upload_file,
        'ensure_max_retention':ensure_max_retention,
    }

