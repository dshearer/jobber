package jobber

import (
    "net"
    "net/rpc"
    "syscall"
    "os"
    "os/user"
    "strconv"
)

type RealIpcServer struct {
    cmdChan chan ICmd
}

func (s *RealIpcServer) doCmd(cmd ICmd, result *string) error {
    // send cmd
    s.cmdChan <- cmd
    
    // get resp
    var resp ICmdResp
    resp = <-cmd.RespChan()
    
    if resp.IsError() {
        errResp := resp.(*ErrorCmdResp)
        return errResp.Error
    } else {
        succResp := resp.(*SuccessCmdResp)
        *result = succResp.Details
        return nil
    }
}

func (s *RealIpcServer) Reload(user string, result *string) error {
    return s.doCmd(&ReloadCmd{user, make(chan ICmdResp, 1)}, result)
}

func (s *RealIpcServer) ListJobs(user string, result *string) error {
    return s.doCmd(&ListJobsCmd{user, make(chan ICmdResp, 1)}, result)
}

func (s *RealIpcServer) ListHistory(user string, result *string) error {
    return s.doCmd(&ListHistoryCmd{user, make(chan ICmdResp, 1)}, result)
}

func (s *RealIpcServer) Stop(user string, result *string) error {
    return s.doCmd(&StopCmd{user, make(chan ICmdResp, 1)}, result)
}

type IpcServer struct {
    realServer RealIpcServer
    listener *net.UnixListener
}

func NewIpcServer(cmdChan chan ICmd) *IpcServer {
    server := &IpcServer{}
    server.realServer.cmdChan = cmdChan
    return server
}

func (s *IpcServer) Launch() error {
    var err error
    
    // set umask
    oldUmask := syscall.Umask(0177)
    
    // make socket
    addr, err := net.ResolveUnixAddr("unix", DaemonSocketAddr)
    if err != nil {
        syscall.Umask(oldUmask)
        return err
    }
    s.listener, err = net.ListenUnix("unix", addr)
    if err != nil {
        syscall.Umask(oldUmask)
        return err
    }
    
    // restore umask
    syscall.Umask(oldUmask)
    
    // change socket's owner
    jobberUser, err := user.Lookup("jobber_client")
    if err != nil {
        return err
    }
    uid, err := strconv.Atoi(jobberUser.Uid)
    if err != nil {
        return err
    }
    os.Chown(DaemonSocketAddr, uid, 0)
    
    // make RPC server
    rpcServer := rpc.NewServer()
    rpcServer.Register(&s.realServer)
    go rpcServer.Accept(s.listener)
    
    return nil
}

func (s *IpcServer) Stop() {
    s.listener.Close()
}
