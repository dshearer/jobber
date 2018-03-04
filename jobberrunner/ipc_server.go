package main

import (
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"

	"github.com/dshearer/jobber/common"
)

type IpcService struct {
	cmdChan chan<- CmdContainer
}

func (self *IpcService) Reload(
	cmd common.ReloadCmd,
	resp_p *common.ReloadCmdResp) error {

	// send command
	respChan := make(chan common.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(common.ReloadCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) ListJobs(
	cmd common.ListJobsCmd,
	resp_p *common.ListJobsCmdResp) error {

	// send command
	respChan := make(chan common.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(common.ListJobsCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) Log(
	cmd common.LogCmd,
	resp_p *common.LogCmdResp) error {

	// send command
	respChan := make(chan common.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(common.LogCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) Test(
	cmd common.TestCmd,
	resp_p *common.TestCmdResp) error {

	// send command
	respChan := make(chan common.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(common.TestCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) Cat(
	cmd common.CatCmd,
	resp_p *common.CatCmdResp) error {

	// send command
	respChan := make(chan common.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(common.CatCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) Pause(
	cmd common.PauseCmd,
	resp_p *common.PauseCmdResp) error {

	// send command
	respChan := make(chan common.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(common.PauseCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) Resume(
	cmd common.ResumeCmd,
	resp_p *common.ResumeCmdResp) error {

	// send command
	respChan := make(chan common.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(common.ResumeCmdResp)
	if !ok {
		return &common.Error{What: "Unexpected response type"}
	}
	*resp_p = concreteResp
	return nil
}

func (self *IpcService) Init(
	cmd common.InitCmd,
	resp_p *common.InitCmdResp) error {

	// send command
	respChan := make(chan common.ICmdResp, 1)
	self.cmdChan <- CmdContainer{cmd, respChan}

	// get response
	resp := <-respChan
	if err := resp.Error(); err != nil {
		return err
	}
	concreteResp, ok := resp.(common.InitCmdResp)
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
			common.Logger.Printf("%v", err)
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
