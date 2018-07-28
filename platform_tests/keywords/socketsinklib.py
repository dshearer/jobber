import socket
import json
import subprocess

_IPC_SOCKET_PATH = "/var/jobber/0/cmd.sock"

class socketsinklib(object):
    ROBOT_LIBRARY_VERSION = 1.0

    def dump_socket_to_disk(self, proto, address, output_path):
        if proto == 'unix':
            cmd = ['socat', '-u', 'UNIX-CONNECT:' + address, 'CREATE:' + output_path]
        else:
            cmd = ['socat', '-u', 'TCP:localhost' + address, 'CREATE:' + output_path]
        return subprocess.Popen(cmd)

    def terminate_process(self, proc):
        proc.terminate()
