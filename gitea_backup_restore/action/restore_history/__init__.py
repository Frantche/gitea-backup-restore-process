import os

def increment(settings, backup_filename):
    with open(settings.backup_file_log,'a+') as f:
        f.write(f'{backup_filename}\n')

def check(settings):
    result = False

    if os.path.exists(settings.backup_file_log):
        with open(settings.backup_file_log,'r') as f:
            if settings.backup_filename in f.read():
                result = True
                return result
    return result