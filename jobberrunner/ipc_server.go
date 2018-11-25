package main

import (
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"

	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/ipc"
)

type IpcService struct {
	cmdChan chan<- CmdContainer
}

func (self *IpcService) Reload(
	cmd ipc.ReloadCmd,
	resp_p *ipc.ReloadCmdResp) error {

	// send command
	respChan := make(chan ipc.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(ipc.ReloadCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) ListJobs(
	cmd ipc.ListJobsCmd,
	resp_p *ipc.ListJobsCmdResp) error {

	// send command
	respChan := make(chan ipc.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(ipc.ListJobsCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) Log(
	cmd ipc.LogCmd,
	resp_p *ipc.LogCmdResp) error {

	// send command
	respChan := make(chan ipc.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(ipc.LogCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) Test(
	cmd ipc.TestCmd,
	resp_p *ipc.TestCmdResp) error {

	// send command
	respChan := make(chan ipc.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(ipc.TestCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) Cat(
	cmd ipc.CatCmd,
	resp_p *ipc.CatCmdResp) error {

	// send command
	respChan := make(chan ipc.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(ipc.CatCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) Pause(
	cmd ipc.PauseCmd,
	resp_p *ipc.PauseCmdResp) error {

	// send command
	respChan := make(chan ipc.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(ipc.PauseCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) Resume(
	cmd ipc.ResumeCmd,
	resp_p *ipc.ResumeCmdResp) error {

	// send command
	respChan := make(chan ipc.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(ipc.ResumeCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) Init(
	cmd ipc.InitCmd,
	resp_p *ipc.InitCmdResp) error {

	// send command
	respChan := make(chan ipc.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(ipc.InitCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) SetJob(
	cmd ipc.SetJobCmd,
	resp_p *ipc.SetJobCmdResp) error {

	// send command
	respChan := make(chan ipc.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(ipc.SetJobCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) DeleteJob(
	cmd ipc.DeleteJobCmd,
	resp_p *ipc.DeleteJobCmdResp) error {

	// send command
	respChan := make(chan ipc.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(ipc.DeleteJobCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

type IpcServer interface {
	Launch() error
	Stop()
}

func serve(listener net.Listener, rpcServer *rpc.Server) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			common.ErrLogger.Printf("%v", err)
			return
		}
		go func() {
			rpcServer.ServeCodec(jsonrpc.NewServerCodec(conn))
			conn.Close()
		}()
	}
}

type udsIpcServer struct {
	service  IpcService
	listener *net.UnixListener
	sockPath string
}

func NewUdsIpcServer(sockPath string, cmdChan chan<- CmdContainer) IpcServer {
	server := &udsIpcServer{sockPath: sockPath}
	server.service.cmdChan = cmdChan
	return server
}

func (self *udsIpcServer) Launch() error {
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
	rpcServer.HandleHTTP("/", "/debug")

	// serve connections
	go serve(self.listener, rpcServer)

	return nil
}

func (self *udsIpcServer) Stop() {
	self.listener.Close()
	os.Remove(self.sockPath)
}

type inetIcpServer struct {
	service  IpcService
	listener net.Listener
	port     uint
}

func NewInetIpcServer(port uint, cmdChan chan<- CmdContainer) IpcServer {
	server := &inetIcpServer{port: port}
	server.service.cmdChan = cmdChan
	return server
}

func (self *inetIcpServer) Launch() error {
	// make socket
	var err error
	self.listener, err = net.Listen("tcp", fmt.Sprintf(":%v", self.port))
	if err != nil {
		return err
	}

	// make RPC server
	rpcServer := rpc.NewServer()
	rpcServer.Register(&self.service)
	rpcServer.HandleHTTP("/", "/debug")

	// serve connections
	go serve(self.listener, rpcServer)

	return nil
}

func (self *inetIcpServer) Stop() {
	self.listener.Close()
}
