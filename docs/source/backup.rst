Backup
======

Shellflow automatically backup small input/output files. If a size of
file is smaller than 20MB, the file will be compressed with gzip and
backuped to ``shellflow-wf/__backup``. Shellflow rename filename with
SHA256 hash result. Currently no method to investigate backup file name
expect reading log file directly.
