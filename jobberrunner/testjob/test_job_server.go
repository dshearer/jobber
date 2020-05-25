package testjob

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
)

type TestJobServer struct {
	ctx       context.Context
	shell     string
	user      *user.User
	waitGroup sync.WaitGroup
}

func NewTestJobServer(ctx context.Context, shell string, user *user.User) *TestJobServer {
	return &TestJobServer{ctx: ctx, shell: shell, user: user}
}

func (self *TestJobServer) Wait() {
	self.waitGroup.Wait()
}

func (self *TestJobServer) Launch(job *jobfile.Job) (*string, error) {
	// set umask
	oldUmask := syscall.Umask(0077)
	defer syscall.Umask(oldUmask)

	// make socket path
	sockDir, err := ioutil.TempDir(common.PerUserDirPath(self.user), "try-*")
	if err != nil {
		return nil, err
	}
	sockPath := path.Join(sockDir, "output.sock")

	// make socket
	addr, err := net.ResolveUnixAddr("unix", sockPath)
	if err != nil {
		os.RemoveAll(sockDir)
		return nil, err
	}
	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		os.RemoveAll(sockDir)
		return nil, err
	}

	// launch thread
	self.waitGroup.Add(1)
	go self.serve(job, listener, sockDir)
	return &sockPath, nil
}

func (self *TestJobServer) serve(job *jobfile.Job, listener net.Listener, sockDir string) {
	defer os.RemoveAll(sockDir)
	defer listener.Close()
	defer self.waitGroup.Done()

	// wait for connection or timeout
	waitForConnCtx, waitForConnCtxCancel := context.WithTimeout(self.ctx, time.Minute)
	defer waitForConnCtxCancel()
	var conn net.Conn
	var ok bool
	select {
	case conn, ok = <-makeConnChan(listener):
		if !ok {
			/* listener.Accept returned error */
			return
		}
		defer conn.Close()

	case <-waitForConnCtx.Done():
		return
	}

	// launch thread to watch for cancellation (signalled by the client
	// closing the connection)
	subCtx, cancelSubCtx := context.WithCancel(self.ctx)
	defer cancelSubCtx()
	callWhenConnClosed(conn, cancelSubCtx)

	// run job
	thread := testJobThread{Stdout: conn, Stderr: conn}
	if err := thread.Run(subCtx, job, self.shell); err != nil {
		conn.Write([]byte(fmt.Sprintf("Failed to launch thread: %v\n", err)))
		return
	}

	// wait for job to finish or cancellation
	var result *jobfile.RunRec
	select {
	case result = <-thread.ResultChan():

	case <-subCtx.Done():
		common.Logger.Println("Job cancelled")
		// wait for subproc to finish
		result = <-thread.ResultChan()
	}

	// report result
	conn.Write([]byte("\n\n"))
	switch result.Fate {
	case common.SubprocFateSucceeded:
		conn.Write([]byte("Job succeeded\n"))
	case common.SubprocFateFailed:
		conn.Write([]byte("Job failed\n"))
	case common.SubprocFateCancelled:
		conn.Write([]byte("Job was cancelled\n"))
	}
}

func makeConnChan(listener net.Listener) <-chan net.Conn {
	connChan := make(chan net.Conn)
	go func() {
		defer close(connChan)
		conn, err := listener.Accept()
		if err != nil {
			common.ErrLogger.Printf("Accept error: %v\n", err)
			return
		}
		connChan <- conn
	}()
	return connChan
}

func callWhenConnClosed(conn net.Conn, cb func()) {
	go func() {
		var buf [1]byte
		for {
			_, err := conn.Read(buf[:])
			if err == io.EOF {
				cb()
				return
			}
		}
	}()
}
