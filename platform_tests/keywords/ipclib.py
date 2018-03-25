import socket
import json

_IPC_SOCKET_PATH = "/var/jobber/0/cmd.sock"

class ipclib(object):
    ROBOT_LIBRARY_VERSION = 1.0

    def _do_ipc(self, method, params):
        # connect
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        sock.connect(_IPC_SOCKET_PATH)

        # send command
        cmd = {
            "method": method,
            "params": params,
            "id": 1
        }
        cmd_str = json.dumps(cmd)
        print("Sending IPC cmd: {0}".format(cmd_str))
        total_sent = 0
        while total_sent < len(cmd_str):
            n = sock.send(cmd_str[total_sent:])
            if n == 0:
                raise Exception("socket connection broken")
            total_sent += n

        # get response
        resp_str = sock.recv(10*1024)
        print("Got IPC response: {0}".format(resp_str))
        resp = json.loads(resp_str)
        if resp.get("error") is not None:
            raise Exception(resp["error"])
        return resp["result"]

    def do_set_job_cmd(self, job_name, cmd):
        cmd = {
            "job": {
                "name": job_name,
                "cmd": cmd,
                "time": "*"
            }
        }
        self._do_ipc("IpcService.SetJob", [cmd])

    def do_delete_job_cmd(self, job_name):
        cmd = {
            "job": job_name
        }
        self._do_ipc("IpcService.DeleteJob", [cmd])
