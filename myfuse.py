# myfuse.py

import os
import errno
import stat
from fuse import FUSE, Operations

NFS_DIR = os.path.abspath("nfs")
MOUNTPOINT = "/mnt/all-projects"

class Passthrough(Operations):
    def _full_path(self, partial):
        return os.path.join(NFS_DIR, partial.lstrip("/"))

    def getattr(self, path, fh=None):
        full_path = self._full_path(path)
        if not os.path.exists(full_path):
            raise FileNotFoundError(errno.ENOENT, os.strerror(errno.ENOENT), path)

        st = os.lstat(full_path)
        attrs = dict((key, getattr(st, key)) for key in (
            'st_mode', 'st_ino', 'st_dev', 'st_nlink',
            'st_uid', 'st_gid', 'st_size', 'st_atime',
            'st_mtime', 'st_ctime'
        ))
        return attrs

    def readdir(self, path, fh):
        full_path = self._full_path(path)
        if not os.path.isdir(full_path):
            raise NotADirectoryError(errno.ENOTDIR, os.strerror(errno.ENOTDIR), path)

        entries = ['.', '..'] + os.listdir(full_path)
        for entry in entries:
            yield entry

    def open(self, path, flags):
        full_path = self._full_path(path)
        if not os.path.exists(full_path):
            raise FileNotFoundError(errno.ENOENT, os.strerror(errno.ENOENT), path)
        return os.open(full_path, os.O_RDONLY)

    def read(self, path, size, offset, fh):
        os.lseek(fh, offset, os.SEEK_SET)
        return os.read(fh, size)

    def release(self, path, fh):
        os.close(fh)
        return 0

if __name__ == '__main__':
    if not os.path.exists(MOUNTPOINT):
        os.makedirs(MOUNTPOINT)

    print(f"🔗 Mounting FUSE FS at {MOUNTPOINT}")
    FUSE(Passthrough(), MOUNTPOINT, nothreads=True, foreground=True, ro=True)
