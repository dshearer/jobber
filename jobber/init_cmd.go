package main

import (
	"flag"
	"fmt"
	"github.com/dshearer/jobber/common"
	"os"
	"os/user"
)

func doInitCmd(args []string) int {
	// parse flags
	flagSet := flag.NewFlagSet(InitCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(InitCmdStr, "", flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
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
	var resp common.InitCmdResp
	err = CallDaemon(
		"NewIpcService.Init",
		common.InitCmd{},
		&resp,
		usr,
		true,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// handle response
	if resp.Err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", resp.Err)
		return 1
	}
	fmt.Printf("You can now define jobs in %v.\n", resp.JobfilePath)
	fmt.Println("Once you have done so, run 'jobber reload'.")
	return 0
}
