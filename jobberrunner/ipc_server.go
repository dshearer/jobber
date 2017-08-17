package main

import (
	"github.com/dshearer/jobber/common"
	"net"
	"net/rpc"
	"os"
	"syscall"
)

type NewIpcService struct {
	cmdChan  chan<- common.ICmd
	respChan <-chan common.ICmdResp
}

func (self *NewIpcService) Reload(
	cmd common.ReloadCmd,
	resp *common.ReloadCmdResp) error {

	// send command
	self.cmdChan <- cmd

	// get response
	var tmp common.ICmdResp = <-self.respChan
	*resp = *tmp.(*common.ReloadCmdResp)
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
	server.service.respChan = respChan
	return server
}

func (self *IpcServer) Launch() error {
	var err error

	// set umask
	oldUmask := syscall.Umask(0177)

	// make socket
	os.Remove(self.sockPath)
	addr, err := net.ResolveUnixAddr("unix", self.sockPath)
	if err != nil {
		syscall.Umask(oldUmask)
		return err
	}
	self.listener, err = net.ListenUnix("unix", addr)
	if err != nil {
		syscall.Umask(oldUmask)
		return err
	}

	// restore umask
	syscall.Umask(oldUmask)

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
