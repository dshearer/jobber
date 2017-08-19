package main

import (
	"fmt"
	"github.com/dshearer/jobber/common"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strconv"
)

type RunnerProc struct {
	proc       *exec.Cmd
	usr        *user.User
	ExitedChan <-chan error
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

	var runnerProc = RunnerProc{usr: usr}

	// look for jobberrunner
	runnerName := "jobberrunner"
	runnerPath, err := common.FindLibexecProgram(runnerName)
	if err != nil {
		return nil, err
	}

	// launch it
	cmd := fmt.Sprintf("%v \"%v\"", runnerPath, jobfilePath)
	runnerProc.proc = common.Sudo(*usr, cmd)
	// ensure we don't share TTY with the unprivileged process
	runnerProc.proc.Stdin = nil
	runnerProc.proc.Stdout = nil
	runnerProc.proc.Stderr = nil
	if err := runnerProc.proc.Start(); err != nil {
		return nil, err
	}

	// launch thread that waits for subproc
	exitedChan := make(chan error)
	go func() {
		exitedChan <- runnerProc.proc.Wait()
		close(exitedChan)
	}()
	runnerProc.ExitedChan = exitedChan

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

	// kill the su process
	self.proc.Process.Signal(os.Kill)
	<-self.ExitedChan

	pidPath := common.RunnerPidFilePath(self.usr)
	if pidF, err := os.Open(pidPath); err == nil {
		// kill the jobberrunner process
		defer pidF.Close()
		defer os.Remove(pidPath)
		b, err := ioutil.ReadAll(pidF)
		if err != nil {
			return
		}
		pid, err := strconv.Atoi(string(b))
		if err != nil {
			return
		}
		proc, err := os.FindProcess(pid)
		if err != nil {
			return
		}
		proc.Signal(os.Kill)
	}
}
