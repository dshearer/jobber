package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/dshearer/jobber/common"
	"net/rpc"
	"os"
	"os/user"
	"sort"
	"strings"
	"text/tabwriter"
)

type EnhancedLogDesc struct {
	userName string
	logDesc  common.LogDesc
}

/* For sorting LogDescs: */
type EnhancedLogDescSorter []EnhancedLogDesc

/* For sorting EnhancedLogDescs: */
func (self EnhancedLogDescSorter) Len() int {
	return len(self)
}

/* For sorting EnhancedLogDescs: */
func (self EnhancedLogDescSorter) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

/* For sorting EnhancedLogDescs: */
func (self EnhancedLogDescSorter) Less(i, j int) bool {
	return self[i].logDesc.Time.After(self[j].logDesc.Time)
}

/* For sorting LogDescs: */
type LogDescSorter []common.LogDesc

/* For sorting LogDescs: */
func (self LogDescSorter) Len() int {
	return len(self)
}

/* For sorting LogDescs: */
func (self LogDescSorter) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

/* For sorting LogDescs: */
func (self LogDescSorter) Less(i, j int) bool {
	return self[i].Time.After(self[j].Time)
}

func sendLogCmd(usr *user.User) (*common.LogCmdResp, error) {
	// connect to user's daemon
	daemonConn, err := connectToDaemon(common.CmdSocketPath(usr))
	if err != nil {
		return nil, err
	}
	defer daemonConn.Close()
	daemonClient := rpc.NewClient(daemonConn)

	// send command
	var result common.LogCmdResp
	err = daemonClient.Call(
		"NewIpcService.Log",
		common.LogCmd{},
		&result,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func doLogCmd_allUsers() int {
	// get all users
	users, err := common.AllUsersWithSockets()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to get all users: %v\n", err,
		)
		return 1
	}

	// send cmd
	logDescs := make([]EnhancedLogDesc, 0)
	for _, usr := range users {
		resp, err := sendLogCmd(usr)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to get log for %v: %v.\n", usr.Name, err)
		}
		for _, log := range resp.Logs {
			logDescs = append(logDescs, EnhancedLogDesc{usr.Name, log})
		}
	}

	// handle response
	sort.Sort(EnhancedLogDescSorter(logDescs))
	var buffer bytes.Buffer
	var writer *tabwriter.Writer = tabwriter.NewWriter(&buffer, 5, 0,
		2, ' ', 0)
	fmt.Fprintf(writer, "TIME\tJOB\tSUCCEEDED\tRESULT\tUSER\n")
	strs := make([]string, 0)
	for _, e := range logDescs {
		s := fmt.Sprintf(
			"%v\t%v\t%v\t%v\t%v\t",
			e.logDesc.Time.Format("Jan _2 15:04:05 2006"),
			e.logDesc.Job,
			e.logDesc.Succeeded,
			e.logDesc.Result,
			e.userName)
		strs = append(strs, s)
	}
	fmt.Fprintf(writer, "%v", strings.Join(strs, "\n"))
	writer.Flush()
	fmt.Printf("%v\n", buffer.String())
	return 0
}

func doLogCmd_currUser() int {
	// get current user
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to get current user: %v\n", err,
		)
		return 1
	}

	// send command
	resp, err := sendLogCmd(usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// handle response
	sort.Sort(LogDescSorter(resp.Logs))
	var buffer bytes.Buffer
	var writer *tabwriter.Writer = tabwriter.NewWriter(&buffer, 5, 0,
		2, ' ', 0)
	fmt.Fprintf(writer, "TIME\tJOB\tSUCCEEDED\tRESULT\t\n")
	strs := make([]string, 0)
	for _, e := range resp.Logs {
		s := fmt.Sprintf(
			"%v\t%v\t%v\t%v\t",
			e.Time.Format("Jan _2 15:04:05 2006"),
			e.Job,
			e.Succeeded,
			e.Result)
		strs = append(strs, s)
	}
	fmt.Fprintf(writer, "%v", strings.Join(strs, "\n"))
	writer.Flush()
	fmt.Printf("%v\n", buffer.String())
	return 0
}

func doLogCmd(args []string) int {
	// parse flags
	flagSet := flag.NewFlagSet(LogCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(LogCmdStr, "", flagSet)
	var help_p = flagSet.Bool("h", false, "help")
	var allUsers_p = flagSet.Bool("a", false, "all-users")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		return 0
	}

	if *allUsers_p {
		return doLogCmd_allUsers()
	} else {
		return doLogCmd_currUser()
	}
}
