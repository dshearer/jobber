package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/dshearer/jobber/common"
)

type RunnerProc struct {
	proc             *exec.Cmd
	quitSockPath     string
	quitSockListener net.Listener
	quitSockConn     net.Conn
	ExitedChan       <-chan error
}

type acceptResult struct {
	conn net.Conn
	err  error
}

func makeAcceptedChan(listener net.Listener) <-chan acceptResult {
	c := make(chan acceptResult, 1)
	go func() {
		conn, err := listener.Accept()
		c <- acceptResult{conn: conn, err: err}
		close(c)
	}()
	return c
}

func quotify(s string) string {
	return fmt.Sprintf(`"%v"`, s)
}

func LaunchRunner(usr *user.User,
	jobfilePath string) (*RunnerProc, error) {
	/*
		This thread is responsible for spawning an instance of
		jobberrunner under the privileges of a particular user, and
		restarting this instance as needed.

		In order to avoid a certain vulnerability
		(http://www.halfdog.net/Security/2012/TtyPushbackPrivilegeEscalation/),
		we must avoid the possibility that the jobbermaster process
		(running as root) shares a controlling terminal with any
		unprivileged child process, such as the jobberrunner process.
	*/

	/*
		Normally, the runner process will continue indefinitely.

		Before launching the runner process, we make a unix socket and
		connect to it as a server.  When the runner process gets going,
		it connects to this socket and then tries to read from the
		connection.  The runner process will continue as long as this
		read does not return, and so we can get the runner process to
		quit by closing our connection to the socket.
	*/

	var runnerProc = RunnerProc{
		quitSockPath: filepath.Join(common.PerUserDirPath(usr), "quit.sock"),
	}

	// look for jobberrunner
	runnerName := "jobberrunner"
	runnerPath := common.LibexecProgramPath(runnerName)

	// set umask
	oldUmask := syscall.Umask(0077)
	defer syscall.Umask(oldUmask)

	// make quit socket
	os.Remove(runnerProc.quitSockPath)
	addr, err := net.ResolveUnixAddr("unix", runnerProc.quitSockPath)
	if err != nil {
		return nil, err
	}
	runnerProc.quitSockListener, err = net.ListenUnix("unix", addr)
	if err != nil {
		return nil, err
	}
	if err = common.Chown(runnerProc.quitSockPath, usr); err != nil {
		return nil, err
	}

	// launch it
	cmdParts := []string{
		quotify(runnerPath),
		"-q", quotify(runnerProc.quitSockPath),
		"-u", quotify(common.CmdSocketPath(usr)),
		quotify(jobfilePath),
	}
	cmdParts = append(cmdParts, "-t", quotify(common.TempDirPath()))
	cmd := strings.Join(cmdParts, " ")
	runnerProc.proc = common.Sudo(*usr, cmd)
	// ensure we don't share TTY with the unprivileged process
	runnerProc.proc.Stdin = nil
	runnerProc.proc.Stdout = nil
	runnerProc.proc.Stderr = NewBoundedBuffer(1024)
	if err := runnerProc.proc.Start(); err != nil {
		return nil, err
	}
	runnerProc.ExitedChan = common.MakeCmdExitedChan(runnerProc.proc)

	// wait for it to connect to quit socket (or quit)
	acceptedChan := makeAcceptedChan(runnerProc.quitSockListener)
	select {
	case result := <-acceptedChan:
		if result.err != nil {
			runnerProc.Kill()
			return nil, result.err
		}
		runnerProc.quitSockConn = result.conn

	case <-runnerProc.ExitedChan:
		runnerProc.Kill()
		msg := fmt.Sprintf(
			"jobberrunner for %v exited prematurely.",
			usr.Username,
		)
		stderr := runnerProc.proc.Stderr.(*BoundedBuffer)
		common.ErrLogger.Printf("jobberrunner stderr:\n%v", stderr.String())
		return nil, &common.Error{What: msg}
	}

	return &runnerProc, nil
}

func (self *RunnerProc) Kill() {
	/*
	   The runner process actually consists of at least two processes:
	       1. An su process that spawns jobberrunner
	       2. The jobberrunner process

	   The first is our own child.  The second is a child of the first
	   and has its own process group.  For this reason, killing the
	   first may not kill the second.

	   jobberrunner, when it starts, writes its PID to a certain file.
	   So we can use that to kill the jobberrunner process directly.
	*/

	if self.quitSockConn != nil {
		self.quitSockConn.Close()
		self.quitSockConn = nil
	}
	if self.quitSockListener != nil {
		self.quitSockListener.Close()
		self.quitSockListener = nil
	}
	os.Remove(self.quitSockPath)
}

func (self *RunnerProc) Stderr() string {
	if self.proc == nil {
		return ""
	}
	stderr := self.proc.Stderr.(*BoundedBuffer)
	return stderr.String()
}
