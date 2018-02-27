package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"

	"github.com/dshearer/jobber/common"
)

func doReloadCmd(args []string) int {
	// parse flags
	flagSet := flag.NewFlagSet(ReloadCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(ReloadCmdStr, "", flagSet)
	var help_p = flagSet.Bool("h", false, "help")
	var allUsers_p = flagSet.Bool("a", false, "all-users")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		return 0
	}

	if *allUsers_p {
		// get all users
		users, err := common.AllUsersWithSockets()
		if err != nil {
			fmt.Fprintf(
				os.Stderr, "Failed to get all users: %v\n", err,
			)
			return 1
		}

		for _, usr := range users {
			// send cmd
			var resp common.ReloadCmdResp
			err = CallDaemon(
				"NewIpcService.Reload",
				common.ReloadCmd{},
				&resp,
				usr,
				true,
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return 1
			}

			// handle response
			fmt.Printf(
				"Loaded %v jobs for %v.\n", resp.NumJobs, usr.Name,
			)
		}
		return 0
	} else {
		// get current user
		usr, err := user.Current()
		if err != nil {
			fmt.Fprintf(
				os.Stderr, "Failed to get current user: %v\n", err,
			)
			return 1
		}

		// send cmd
		var resp common.ReloadCmdResp
		err = CallDaemon(
			"NewIpcService.Reload",
			common.ReloadCmd{},
			&resp,
			usr,
			true,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}

		// handle response
		fmt.Printf("Loaded %v jobs.\n", resp.NumJobs)
		return 0
	}
}
