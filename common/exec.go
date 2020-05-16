package common

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
)

func cleanUpTempfile(f *os.File) {
	f.Close()
	os.Remove(f.Name())
}

// An ExecResult describes the result of running a subprocess via ExecAndWait.
// Stdout and Stderr can be used to get the subprocess's stdout and stderr.
// When done with this object (e.g., when done reading Stdout/Stderr), you must
// call Close.
type ExecResult struct {
	Stdout    io.ReadSeeker
	Stderr    io.ReadSeeker
	Succeeded bool
}

func (self *ExecResult) Close() {
	f := func(field *io.ReadSeeker) {
		if *field == nil {
			return
		}
		file := (*field).(*os.File)
		cleanUpTempfile(file)
		*field = nil
	}
	f(&self.Stdout)
	f(&self.Stderr)
}

func (self *ExecResult) _read(max int, f io.ReadSeeker) ([]byte, error) {
	buf := make([]byte, max)
	len, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf[:len], nil
}

func (self *ExecResult) ReadStdout(max int) (data []byte, err error) {
	return self._read(max, self.Stdout)
}

func (self *ExecResult) ReadStderr(max int) ([]byte, error) {
	return self._read(max, self.Stderr)
}

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
	return su_cmd(usr.Username, cmdStr, "/bin/sh")
}

func ExecAndWait(cmd *exec.Cmd, input []byte) (*ExecResult, error) {
	// make temp files for stdout/stderr
	stdout, err := ioutil.TempFile(TempDirPath(), "")
	if err != nil {
		return nil, err
	}
	stderr, err := ioutil.TempFile(TempDirPath(), "")
	if err != nil {
		cleanUpTempfile(stdout)
		return nil, err
	}

	// give them to cmd
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// make stdin pipe
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("Failed to get pipe to stdin: %v", err)
	}

	// start cmd
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("Failed to fork: %v", err)
	}

	// write input
	stdin.Write(input)
	stdin.Close()

	// finish execution
	waitErr := cmd.Wait()
	if waitErr != nil {
		ErrLogger.Printf("ExecAndWait: %v: %v", cmd.Path, waitErr)
		if _, ok := waitErr.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("Failed to fork: %v", waitErr)
		}
	}

	// seek in stdout/stderr
	if _, err := stdout.Seek(0, 0); err != nil {
		cleanUpTempfile(stdout)
		cleanUpTempfile(stderr)
		return nil, fmt.Errorf("Failed to seek stdout: %v", err)
	}
	if _, err := stderr.Seek(0, 0); err != nil {
		cleanUpTempfile(stdout)
		cleanUpTempfile(stderr)
		return nil, fmt.Errorf("Failed to seek stderr: %v", err)
	}

	// return result
	res := &ExecResult{}
	res.Stdout = stdout
	res.Stderr = stderr
	res.Succeeded = (waitErr == nil)
	return res, nil
}
