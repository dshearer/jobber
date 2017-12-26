package main

import (
	"github.com/dshearer/jobber/common"
	"net"
	"net/rpc"
	"os"
)

type NewIpcService struct {
	cmdChan chan<- common.ICmd
}

func (self *NewIpcService) Reload(
	cmd common.ReloadCmd,
	resp *common.ReloadCmdResp) error {

	// send command
	cmd.RespChan = make(chan *common.ReloadCmdResp, 1)
	self.cmdChan <- cmd

	// get response
	*resp = *<-cmd.RespChan
	return resp.Err
}

func (self *NewIpcService) ListJobs(
	cmd common.ListJobsCmd,
	resp *common.ListJobsCmdResp) error {

	// send command
	cmd.RespChan = make(chan *common.ListJobsCmdResp, 1)
	self.cmdChan <- cmd

	// get response
	*resp = *<-cmd.RespChan
	return resp.Err
}

func (self *NewIpcService) Log(
	cmd common.LogCmd,
	resp *common.LogCmdResp) error {

	// send command
	cmd.RespChan = make(chan *common.LogCmdResp, 1)
	self.cmdChan <- cmd

	// get response
	*resp = *<-cmd.RespChan
	return resp.Err
}

func (self *NewIpcService) Test(
	cmd common.TestCmd,
	resp *common.TestCmdResp) error {

	// send command
	cmd.RespChan = make(chan *common.TestCmdResp, 1)
	self.cmdChan <- cmd

	// get response
	*resp = *<-cmd.RespChan
	return resp.Err
}

func (self *NewIpcService) Cat(
	cmd common.CatCmd,
	resp *common.CatCmdResp) error {

	// send command
	cmd.RespChan = make(chan *common.CatCmdResp, 1)
	self.cmdChan <- cmd

	// get response
	*resp = *<-cmd.RespChan
	return resp.Err
}

func (self *NewIpcService) Pause(
	cmd common.PauseCmd,
	resp *common.PauseCmdResp) error {

	// send command
	cmd.RespChan = make(chan *common.PauseCmdResp, 1)
	self.cmdChan <- cmd

	// get response
	*resp = *<-cmd.RespChan
	return resp.Err
}

func (self *NewIpcService) Resume(
	cmd common.ResumeCmd,
	resp *common.ResumeCmdResp) error {

	// send command
	cmd.RespChan = make(chan *common.ResumeCmdResp, 1)
	self.cmdChan <- cmd

	// get response
	*resp = *<-cmd.RespChan
	return resp.Err
}

func (self *NewIpcService) Init(
	cmd common.InitCmd,
	resp *common.InitCmdResp) error {

	// send command
	cmd.RespChan = make(chan *common.InitCmdResp, 1)
	self.cmdChan <- cmd

	// get response
	*resp = *<-cmd.RespChan
	return resp.Err
}

type IpcServer struct {
	service  NewIpcService
	listener *net.UnixListener
	sockPath string
}

func NewIpcServer(sockPath string,
	cmdChan chan<- common.ICmd,
	respChan <-chan common.ICmdResp) *IpcServer {

	server := &IpcServer{sockPath: sockPath}
	server.service.cmdChan = cmdChan
	return server
}

func (self *IpcServer) Launch() error {
	var err error

	// make socket
	os.Remove(self.sockPath)
	addr, err := net.ResolveUnixAddr("unix", self.sockPath)
	if err != nil {
		return err
	}
	self.listener, err = net.ListenUnix("unix", addr)
	if err != nil {
		return err
	}

	// make RPC server
	rpcServer := rpc.NewServer()
	rpcServer.Register(&self.service)
	go rpcServer.Accept(self.listener)

	return nil
}

func (self *IpcServer) Stop() {
	self.listener.Close()
	os.Remove(self.sockPath)
}
