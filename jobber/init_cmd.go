package main

import (
	"flag"
	"fmt"
	"github.com/dshearer/jobber/common"
	"os"
	"os/user"
	"strings"
)

const gDefaultJobfile = `## This is your jobfile: use it to tell Jobber what jobs you want it to
## run on your behalf.  For details of what you can specify here,
## please see https://dshearer.github.io/jobber/doc/.
##
## It consists of two sections: "prefs" and "jobs".  In "prefs" you can
## set various general settings.  In "jobs", you define your jobs.

[prefs]
## The following line makes jobber run a specified program when a job
## fails/succeeds:
#notifyProgram: /home/handleError.sh

## You can specify how info about past runs is stored.  For
## "type: memory" (the default), they are stored in memory and
## are lost when the Jobber service stops.
#runLog:
#    type: memory
#    maxLen: 100  # the max number of entries to remember

## For "type: file", past run logs are stored on disk.  The log file is
## rotated when it reaches a size of 'maxFileLen' MB.  Up to
## 'maxHistories' historical run logs (that is, not including the
## current one) are kept.
#runLog:
#    type: file
#    path: /tmp/claudius
#    maxFileLen: 50m  # in MB
#    maxHistories: 5

[jobs]
## This section must contain a YAML sequence of maps like the following:
#- name: DailyBackup
#  cmd: backup daily  # shell command to execute
#  time: '* * * * * *'  # SEC MIN HOUR MONTH_DAY MONTH WEEK_DAY.
#  onError: Continue  # what to do when the job has an error: Stop, Backoff, or Continue
#  notifyOnError: false  # whether to call notifyProgram when the job has an error
#  notifyOnFailure: true  # whether to call notifyProgram when the job stops due to errors
#  notifyOnSuccess: false  # whether to call notifyProgram when the job succeeds
`

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

	// open file for writing
	path, err := common.JobfilePath(usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return 1
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		if strings.Contains(err.Error(), "file exists") {
			fmt.Fprintf(
				os.Stderr, "Jobfile already exists at %v\n", path,
			)
		} else {
			fmt.Fprintf(
				os.Stderr, "Failed to open %v for writing: %v\n",
				path, err,
			)
		}
		return 1
	}
	defer f.Close()

	// write default jobfile
	fmt.Printf("Writing jobfile at %v\n", path)
	_, err = f.WriteString(gDefaultJobfile)
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to write to %v: %v\n", path, err,
		)
		return 1
	}

	fmt.Printf("\nYou can now define jobs in %v.\n", path)
	fmt.Println("Once you have, run 'jobber reload'.")

	return 0
}
