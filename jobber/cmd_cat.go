package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"

	"github.com/dshearer/jobber/ipc"
)

func doCatCmd(args []string) int {
	// parse flags
	flagSet := flag.NewFlagSet(CatCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(CatCmdStr, "JOB", flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	//	var jobUser_p *string = flagSet.String("u", user.Username, "user")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		return 0
	}

	// get job to cat
	if len(flagSet.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "You must specify a job.\n")
		return 1
	}
	var job string = flagSet.Args()[0]

	// get current user
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to get current user: %v\n", err,
		)
		return 1
	}

	// send command
	var resp ipc.CatCmdResp
	err = CallDaemon(
		"IpcService.Cat",
		ipc.CatCmd{Job: job},
		&resp,
		usr,
		true,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// handle response
	fmt.Printf("%v\n", resp.Result)
	return 0
}
