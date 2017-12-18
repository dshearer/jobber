package main

import (
	"fmt"
	"github.com/dshearer/jobber/common"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"syscall"
)

type RunnerProc struct {
	proc             *exec.Cmd
	usr              *user.User
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

	var runnerProc = RunnerProc{usr: usr}

	// look for jobberrunner
	runnerName := "jobberrunner"
	runnerPath, err := common.FindLibexecProgram(runnerName)
	if err != nil {
		return nil, err
	}

	// open log file
	logFilePath := filepath.Join(usr.HomeDir, ".jobber-log")
	logF, err := os.OpenFile(
		logFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		common.ErrLogger.Printf(
			"Failed to make/open log file %v",
			logFilePath,
		)
		logF = nil
	} else {
		defer logF.Close()
	}

	// set umask
	oldUmask := syscall.Umask(0077)
	defer syscall.Umask(oldUmask)

	// make quit socket
	os.Remove(common.QuitSocketPath(usr))
	addr, err := net.ResolveUnixAddr("unix", common.QuitSocketPath(usr))
	if err != nil {
		return nil, err
	}
	runnerProc.quitSockListener, err = net.ListenUnix("unix", addr)
	if err != nil {
		return nil, err
	}
	if err = common.Chown(common.QuitSocketPath(usr), usr); err != nil {
		return nil, err
	}

	// launch it
	cmd := fmt.Sprintf("%v -c \"%v\"", runnerPath, jobfilePath)
	runnerProc.proc = common.Sudo(*usr, cmd)
	// ensure we don't share TTY with the unprivileged process
	runnerProc.proc.Stdin = nil
	runnerProc.proc.Stdout = logF
	runnerProc.proc.Stderr = logF
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
		common.Logger.Printf("jobberrunner for %v has started.",
			usr.Username)
		runnerProc.quitSockConn = result.conn

	case <-runnerProc.ExitedChan:
		runnerProc.Kill()
		msg := fmt.Sprintf(
			"jobberrunner for %v exited prematurely.",
			usr.Username,
		)
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
	os.Remove(common.QuitSocketPath(self.usr))
}
