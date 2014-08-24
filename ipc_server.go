package jobber

import (
    "net"
    "net/rpc"
)

type RealIpcServer struct {
    cmdChan chan ICmd
}

func (s *RealIpcServer) ListJobs(arg int, result *string) error {
    respChan := make(chan ICmdResp, 1)
    cmd := &ListJobsCmd{respChan}
    
    // send cmd
    s.cmdChan <- cmd
    
    // get resp
    var resp ICmdResp
    resp = <-respChan
    
    if resp.IsError() {
        errResp := resp.(*ErrorCmdResp)
        return errResp.Error
    } else {
        succResp := resp.(*SuccessCmdResp)
        *result = succResp.Details
        return nil
    }
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
    
    // make socket
    addr, err := net.ResolveUnixAddr("unix", DaemonSocketAddr)
    if err != nil {
        return err
    }
    s.listener, err = net.ListenUnix("unix", addr)
    if err != nil {
        return err
    }
    
    // make RPC server
    rpcServer := rpc.NewServer()
    rpcServer.Register(&s.realServer)
    go rpcServer.Accept(s.listener)
    
    return nil
}

func (s *IpcServer) Stop() {
    s.listener.Close()
}
