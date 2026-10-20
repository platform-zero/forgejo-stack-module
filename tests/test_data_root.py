import os
import subprocess
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).parents[1]
SOURCE = (ROOT / 'stack.config/forgejo/custom-entrypoint.sh').read_text()
PREFIX = SOURCE.split('prepare_forgejo_data_root /data', 1)[0]

class DataRoot(unittest.TestCase):
    def run_prepare(self, path, uid=None, gid=None):
        env = dict(os.environ, USER_UID=str(os.geteuid() if uid is None else uid), USER_GID=str(os.getegid() if gid is None else gid))
        return subprocess.run(['bash', '-c', PREFIX+'\nprepare_forgejo_data_root "$1"', '_', str(path)], env=env, capture_output=True, text=True)
    def test_root0700_parent_is_owned_by_server_not_wrapper_without_recursion(self):
        with tempfile.TemporaryDirectory() as temp:
            root=Path(temp);data=root/'data';data.mkdir(mode=0o700);child=data/'preserved';child.write_text('authored data');token=root/'runner-token';token.mkdir(mode=0o700)
            before=child.stat();token_before=token.stat();uid=1000 if os.geteuid()==0 else os.geteuid();gid=1000 if os.geteuid()==0 else os.getegid()
            result=self.run_prepare(data,uid,gid);self.assertEqual(0,result.returncode,result.stderr)
            self.assertEqual((uid,gid,0o700),(data.stat().st_uid,data.stat().st_gid,data.stat().st_mode&0o777))
            self.assertEqual((before.st_uid,before.st_gid,before.st_ino),(child.stat().st_uid,child.stat().st_gid,child.stat().st_ino));self.assertEqual('authored data',child.read_text())
            self.assertEqual((token_before.st_uid,token_before.st_gid,token_before.st_ino),(token.stat().st_uid,token.stat().st_gid,token.stat().st_ino))
    def test_linked_parent_refuses_without_touching_target(self):
        with tempfile.TemporaryDirectory() as temp:
            root=Path(temp);target=root/'preserved';target.mkdir();before=target.stat();link=root/'alias';link.symlink_to(target)
            self.assertNotEqual(0,self.run_prepare(link).returncode);self.assertEqual((before.st_uid,before.st_gid),(target.stat().st_uid,target.stat().st_gid))
    def test_invalid_identity_refuses_before_ownership_change(self):
        with tempfile.TemporaryDirectory() as temp:
            path=Path(temp);before=path.stat();self.assertNotEqual(0,self.run_prepare(path,uid='1000; false').returncode);self.assertEqual(before.st_uid,path.stat().st_uid)
    def test_regular_file_is_not_a_volume_root(self):
        with tempfile.TemporaryDirectory() as temp:
            path=Path(temp)/'preserved';path.write_text('content');self.assertNotEqual(0,self.run_prepare(path).returncode);self.assertEqual('content',path.read_text())
    def test_preparation_precedes_every_server_or_bootstrap_start(self):
        self.assertLess(SOURCE.index('prepare_forgejo_data_root /data'),SOURCE.index('/usr/bin/entrypoint'))
        self.assertLess(SOURCE.index('prepare_forgejo_data_root /data'),SOURCE.index('/generate-runner-token.sh &'))
        self.assertNotIn('chown -R',SOURCE)

if __name__=='__main__':unittest.main()
