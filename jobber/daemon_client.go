package main

import (
	"fmt"
	"github.com/dshearer/jobber/common"
	"net"
	"net/rpc"
	"os"
	"os/user"
	"time"
)

const gTimeout = 5 * time.Second

const gNoSocketErrMsg = `Jobber doesn't seem to be running for user %v.
(No socket at %v.)`
const gTimeoutErrMsg = "Call to Jobber timed out."
const gCallErrMgs = "Call to Jobber failed: %v"

func CallDaemon(method string, args, reply interface{},
	usr *user.User, timeout bool) error {

	// make sure the daemon is running
	socketPath := common.CmdSocketPath(usr)
	_, err := os.Stat(socketPath)
	if os.IsNotExist(err) {
		msg := fmt.Sprintf(
			gNoSocketErrMsg,
			usr.Username,
			socketPath,
		)
		return &common.Error{What: msg, Cause: err}
	} else if err != nil {
		return err
	}

	// connect to daemon
	addr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return err
	}
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// make client
	client := rpc.NewClient(conn)
	defer client.Close()

	// make timeout timer
	timeoutTimer := time.NewTimer(gTimeout)
	if timeout {
		defer timeoutTimer.Stop()
	} else {
		timeoutTimer.Stop()
	}

	// send request
	call := client.Go(method, args, reply, nil)
	select {
	case <-call.Done:
		if call.Error != nil {
			msg := fmt.Sprintf(gCallErrMgs, call.Error)
			return &common.Error{What: msg, Cause: err}
		}
		return nil

	case <-timeoutTimer.C:
		return &common.Error{What: gTimeoutErrMsg}
	}
}
