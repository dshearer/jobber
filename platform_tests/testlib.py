import subprocess as sp
import os
import shutil
import tempfile
import pwd
import time

_NORMUSER = 'normuser'
_RUNNER_LOG_FILE_FOR_ROOT = '/root/.jobber-log'
_RUNNER_LOG_FILE_FOR_NORMUSER = '/home/{0}/.jobber-log'.\
    format(_NORMUSER)
_PREFS_PATH = '/etc/jobber.conf'
_OLD_PREFS_PATH = '/etc/jobber.conf.old'

_NOTIFY_PROGRAM = '''
import json
import sys

def main():
    data = json.load(sys.stdin)
    with open('{notify_output_path}', 'w') as f:
        f.write("succeeded: {{0}}, status: {{1}}".format(
            data['succeeded'],
            data['job']['status']
        ))

if __name__ == '__main__':
    main()
'''

def sp_check_output(args):
    proc = sp.Popen(args, stdout=sp.PIPE, stderr=sp.PIPE)
    out, err = proc.communicate()
    if proc.returncode != 0:
        msg = "{args} failed.\nStdout:\n{out}\nStderr:\n{err}".format(
                args=args,
                out=out,
                err=err
            )
        raise AssertionError(msg)
    if len(err) > 0:
        print("STDERR: {0}".format(err))
    return out

def _find_file(name, dir):
    for dirpath, dirnames, filenames in os.walk(dir):
        if name in filenames:
            return os.path.join(dirpath, name)
    return None

def find_program(name):
    dirs = ['/bin', '/sbin', '/usr/bin', '/usr/sbin', '/usr/local/bin',
            '/usr/local/sbin']
    for dir in dirs:
        path = _find_file(name, dir)
        if path is not None:
            return path
    raise Exception("Cannot find program {0}".format(name))

def using_systemd():
    try:
        find_program('systemctl')
    except:
        return False
    else:
        return True

def get_jobbermaster_logs():
    if using_systemd():
        return sp_check_output(['journalctl', '-u', 'jobber'])
    else:
        args = ['tail', '-n', '20', '/var/log/messages']
        lines = sp_check_output(args).split('\n')
        lines = [l for l in lines if 'jobbermaster' in l]
        return '/n'.join(lines)

class testlib(object):
    ROBOT_LIBRARY_VERSION = 1.0

    def __init__(self):
        # get paths to stuff
        self._root_jobfile_path = '/root/.jobber'
        self._normuser_jobfile_path = '/home/' + _NORMUSER + '/.jobber'
        self._jobber_path = find_program('jobber')
        self._python_path = find_program('python')
        self._tmpfile_dir = '/JobberTestTmp'
        self._next_tmpfile_nbr = 1

    def make_tempfile_dir(self):
        # make temp-file dir
        os.mkdir(self._tmpfile_dir)
        os.chmod(self._tmpfile_dir, 0777)

    def rm_tempfile_dir(self):
        shutil.rmtree(self._tmpfile_dir)

    def make_tempfile(self, create=False):
        path = os.path.join(self._tmpfile_dir,
            "tmp-{0}".format(self._next_tmpfile_nbr))
        self._next_tmpfile_nbr += 1
        if create:
            f = open(path, "w")
            f.close()
        return path

    def restart_service(self):
        # restart jobber service
        try:
            if using_systemd():
                sp_check_output(['systemctl', 'restart', 'jobber'])
            else:
                sp_check_output(['service', 'jobber', 'restart'])
        except Exception as e:
            self.print_debug_info()
            raise e

        # wait for it to be ready
        started = False
        stop_time = time.time() + 10
        while time.time() < stop_time and not started:
            args = [self._jobber_path, 'list']
            proc = sp.Popen(args, stdout=sp.PIPE, stderr=sp.PIPE)
            _, err = proc.communicate()
            if proc.returncode == 0:
                started = True
            else:
                time.sleep(1)
        if not started:
            msg = "Failed to start jobber service!"
            msg += " ('jobber list' returned '{0}')".\
                format(err.strip())
            raise AssertionError(msg)

        # sometimes not all jobberrunner procs have started yet
        time.sleep(2)

    def print_debug_info(self):
        log = ''

        # get service status
        log += "Jobber service status:\n"
        if using_systemd():
            args = ['systemctl', 'status', 'jobber']
        else:
            args = ['service', 'jobber', 'status']
        try:
            log += sp_check_output(args)
        except Exception as e:
            log += "[{0}]".format(e)

        # get syslog msgs
        log += "\n\njobbermaster logs:\n"
        try:
            log += get_jobbermaster_logs()
        except Exception as e:
            log += "[{0}]".format(e)

        # get jobberrunner logs
        log_files = [
            _RUNNER_LOG_FILE_FOR_ROOT,
            _RUNNER_LOG_FILE_FOR_NORMUSER,
        ]
        for lf in log_files:
            log += "\n\n{0}:\n".format(lf)
            try:
                with open(lf) as f:
                    log += f.read()
            except Exception as e:
                log += "[{0}]".format(e)

        # get prefs file
        log += "\n\nPrefs:\n"
        try:
            with open(_PREFS_PATH) as f:
                log += f.read()
        except Exception as e:
            log += "[{0}]".format(e)

        print(log)

    def make_jobfile(self, job_name, cmd, time="*", \
                     notify_on_error=False, notify_on_success=False,
                     notify_output_path=None, file_run_log_path=None):
        # make jobs section
        jobs_sect = """[jobs]
- name: {job_name}
  cmd: {cmd}
  time: '{time}'
  notifyOnError: {notify_on_error}
  notifyOnSuccess: {notify_on_success}
""".format(job_name=job_name, cmd=cmd, time=time,
           notify_on_error=str(notify_on_error).lower(),
           notify_on_success=str(notify_on_success).lower())

        # make prefs section
        prefs_sect = """[prefs]
logPath: .jobber-log
"""
        if notify_on_error or notify_on_success:
            # make notify program
            output_path = self.make_tempfile()
            notify_prog = _NOTIFY_PROGRAM.format(notify_output_path=\
                                                 notify_output_path)
            shebang = "#!" + self._python_path + "\n"
            notify_prog = shebang + notify_prog
            notify_prog_path = self.make_tempfile()
            with open(notify_prog_path, 'w') as f:
                f.write(notify_prog)
            os.chmod(notify_prog_path, 0755)

            prefs_sect += "notifyProgram: {0}\n".format(
                notify_prog_path)

            print("Contents of {0}:\n{1}".\
                  format(notify_prog_path, notify_prog))

        if file_run_log_path is not None:
            run_log_pref = """runLog:
    type: file
    path: {0}
""".format(file_run_log_path)
            prefs_sect += run_log_pref

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

    def install_normuser_jobfile(self, contents, reload=True):
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

        if reload:
        # reload it
            output = sp_check_output(['sudo', '-u', _NORMUSER, \
                                      self._jobber_path, 'reload'])
            return int(output.split()[1])
        else:
            return 0

    def rm_jobfiles(self):
        # rm jobfile
        if os.path.exists(self._root_jobfile_path):
            os.unlink(self._root_jobfile_path)
        if os.path.exists(self._normuser_jobfile_path):
            os.unlink(self._normuser_jobfile_path)

    def jobber_log_as_root(self, all_users=False):
        args = [self._jobber_path, 'log']
        if all_users:
            args.append('-a')
        return sp_check_output(args).strip()

    def jobber_log_as_normuser(self, all_users=False):
        args = ['sudo', '-u', _NORMUSER, self._jobber_path, 'log']
        if all_users:
            args.append('-a')
        return sp_check_output(args).strip()

    def pause_job(self, job):
        sp_check_output([self._jobber_path, 'pause', job])

    def resume_job(self, job):
        sp_check_output([self._jobber_path, 'resume', job])

    def test_job(self, job):
        sp_check_output([self._jobber_path, 'test', job])

    def jobber_init(self):
        sp_check_output([self._jobber_path, 'init'])

    def chmod(self, path, mode):
        os.chmod(path, int(mode, base=8))
        stat = os.stat(path)
        print("Mode of {path} is now {mode}".\
              format(path=path, mode=oct(stat.st_mode & 0777)))

    def chown(self, path, user):
        pwnam = pwd.getpwnam(user)
        os.chown(path, pwnam.pw_uid, pwnam.pw_gid)

    def set_prefs(self, include_users='', exclude_users=''):
        # make prefs
        prefs = "users-include:\n"
        for user in include_users.split(','):
            prefs += "    - username: {0}\n".format(user)
        prefs = "users-exclude:\n"
        for user in exclude_users.split(','):
            prefs += "    - username: {0}\n".format(user)

        # remove old prefs
        try:
            os.rename(_PREFS_PATH, _OLD_PREFS_PATH)
        except OSError as e:
            if e.errno == 2:
                pass
            else:
                raise e

        # write to disk
        with open(_PREFS_PATH, 'w') as f:
            f.write(prefs)

    def restore_prefs(self):
        try:
            os.rename(_OLD_PREFS_PATH, _PREFS_PATH)
        except OSError as e:
            if e.errno == 2:
                pass
            else:
                raise e

    def kill_master_proc(self):
        # get pid of jobbermaster
        master_pid = sp_check_output(['pgrep', 'jobbermaster']).strip()
        if len(master_pid) == 0:
            raise AssertionError("jobbermaster isn't running")

        # kill it
        sp_check_output(['kill', '-9', master_pid])
        time.sleep(1)

    def runner_proc_info(self):
        args = ['ps', '-C', 'jobberrunner', '-o', 'user,uid,tty']
        proc = sp.Popen(args, stdout=sp.PIPE, stderr=sp.PIPE)
        output, _ = proc.communicate()
        records = [line for line in output.split('\n')[1:] \
                   if len(line.strip()) > 0]
        records.sort()
        return '\n'.join(records)

    def nbr_of_runner_procs_should_be_same(self, orig_proc_info):
        new_proc_info = self.runner_proc_info()
        if orig_proc_info != new_proc_info:
            print("Original runner procs:\n{0}".format(orig_proc_info))
            print("New runner procs:\n{0}".format(new_proc_info))
            raise AssertionError("Number of runner procs has changed!")

    def runner_procs_should_not_have_tty(self):
        # This is to avoid a particular vulnerability
        # (http://www.halfdog.net/Security/2012/TtyPushbackPrivilegeEscalation/)
        proc_info = self.runner_proc_info()
        for line in proc_info.split('\n'):
            try:
                tty = line.split()[2]
            except IndexError as _:
                print("Error: " + line)
                raise
            if tty != '?':
                print("Runner procs:\n{0}".format(proc_info))
                raise AssertionError("A runner proc has a controlling tty")

    def there_should_be_no_runner_procs(self):
        proc_info = self.runner_proc_info()
        if len(proc_info) > 0:
            print("Runner procs:\n{0}".format(proc_info))
            raise AssertionError("There are still runner procs")

    def _check_jobber_list_output(self, output, exp_job_names):
        lines = output.split("\n")
        if len(lines) <= 1:
            msg = "Expected output to have multiple lines: \"{0}\"".\
                format(output)
            raise AssertionError(msg)
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
        print("Cmd: {0}".format(args))
        output = sp_check_output(args).strip()
        print(output)

        # check output
        self._check_jobber_list_output(output, job_names)

    def jobber_list_as_normuser_should_return(self, job_names, \
                                              all_users=False):
        # do 'jobber list'
        args = ['sudo', '-u', _NORMUSER, self._jobber_path, 'list']
        if all_users:
            args.append('-a')
        output = sp_check_output(args).strip()
        print(output)

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

    def nbr_of_lines_in_string_should_be_greater_than(self, string,
                                                      nbr, msg=None):
        lines = string.split("\n")
        if len(lines) <= int(nbr):
            base_msg = ("Number of lines in string should be > {nbr}, " \
                   "but was {actual}").format(nbr=nbr,
                                              actual=len(lines))
            if msg is None:
                raise AssertionError(base_msg)
            else:
                raise AssertionError("{msg}: {base_msg}".\
                                    format(msg=msg, base_msg=base_msg))

    def jobbermaster_has_not_crashed(self):
        try:
            logs = get_jobbermaster_logs()
        except:
            pass
        else:
            if "panic" in logs:
                print(logs)
                raise AssertionError("jobbermaster crashed")

    def jobberrunner_for_root_has_not_crashed(self):
        try:
            with open(_RUNNER_LOG_FILE_FOR_ROOT) as f:
                logs = f.read()
        except:
            pass
        else:
            if "panic" in logs:
                print(logs)
                raise AssertionError("jobberrunner for root crashed")

    def jobberrunner_for_normuser_has_not_crashed(self):
        try:
            with open(_RUNNER_LOG_FILE_FOR_NORMUSER) as f:
                logs = f.read()
        except:
            pass
        else:
            if "panic" in logs:
                print(logs)
                raise AssertionError("jobberrunner for normuser crashed")

    def jobberrunner_should_not_be_running_for_user(self, username):
        proc_info = self.runner_proc_info()
        if username in proc_info:
            print("Runner procs:\n{0}".format(proc_info))
            print("Logs: \n" + get_jobbermaster_logs())
            with open(_PREFS_PATH) as f:
                print("Prefs:\n{0}".format(f.read()))
            raise AssertionError("jobberrunner is running for {0}".\
                                 format(username))

    def jobfile_for_root_should_exist(self):
        try:
            os.stat(self._root_jobfile_path)
        except OSError as e:
            if e.errno == 2:
                raise AssertionError("Jobfile for root does not exist")
            else:
                raise e

    def jobfile_for_root_should_not_exist(self):
        try:
            os.stat(self._root_jobfile_path)
            raise AssertionError("Jobfile for root exists")
        except OSError as e:
            if e.errno != 2:
                raise e

    def prefs_file_should_exist(self):
        try:
            os.stat(_PREFS_PATH)
        except OSError as e:
            if e.errno == 2:
                raise AssertionError("Prefs file does not exist")
            else:
                raise e
