import subprocess as sp
import os
import stat
import shutil
import tempfile
import pwd
import time
import json
import yaml

_NORMUSER = 'normuser'
_RUNNER_LOG_FILE_FOR_ROOT = '/root/.jobber-log'
_RUNNER_LOG_FILE_FOR_NORMUSER = '/home/{0}/.jobber-log'.\
    format(_NORMUSER)
_CONFIG_PATH = '/etc/jobber.conf'
_OLD_CONFIG_PATH = '/etc/jobber.conf.old'

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

_OCTAL_777 = int('777', base=8)
_OCTAL_755 = int('755', base=8)
_OCTAL_600 = int('600', base=8)

class _ProcInfo(object):
    '''Info about a process'''
    def __init__(self, pid, username, uid, tty, program):
        self.pid = pid
        self.username = username
        self.uid = uid
        self.tty = tty
        self.program = program

    def __eq__(self, other):
        if self.pid != other.pid:
            return False
        if self.username != other.username:
            return False
        if self.uid != other.uid:
            return False
        if self.tty != other.tty:
            return False
        if self.program != other.program:
            return False
        return True

    def __ne__(self, other):
        return not (self == other)

    def __lt__(self, other):
        return (self.pid, self.username, self.uid, self.tty, self.program) < \
            (other.pid, other.username, other.uid, other.tty, other.program)
    
    def __repr__(self):
        return "[{} {} {} {} {}]".format(self.pid, self.username, self.uid, \
            self.tty, self.program)

def _get_proc_info(program_name):
    '''
    :return: List of instances of _ProcInfo.
    '''
    def split_by_whitespace(s, maxsplit):
        s = ' '.join(s.split())
        return s.split(' ', maxsplit)

    args = ['ps', '-ax', '-o', 'pid,user,uid,tty,command']
    proc = sp.Popen(args, stdout=sp.PIPE, stderr=sp.PIPE)
    infos = []
    skipped_first = False
    for line in proc.stdout:
        line = bytes(line).decode('ascii')
        if not skipped_first:
            skipped_first = True
            continue
        parts = split_by_whitespace(line.strip(), 4)
        pid, username, uid, tty, program = parts
        if program_name not in program:
            continue
        # Darwin uses '??' to mean 'no TTY'; Linux uses '?'
        if tty == '?' or tty == '??':
            tty = None
        infos.append(_ProcInfo(pid, username, uid, tty, program))
    if proc.wait() != 0:
        print(proc.stderr.read())
        return Exception("ps returned non-0")
    return infos

def sp_check_output(args):
    proc = sp.Popen(args, stdout=sp.PIPE, stderr=sp.PIPE)
    out, err = proc.communicate()
    out = bytes(out).decode('ascii')
    err = bytes(err).decode('ascii')
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

def program_exists(name):
    try:
        find_program(name)
    except:
        return False
    else:
        return True

class SystemDServiceCtl(object):
    def restart_jobber(self):
        sp_check_output(['systemctl', 'restart', 'jobber'])

    def get_jobber_status(self):
        return sp_check_output(['systemctl', 'status', 'jobber'])

    def get_jobbermaster_logs(self):
        return sp_check_output(['journalctl', '-u', 'jobber'])

class LaunchctlServiceCtl(object):
    def restart_jobber(self):
        sp_check_output(['launchctl', 'stop', 'info.nekonya.jobber'])
        sp_check_output(['launchctl', 'start', 'info.nekonya.jobber'])

    def get_jobber_status(self):
        return 'unknown'

    def get_jobbermaster_logs(self):
        path = '/var/log/system.log'
        if not os.path.isfile(path):
            return '[Unknown path to system log]'
        args = ['tail', '-n', '20', path]
        lines = sp_check_output(args).split('\n')
        lines = [l for l in lines if 'jobbermaster' in l]
        return '/n'.join(lines)

class BrewServiceCtl(object):
    def restart_jobber(self):
        sp_check_output(['brew', 'services', 'restart', 'jobber'])

    def get_jobber_status(self):
        return 'unknown'

    def get_jobbermaster_logs(self):
        path = '/var/log/system.log'
        if not os.path.isfile(path):
            return '[Unknown path to system log]'
        args = ['tail', '-n', '20', path]
        lines = sp_check_output(args).split('\n')
        lines = [l for l in lines if 'jobbermaster' in l]
        return '/n'.join(lines)

def parse_list_arg(s):
    parts = s.split(',')
    return set([p for p in parts if len(p) > 0])

class testlib(object):
    ROBOT_LIBRARY_VERSION = 1.0

    def __init__(self):
        # get paths to stuff
        self._jobber_path = find_program('jobber')
        self._python_path = find_program('python')
        self._tmpfile_dir = None
        self._next_tmpfile_nbr = 1

        # make service control object
        if program_exists('systemctl'):
            self._servicectl = SystemDServiceCtl()
        elif program_exists('brew') and 'homebrew.mxcl.jobber.plist' in \
                sp_check_output(['brew', 'services']):
            self._servicectl = BrewServiceCtl()
        elif program_exists('launchctl'):
            self._servicectl = LaunchctlServiceCtl()
        else:
            raise Exception("Cannot determine how to control Jobber service")

    @property
    def _root_jobfile_path(self):
        root_entry = pwd.getpwuid(0)
        return os.path.join(root_entry.pw_dir, '.jobber')

    @property
    def _normuser_jobfile_path(self):
        normuser_entry = pwd.getpwnam(_NORMUSER)
        return os.path.join(normuser_entry.pw_dir, '.jobber')

    def make_tempfile_dir(self):
        # make temp-file dir
        self._tmpfile_dir = tempfile.mkdtemp()
        os.chmod(self._tmpfile_dir, _OCTAL_777)

    def rm_tempfile_dir(self):
        shutil.rmtree(self._tmpfile_dir)
        self._tmpfile_dir = None

    def make_tempfile(self, create=False):
        path = os.path.join(self._tmpfile_dir,
            "tmp-{0}".format(self._next_tmpfile_nbr))
        self._next_tmpfile_nbr += 1
        if create:
            open(path, "w").close()
        return path

    def restart_service(self):
        # restart jobber service
        try:
            self._servicectl.restart_jobber()
        except Exception as e:
            self.print_debug_info()
            raise e

        # wait for it to be ready
        started = False
        stop_time = time.time() + 20
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
            self.print_debug_info()
            raise AssertionError(msg)

        # sometimes not all jobberrunner procs have started yet
        time.sleep(2)

    def print_debug_info(self):
        log = ''

        # get service status
        log += "Jobber service status:\n"
        try:
            log += self._servicectl.get_jobber_status()
        except Exception as e:
            log += "[{0}]".format(e)

        # get syslog msgs
        log += "\n\njobbermaster logs:\n"
        try:
            log += self._servicectl.get_jobbermaster_logs()
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

        # get config file
        log += "\nConfig:\n"
        try:
            with open(_CONFIG_PATH) as f:
                tmp = f.read()
                if len(tmp) == 0:
                    log += "[empty]"
                else:
                    log += tmp
        except Exception as e:
            log += "[{0}]".format(e)

        print(log)

    def make_jobfile(self, job_name, cmd, time="*", \
                     notify_on_error=False, notify_on_success=False,
                     notify_output_path=None, file_run_log_path=None,
                     stdout_output_dir=None, stdout_output_max_age=None,
                     stderr_output_dir=None, stderr_output_max_age=None,
                     unix_result_sink_path=None, tcp_result_sink_port=None,
                     log_path=None):

        jobfile = {
            'version': '1.4',
            'prefs': {},
            'jobs': {}
        }

        # make jobs section
        job = {
            'cmd': cmd,
            'time': time,
            'notifyOnError': [],
            'notifyOnSuccess': []
        }
        jobfile['jobs'][job_name] = job

        # make prefs section
        if log_path is not None:
            jobfile['prefs']['logPath'] = log_path

        def install_result_sink(sink):
            for job_name in jobfile['jobs']:
                job = jobfile['jobs'][job_name]
                if notify_on_error:
                    job['notifyOnError'].append(sink)
                if notify_on_success:
                    job['notifyOnSuccess'].append(sink)

        if notify_output_path is not None:
            # make notify program
            output_path = self.make_tempfile()
            notify_prog = _NOTIFY_PROGRAM.format(notify_output_path=\
                                                 notify_output_path)
            shebang = "#!" + self._python_path + "\n"
            notify_prog = shebang + notify_prog
            notify_prog_path = self.make_tempfile()
            with open(notify_prog_path, 'w') as f:
                f.write(notify_prog)
            os.chmod(notify_prog_path, _OCTAL_755)

            # make result sink
            result_sink = {
                'type': 'program',
                'path': notify_prog_path
            }
            install_result_sink(result_sink)

            print("Contents of {0}:\n{1}".\
                  format(notify_prog_path, notify_prog))

        if file_run_log_path is not None:
            jobfile['prefs']['runLog'] = {'type': 'file', 'path': file_run_log_path}

        if stdout_output_dir is not None:
            result_sink = {
                'type': 'filesystem',
                'path': stdout_output_dir,
                'maxAgeDays': int(stdout_output_max_age),
                'data': ['stdout']
            }
            install_result_sink(result_sink)

        if stderr_output_dir is not None:
            result_sink = {
                'type': 'filesystem',
                'path': stderr_output_dir,
                'maxAgeDays': int(stderr_output_max_age),
                'data': ['stderr']
            }
            install_result_sink(result_sink)

        if unix_result_sink_path is not None:
            result_sink = {
                'type': 'socket',
                'proto': 'unix',
                'address': unix_result_sink_path
            }
            install_result_sink(result_sink)

        if tcp_result_sink_port is not None:
            result_sink = {
                'type': 'socket',
                'proto': 'tcp',
                'address': tcp_result_sink_port
            }
            install_result_sink(result_sink)

        return json.dumps(jobfile)

    def install_jobfile(self, contents, for_root=False, reload=True, exp_num_jobs=1):
        # make jobfile
        if for_root:
            with open(self._root_jobfile_path, 'w') as f:
                f.write(contents)
        else:
            pwnam = pwd.getpwnam(_NORMUSER)
            os.setegid(pwnam.pw_gid)
            os.seteuid(pwnam.pw_uid)
            try:
                with open(self._normuser_jobfile_path, 'w') as f:
                    f.write(contents)
                os.chmod(self._normuser_jobfile_path, _OCTAL_600)
            finally:
                os.seteuid(0)
                os.setegid(0)

        # reload it
        if reload:
            if for_root:
                output = sp_check_output([self._jobber_path, 'reload'])
            else:
                output = sp_check_output(['sudo', '-u', _NORMUSER, \
                                          self._jobber_path, 'reload'])
            print(output)
            num_jobs = int(output.split()[1])

            # check number of loaded jobs
            if num_jobs != exp_num_jobs:
                msg = ("Failed to load jobfile: expected to load {0} jobs " + \
                    "but loaded {1}").format(exp_num_jobs, num_jobs)
                raise AssertionError(msg)

    def rm_jobfiles(self):
        # rm jobfile
        if os.path.exists(self._root_jobfile_path):
            os.unlink(self._root_jobfile_path)
        if os.path.exists(self._normuser_jobfile_path):
            os.unlink(self._normuser_jobfile_path)

    def jobber_log(self, as_root=False, all_users=False):
        '''
        :return: Number of run log entries
        '''
        if as_root:
            args = [self._jobber_path, 'log']
        else:
            args = ['sudo', '-u', _NORMUSER, self._jobber_path, 'log']
        if all_users:
            args.append('-a')
        output = sp_check_output(args).strip()

        # get run log entries
        lines = output.split("\n")
        if len(lines) == 0:
            msg = "Expected output to have some lines: \"{0}\"".\
                format(output)
            raise AssertionError(msg)
        return len(lines[1:]) # ignore header line

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
              format(path=path, mode=oct(stat.st_mode & _OCTAL_777)))

    def chown(self, path, user):
        pwnam = pwd.getpwnam(user)
        os.chown(path, pwnam.pw_uid, pwnam.pw_gid)

    def set_config(self, include_users='', exclude_users=''):
        with open(_CONFIG_PATH) as f:
            config = yaml.safe_load(f)

        # make config
        config['users-include'] = []
        config['users-exclude'] = []
        for user in parse_list_arg(include_users):
            config['users-include'].append({'username': user})
        for user in parse_list_arg(exclude_users):
            config['users-exclude'].append({'username': user})

        # remove old config
        try:
            os.rename(_CONFIG_PATH, _OLD_CONFIG_PATH)
        except OSError as e:
            if e.errno == 2:
                pass
            else:
                raise e

        # write to disk
        with open(_CONFIG_PATH, 'w') as f:
            yaml.safe_dump(config, f)

    def restore_prefs(self):
        try:
            os.rename(_OLD_CONFIG_PATH, _CONFIG_PATH)
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

    def _parse_job_output_file_name(self, fname):
        '''
        :return: (timestamp, suffix)
        '''

        err_msg = "Invalid job output file name: \"{0}\"".format(fname)
        parts = fname.split('.')
        if len(parts) != 2:
            raise AssertionError(err_msg)
        try:
            ts = int(parts[0])
        except ValueError:
            raise AssertionError(err_msg)
        return ts, parts[1]

    def max_job_output_file_age(self, dir_path):
        '''
        Get the maximum age (in days) of any job output file in the specified
        directory.
        '''

        # list dir contents
        for _, _, fn in os.walk(dir_path):
            filenames = fn
            break

        if len(filenames) == 0:
            return None

        # examine timestamps
        now = time.time()
        day_sec = 60 * 60 * 24
        ages = []
        for fname in filenames:
            ts, _ = self._parse_job_output_file_name(fname)
            age = int((now - ts)/day_sec)
            ages.append(age)
            print("Age of {0}: {1} days".format(fname, age))

        return max(ages)

    def decrease_job_output_files_timestamps(self, dir_path, days):
        '''
        Decrease the timestamps on the job-output files in the specified
        directory by the specified amount of days.
        '''

        # list dir contents
        filenames = None
        for _, _, fn in os.walk(dir_path):
            filenames = fn
            break
        if filenames is None:
            raise AssertionError("No job output files")

        # rename files
        day_sec = 60 * 60 * 24
        for fname in filenames:
            ts, suffix = self._parse_job_output_file_name(fname)
            new_fname = "{0}.{1}".format(ts - days*day_sec, suffix)
            old_path = os.path.join(dir_path, fname)
            new_path = os.path.join(dir_path, new_fname)
            os.rename(old_path, new_path)
            print("Renamed {0} to {1}".format(old_path, new_path))

    def directory_exists(self, path):
        try:
            st = os.stat(path)
            if stat.S_ISDIR(st.st_mode):
                return True
            else:
                print("{0} is not a directory".format(path))
                return False
        except OSError as e:
            if e.errno == 2:
                return False
            else:
                raise e

    def runner_proc_info(self):
        '''
        :return: List of instances of _ProcInfo.
        '''
        return _get_proc_info("jobberrunner")

    def nbr_of_runner_procs_should_be_same(self, orig_proc_info):
        new_proc_info = self.runner_proc_info()
        if len(orig_proc_info) != len(new_proc_info):
            print("Original runner procs:\n{0}".format(orig_proc_info))
            print("New runner procs:\n{0}".format(new_proc_info))
            raise AssertionError("Number of runner procs has changed!")

    def runner_procs_should_not_have_tty(self):
        # This is to avoid a particular vulnerability
        # (http://www.halfdog.net/Security/2012/TtyPushbackPrivilegeEscalation/)
        proc_info = self.runner_proc_info()
        print("proc_info: {}".format(proc_info))
        for proc in proc_info:
            if proc.tty is not None:
                print("Runner procs:\n{0}".format(proc_info))
                raise AssertionError("A runner proc has a controlling tty")

    def there_should_be_no_runner_procs(self):
        proc_info = self.runner_proc_info()
        if len(proc_info) > 0:
            print("Runner procs:\n{0}".format(proc_info))
            raise AssertionError("There are still runner procs")

    def jobber_list_should_return(self, exp_job_names, as_root=False,
                                  all_users=False):
        exp_job_names = parse_list_arg(exp_job_names)

        # do 'jobber list'
        if as_root:
            args = [self._jobber_path, 'list']
        else:
            args = ['sudo', '-u', _NORMUSER, self._jobber_path, 'list']
        if all_users:
            args.append('-a')
        print("Cmd: {0}".format(args))
        output = sp_check_output(args).strip()
        print(output)

        # check output
        lines = output.split("\n")
        if len(lines) == 0:
            msg = "Expected output to have some lines: \"{0}\"".\
                format(output)
            raise AssertionError(msg)
        listed_jobs = set([line.split()[0] for line in lines[1:]])
        if listed_jobs != exp_job_names:
            msg = "Expected listed jobs to be {exp}, but was {act}".\
                format(exp=exp_job_names, act=listed_jobs)
            raise AssertionError(msg)

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
            logs = self._servicectl.get_jobbermaster_logs()
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

    def jobberrunner_should_be_running_for_user(self, username):
        proc_info = self.runner_proc_info()
        if not any(p.username == username for p in proc_info):
            print("Runner procs:\n{0}\n".format(proc_info))
            self.print_debug_info()
            raise AssertionError("jobberrunner is not running for {0}".\
                                 format(username))

    def jobberrunner_should_not_be_running_for_user(self, username):
        proc_info = self.runner_proc_info()
        if any(p.username == username for p in proc_info):
            print("Runner procs:\n{0}\n".format(proc_info))
            self.print_debug_info()
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

    def config_file_should_exist(self):
        try:
            os.stat(_CONFIG_PATH)
        except OSError as e:
            if e.errno == 2:
                raise AssertionError("Prefs file does not exist")
            else:
                raise e

    def jobber_procs_should_not_have_inet_sockets(self):
        jobber_procs = _get_proc_info("jobber")
        for jproc in jobber_procs:
            cmd = [
                "lsof",
                "-a", # AND the options (rather than ORing them)
                "-n", # don't convert IP to domain
                "-P", # don't convert port to service name
                "-i", # show IP sockets
                "-p", jproc.pid
                ]
            proc = sp.Popen(cmd, stdout=sp.PIPE, stderr=open(os.devnull))
            output, _ = proc.communicate()
            output = bytes(output).decode('ascii').strip()
            if len(output) > 0:
                msg = "{} process has inet sockets:\n{}".\
                    format(jproc.program, output)
                raise AssertionError(msg)
