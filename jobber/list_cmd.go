package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/dshearer/jobber/common"
	"net/rpc"
	"os"
	"os/user"
	"strings"
	"text/tabwriter"
)

func sendListCmd(usr *user.User) (*common.ListJobsCmdResp, error) {
	// connect to user's daemon
	daemonConn, err := connectToDaemon(common.CmdSocketPath(usr))
	if err != nil {
		return nil, err
	}
	defer daemonConn.Close()
	daemonClient := rpc.NewClient(daemonConn)

	// send command
	var result common.ListJobsCmdResp
	err = daemonClient.Call(
		"NewIpcService.ListJobs",
		common.ListJobsCmd{},
		&result,
	)
	if err != nil {
		return nil, err
	}

	//	fmt.Printf("User: %v; socket: %v, num jobs: %v\n",
	//		usr,
	//		common.SocketPath(usr),
	//		len(result.Jobs),
	//	)

	return &result, nil
}

type ListRespRec struct {
	usr  *user.User
	resp *common.ListJobsCmdResp
}

func doListCmd_allUsers() int {
	// get all users
	users, err := common.AllUsersWithSockets()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to get all users: %v\n", err,
		)
		return 1
	}

	// send cmd
	var responses []ListRespRec
	for _, usr := range users {
		resp, err := sendListCmd(usr)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to list jobs for %v: %v\n", usr.Name, err)
			continue
		}
		rec := ListRespRec{usr: usr, resp: resp}
		responses = append(responses, rec)
	}

	// make table header
	var buffer bytes.Buffer
	var writer *tabwriter.Writer = tabwriter.NewWriter(&buffer,
		5, 0, 2, ' ', 0)
	headers := [...]string{
		"NAME",
		"STATUS",
		"SEC/MIN/HR/MDAY/MTH/WDAY",
		"NEXT RUN TIME",
		"NOTIFY ON ERR",
		"NOTIFY ON FAIL",
		"ERR HANDLER",
		"USER",
	}
	fmt.Fprintf(writer, "%v\n", strings.Join(headers[:], "\t"))

	// make table rows
	var rows []string
	for _, respRec := range responses {
		var userName string
		if len(respRec.usr.Name) > 0 {
			userName = respRec.usr.Name
		} else {
			userName = respRec.usr.Username
		}
		for _, j := range respRec.resp.Jobs {
			nextRunTime := "none"
			if j.NextRunTime != nil {
				nextRunTime =
					j.NextRunTime.Format("Jan _2 15:04:05 2006")
			}
			var s string
			s = fmt.Sprintf(
				"%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v",
				j.Name,
				j.Status,
				j.Schedule,
				nextRunTime,
				j.NotifyOnErr,
				j.NotifyOnFail,
				j.ErrHandler,
				userName)
			rows = append(rows, s)
		}
	}
	fmt.Fprintf(writer, "%v", strings.Join(rows, "\n"))
	writer.Flush()
	fmt.Printf("%v\n", buffer.String())
	return 0
}

func doListCmd_currUser() int {
	// get current user
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to get current user: %v\n", err,
		)
		return 1
	}

	// send cmd
	resp, err := sendListCmd(usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// make table header
	var buffer bytes.Buffer
	var writer *tabwriter.Writer = tabwriter.NewWriter(&buffer,
		5, 0, 2, ' ', 0)
	headers := [...]string{
		"NAME",
		"STATUS",
		"SEC/MIN/HR/MDAY/MTH/WDAY",
		"NEXT RUN TIME",
		"NOTIFY ON ERR",
		"NOTIFY ON FAIL",
		"ERR HANDLER",
	}
	fmt.Fprintf(writer, "%v\n", strings.Join(headers[:], "\t"))

	// handle response
	strs := make([]string, 0, len(resp.Jobs))
	for _, j := range resp.Jobs {
		nextRunTime := "none"
		if j.NextRunTime != nil {
			nextRunTime =
				j.NextRunTime.Format("Jan _2 15:04:05 2006")
		}
		var s string
		if usr != nil {
			s = fmt.Sprintf("%v\t", usr.Name)
		}
		s = fmt.Sprintf(
			"%v\t%v\t%v\t%v\t%v\t%v\t%v",
			j.Name,
			j.Status,
			j.Schedule,
			nextRunTime,
			j.NotifyOnErr,
			j.NotifyOnFail,
			j.ErrHandler)
		strs = append(strs, s)
	}
	fmt.Fprintf(writer, "%v", strings.Join(strs, "\n"))
	writer.Flush()
	fmt.Printf("%v\n", buffer.String())

	return 0
}

func doListCmd(args []string) int {
	// parse flags
	flagSet := flag.NewFlagSet(ListCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(ListCmdStr, "", flagSet)
	var help_p = flagSet.Bool("h", false, "help")
	var allUsers_p = flagSet.Bool("a", false, "all-users")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		return 0
	}

	if *allUsers_p {
		return doListCmd_allUsers()
	} else {
		return doListCmd_currUser()
	}
}
