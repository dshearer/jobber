package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/dshearer/jobber/ipc"
)

func doInitCmd(args []string) int {
	// parse flags
	flagSet := flag.NewFlagSet(InitCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(InitCmdStr, "", flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	var timeout_p = flagSet.Duration("t", 5 * time.Second, "timeout")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		return 0
	}

	// get current user
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to get current user: %v\n", err,
		)
		return 1
	}

	// send command
	var resp ipc.InitCmdResp
	err = CallDaemon(
		"IpcService.Init",
		ipc.InitCmd{},
		&resp,
		usr,
		timeout_p,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// handle response
	fmt.Printf("You can now define jobs in %v.\n", resp.JobfilePath)
	fmt.Println("Once you have done so, run 'jobber reload'.")
	return 0
}
