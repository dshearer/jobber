package common

import (
	"io/ioutil"
	"os/exec"
	"os/user"
)

type ExecResult struct {
	Stdout    []byte
	Stderr    []byte
	Succeeded bool
}

/*func Sudo(usr user.User, cmd string, args ...string) *exec.Cmd {
	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		panic("Invalid user ID")
	}
	gid, err := strconv.Atoi(usr.Gid)
	if err != nil {
		panic("Invalid group ID")
	}

	proc := exec.Command(program, args...)
	proc.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}
	return proc
}*/

func MakeCmdExitedChan(cmd *exec.Cmd) <-chan error {
	c := make(chan error, 1)
	go func() {
		c <- cmd.Wait()
		close(c)
	}()
	return c
}

/*
Returns an unstarted process descriptor.
*/
func Sudo(usr user.User, cmdStr string) *exec.Cmd {
	return sudo_cmd(usr.Username, cmdStr, "/bin/sh")
}

func ExecAndWait(cmd *exec.Cmd, input *[]byte) (*ExecResult, *Error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, &Error{"Failed to get pipe to stdout.", err}
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, &Error{"Failed to get pipe to stderr.", err}
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, &Error{"Failed to get pipe to stdin.", err}
	}

	// start cmd
	if err = cmd.Start(); err != nil {
		return nil, &Error{"Failed to execute command.", err}
	}

	if input != nil {
		// write input
		stdin.Write(*input)
	}
	stdin.Close()

	// read output
	stdoutBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, &Error{"Failed to read stdout.", err}
	}
	stderrBytes, err := ioutil.ReadAll(stderr)
	if err != nil {
		return nil, &Error{"Failed to read stderr.", err}
	}

	// finish execution
	waitErr := cmd.Wait()
	if waitErr != nil {
		ErrLogger.Printf("ExecAndWait: %v: %v", cmd.Path, waitErr)
		_, flag := waitErr.(*exec.ExitError)
		if !flag {
			return nil, &Error{"Failed to execute command.", waitErr}
		}
	}

	// return result
	res := &ExecResult{}
	res.Stdout = stdoutBytes
	res.Stderr = stderrBytes
	res.Succeeded = (waitErr == nil)
	return res, nil
}
