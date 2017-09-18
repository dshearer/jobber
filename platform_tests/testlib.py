import subprocess as sp
import os
import shutil
import tempfile
import pwd

_NORMUSER = 'normuser'

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
        self._root_jobfile_path = '/root/.jobber'
        self._normuser_jobfile_path = '/home/' + _NORMUSER + '/.jobber'
        self._jobber_path = \
            sp_check_output(['find', '/usr', '-name', \
                             'jobber', '-type', 'f']).strip()
        self._tmpfile_dir = '/JobberTestTmp'

    def make_tempfile_dir(self):
        # make temp-file dir
        os.mkdir(self._tmpfile_dir)
        os.chmod(self._tmpfile_dir, 0777)

    def rm_tempfile_dir(self):
        shutil.rmtree(self._tmpfile_dir)
    
    def make_tempfile(self):
        fd, path = tempfile.mkstemp(dir=self._tmpfile_dir)
        os.close(fd)
        os.chmod(path, 0666)
        return path
    
    def make_jobfile(self, job_name, cmd, time="*", notify_prog=None):
        jobs_sect = """[jobs]
- name: {job_name}
  cmd: {cmd}
  time: '{time}'
  notifyOnError: true
""".format(job_name=job_name, cmd=cmd, time=time)
        if notify_prog is None:
            return jobs_sect
        else:
            prefs_sect = """[prefs]
notifyProgram: {notify_prog}

""".format(notify_prog=notify_prog)
            return prefs_sect + jobs_sect

    def install_root_jobfile(self, contents):
        '''
        :return: Number of jobs loaded.
        '''
        
        # make jobfile
        with open(self._root_jobfile_path, 'w') as f:
            f.write(contents)

        # reload it
        output = sp_check_output([self._jobber_path, 'reload'])
        return int(output.split()[1])

    def install_normuser_jobfile(self, contents):
        '''
        :return: Number of jobs loaded.
        '''
        
        # make jobfile
        print("Installing jobfile at {path}".\
              format(path=self._normuser_jobfile_path))
        pwnam = pwd.getpwnam(_NORMUSER)
        os.setegid(pwnam.pw_gid)
        os.seteuid(pwnam.pw_uid)
        with open(self._normuser_jobfile_path, 'w') as f:
            f.write(contents)
        os.seteuid(0)
        os.setegid(0)

        # reload it
        output = sp_check_output(['sudo', '-u', _NORMUSER, \
                                  self._jobber_path, 'reload'])
        return int(output.split()[1])

    def rm_jobfiles(self):
        # rm jobfile
        try:
            os.unlink(self._root_jobfile_path)
        except OSError: pass
        try:
            os.unlink(self._normuser_jobfile_path)
        except OSError: pass
        
        # reload it
        sp.check_call([self._jobber_path, 'reload', '-a'], \
                      stdout=open(os.devnull, 'w'))
    
    def jobber_log(self):
        return sp_check_output([self._jobber_path, 'log'])
    
    def pause_job(self, job):
        sp.check_call([self._jobber_path, 'pause', job], \
                      stdout=open(os.devnull, 'w'))
    
    def resume_job(self, job):
        sp.check_call([self._jobber_path, 'resume', job], \
                      stdout=open(os.devnull, 'w'))
    
    def test_job(self, job):
        sp.check_call([self._jobber_path, 'test', job], \
                      stdout=open(os.devnull, 'w'))
    
    def chmod(self, path, mode):
        os.chmod(path, int(mode, base=8))
        stat = os.stat(path)
        print("Mode of {path} is now {mode}".\
              format(path=path, mode=oct(stat.st_mode & 0777)))
    
    def chown(self, path, user):
        pwnam = pwd.getpwnam(user)
        os.chown(path, pwnam.pw_uid, pwnam.pw_gid)
    
    def _check_jobber_list_output(self, output, exp_job_names):
        lines = output.split("\n")
        listed_jobs = set([line.split()[0] for line in lines[1:]])
        exp_job_names = set(exp_job_names.split(","))
        if listed_jobs != exp_job_names:
            msg = "Expected listed jobs to be {exp}, but was {act}".\
                format(exp=exp_job_names, act=listed_jobs)
            raise AssertionError(msg)
    
    def jobber_list_as_root_should_return(self, job_names, \
                                          all_users=False):
        # do 'jobber list'
        all_users = bool(all_users)
        args = [self._jobber_path, 'list']
        if all_users:
            args.append('-a')
        output = sp_check_output(args).strip()
        
        # check output
        self._check_jobber_list_output(output, job_names)
    
    def jobber_list_as_normuser_should_return(self, job_names, \
                                              all_users=False):
        # do 'jobber list'
        output = sp_check_output(['sudo', '-u', _NORMUSER, \
                                  self._jobber_path, 'list']).strip()
        
        # check output
        self._check_jobber_list_output(output, job_names)
    
    def nbr_of_lines_in_string_should_be(self, string, nbr, msg=None):
        lines = string.split("\n")
        if len(lines) != int(nbr):
            base_msg = ("Number of lines in string should be {nbr}, " \
                   "but was {actual}").format(nbr=nbr, 
                                              actual=len(lines))
            if msg is None:
                raise AssertionError(base_msg)
            else:
                raise AssertionError("{msg}: {base_msg}".\
                                    format(msg=msg, base_msg=base_msg))