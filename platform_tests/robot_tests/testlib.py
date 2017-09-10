import subprocess as sp
import os
import shutil
import tempfile

def sp_check_output(args):
    proc = sp.Popen(args, stdout=sp.PIPE)
    out, _ = proc.communicate()
    if proc.returncode != 0:
        raise sp.CalledProcessError(proc.returncode, args, out)
    return out

class testlib(object):
    ROBOT_LIBRARY_VERSION = 1.0

    def __init__(self):
        # get paths to stuff
        self._jobfile_path = os.path.expanduser('~/.jobber')
        self._jobber_path = \
            sp_check_output(['find', '/usr', '-name', \
                             'jobber', '-type', 'f']).strip()
        self._tmpfile_dir = '/root/JobberTest'

    def make_tempfile_dir(self):
        # make temp-file dir
        os.mkdir(self._tmpfile_dir)

    def rm_tempfile_dir(self):
        shutil.rmtree(self._tmpfile_dir)
    
    def make_tempfile(self):
        fd, path = tempfile.mkstemp(dir=self._tmpfile_dir)
        os.close(fd)
        return path
    
    def make_jobfile(self, job_name, cmd, notify_prog=None):
        jobs_sect = """[jobs]
- name: {job_name}
  cmd: {cmd}
  time: '*'
  notifyOnError: true
""".format(job_name=job_name, cmd=cmd)
        if notify_prog is None:
            return jobs_sect
        else:
            prefs_sect = """[prefs]
notifyProgram: {notify_prog}

""".format(notify_prog=notify_prog)
            return prefs_sect + jobs_sect

    def install_jobfile(self, contents):
        # make jobfile
        with open(self._jobfile_path, 'w') as f:
            f.write(contents)

        # reload it
        sp.check_call([self._jobber_path, 'reload'], \
                      stdout=open(os.devnull, 'w'))

    def rm_jobfile(self):
        # rm jobfile
        try:
            os.unlink(self._jobfile_path)
        except OSError: pass
        
        # reload it
        sp.check_call([self._jobber_path, 'reload'], \
                      stdout=open(os.devnull, 'w'))
    
    def jobber_list(self):
        return sp_check_output([self._jobber_path, 'list'])
    
    def jobber_log(self):
        return sp_check_output([self._jobber_path, 'log'])
    
    def pause_job(self, job):
        sp.check_call([self._jobber_path, 'pause', job], \
                      stdout=open(os.devnull, 'w'))
    
    def resume_job(self, job):
        sp.check_call([self._jobber_path, 'resume', job], \
                      stdout=open(os.devnull, 'w'))
    
    def chmod(self, path, mode):
        os.chmod(path, int(mode, base=8))
        stat = os.stat(path)
        print("Mode of {path} is now {mode}".\
              format(path=path, mode=oct(stat.st_mode & 0777)))