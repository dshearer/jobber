package main

import (
	"fmt"
	"net/rpc/jsonrpc"
	"os"
	"os/user"
	"time"

	"github.com/dshearer/jobber/common"
)

const gNoSocketErrMsg = `Jobber doesn't seem to be running for user %v.
(No socket at %v.)`
const gTimeoutErrMsg = "Call to Jobber timed out."
const gCallErrMgs = "Call to Jobber failed: %v"

func CallDaemon(method string, args, reply interface{},
	usr *user.User, timeout *time.Duration) error {

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

	// make client
	client, err := jsonrpc.Dial("unix", socketPath)
	if err != nil {
		return err
	}
	defer client.Close()

	var timeoutC <-chan time.Time
	if timeout != nil {
		// make timeout timer
		timeoutTimer := time.NewTimer(*timeout)
		timeoutC = timeoutTimer.C
		defer timeoutTimer.Stop()
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

	case <-timeoutC:
		return &common.Error{What: gTimeoutErrMsg}
	}
}
