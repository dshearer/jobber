package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/dshearer/jobber/common"
	"net"
	"net/rpc"
	"os"
	"os/user"
	"sort"
	"strings"
	"text/tabwriter"
)

const (
	ListCmdStr   = "list"
	LogCmdStr    = "log"
	ReloadCmdStr = "reload"
	//	StopCmdStr   = "stop"
	TestCmdStr   = "test"
	CatCmdStr    = "cat"
	PauseCmdStr  = "pause"
	ResumeCmdStr = "resume"
)

var CmdStrs = [...]string{
	ListCmdStr,
	LogCmdStr,
	ReloadCmdStr,
	//	StopCmdStr,
	TestCmdStr,
	CatCmdStr,
	PauseCmdStr,
	ResumeCmdStr,
}

type CmdHandler func([]string, *rpc.Client)

var CmdHandlers = map[string]CmdHandler{
	ListCmdStr:   doListCmd,
	LogCmdStr:    doLogCmd,
	ReloadCmdStr: doReloadCmd,
	TestCmdStr:   doTestCmd,
	CatCmdStr:    doCatCmd,
	PauseCmdStr:  doPauseCmd,
	ResumeCmdStr: doResumeCmd,
}

/* For sorting RunLogEntries: */
type LogDescSorter []common.LogDesc

/* For sorting RunLogEntries: */
func (self LogDescSorter) Len() int {
	return len(self)
}

/* For sorting RunLogEntries: */
func (self LogDescSorter) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

/* For sorting RunLogEntries: */
func (self LogDescSorter) Less(i, j int) bool {
	return self[i].Time.After(self[j].Time)
}

func usage() {
	fmt.Printf("Usage: %v [flags] COMMAND\n\n", os.Args[0])

	fmt.Printf("Flags:\n")
	flag.PrintDefaults()
	fmt.Printf("\n")

	fmt.Printf("Commands:\n")
	for _, cmd := range CmdStrs {
		fmt.Printf("    %v\n", cmd)
	}
}

func subcmdUsage(subcmd string, posargs string, flagSet *flag.FlagSet) func() {
	return func() {
		fmt.Printf(
			"\nUsage: %v %v [flags] %v\nFlags:\n",
			os.Args[0],
			subcmd,
			posargs,
		)
		flagSet.PrintDefaults()
	}
}

func failIfNotRoot(user *user.User) {
	if user.Uid != "0" {
		fmt.Fprintf(os.Stderr, "You must be root.\n")
		os.Exit(1)
	}
}

func main() {
	flag.Usage = usage

	var helpFlag_p = flag.Bool("h", false, "help")
	var versionFlag_p = flag.Bool("v", false, "version")
	flag.Parse()

	if *helpFlag_p {
		usage()
		os.Exit(0)
	} else if *versionFlag_p {
		fmt.Printf("%v\n", common.LongVersionStr())
		os.Exit(0)
	}

	if len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "Please specify a command.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// get current user
	user, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current user: %v\n", err)
		os.Exit(1)
	}

	// make sure the daemon is running
	if _, err := os.Stat(common.SocketPath(user)); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "jobberrunner isn't running.\n")
		fmt.Fprintf(
			os.Stderr,
			"(No socket at %v)\n",
			common.SocketPath(user),
		)
		os.Exit(1)
	}

	// connect to daemon
	addr, err := net.ResolveUnixAddr("unix", common.SocketPath(user))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't resolve Unix addr: %v\n", err)
		os.Exit(1)
	}
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't connect to daemon: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	rpcClient := rpc.NewClient(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't make RPC client: %v\n", err)
		os.Exit(1)
	}

	// do command
	handler, ok := CmdHandlers[flag.Arg(0)]
	if !ok {
		fmt.Fprintf(
			os.Stderr,
			"Invalid command: \"%v\".\n\n",
			flag.Arg(0),
		)
		flag.Usage()
		os.Exit(1)
	}
	handler(flag.Args()[1:], rpcClient)
}

func doReloadCmd(args []string, rpcClient *rpc.Client) {
	// parse flags
	flagSet := flag.NewFlagSet(ReloadCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(ReloadCmdStr, "", flagSet)
	var help_p = flagSet.Bool("h", false, "help")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	}

	// send command
	var result common.ReloadCmdResp
	err := rpcClient.Call(
		"NewIpcService.Reload",
		common.ReloadCmd{},
		&result,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// handle response
	fmt.Printf("Loaded %v jobs.\n", result.NumJobs)
}

func doListCmd(args []string, rpcClient *rpc.Client) {
	// parse flags
	flagSet := flag.NewFlagSet(ListCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(ListCmdStr, "", flagSet)
	var help_p = flagSet.Bool("h", false, "help")
	//var allUsers_p = flagSet.Bool("a", false, "all-users")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	}

	// send command
	var resp common.ListJobsCmdResp
	err := rpcClient.Call(
		"NewIpcService.ListJobs",
		common.ListJobsCmd{},
		&resp,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// handle response
	var buffer bytes.Buffer
	var writer *tabwriter.Writer = tabwriter.NewWriter(&buffer, 5, 0, 2, ' ', 0)
	fmt.Fprintf(writer, "NAME\tSTATUS\tSEC/MIN/HR/MDAY/MTH/WDAY\tNEXT RUN TIME\tNOTIFY ON ERR\tNOTIFY ON FAIL\tERR HANDLER\n")
	strs := make([]string, 0, len(resp.Jobs))
	for _, j := range resp.Jobs {
		nextRunTime := "none"
		if j.NextRunTime != nil {
			nextRunTime = j.NextRunTime.Format("Jan _2 15:04:05 2006")
		}
		s := fmt.Sprintf(
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
}

func doLogCmd(args []string, rpcClient *rpc.Client) {
	// parse flags
	flagSet := flag.NewFlagSet(LogCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(LogCmdStr, "", flagSet)
	var help_p = flagSet.Bool("h", false, "help")
	//	var allUsers_p = flagSet.Bool("a", false, "all-users")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	}

	// send command
	var resp common.LogCmdResp
	err := rpcClient.Call(
		"NewIpcService.Log",
		common.LogCmd{},
		&resp,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// handle response
	sort.Sort(LogDescSorter(resp.Logs))
	var buffer bytes.Buffer
	var writer *tabwriter.Writer = tabwriter.NewWriter(&buffer, 5, 0, 2, ' ', 0)
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
}

func doTestCmd(args []string, rpcClient *rpc.Client) {
	// parse flags
	flagSet := flag.NewFlagSet(TestCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(TestCmdStr, "JOB", flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	//	var jobUser_p *string = flagSet.String("u", user.Username, "user")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	}

	// get job to test
	if len(flagSet.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "You must specify a job.\n")
		os.Exit(1)
	}
	var job string = flagSet.Args()[0]

	// send command
	var resp common.TestCmdResp
	fmt.Printf("Running job \"%v\"...\n", job)
	err := rpcClient.Call(
		"NewIpcService.Test",
		common.TestCmd{Job: job},
		&resp,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// handle response
	fmt.Printf("%v\n", resp.Result)
}

func doCatCmd(args []string, rpcClient *rpc.Client) {
	// parse flags
	flagSet := flag.NewFlagSet(CatCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(CatCmdStr, "JOB", flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	//	var jobUser_p *string = flagSet.String("u", user.Username, "user")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	}

	// get job to cat
	if len(flagSet.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "You must specify a job.\n")
		os.Exit(1)
	}
	var job string = flagSet.Args()[0]

	// send command
	var resp common.CatCmdResp
	err := rpcClient.Call(
		"NewIpcService.Cat",
		common.CatCmd{Job: job},
		&resp,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// handle response
	fmt.Printf("%v\n", resp.Result)
}

func doPauseCmd(args []string, rpcClient *rpc.Client) {
	// parse flags
	flagSet := flag.NewFlagSet(PauseCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(PauseCmdStr, "[JOBS...]", flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	}

	// get jobs
	var jobs []string = flagSet.Args()

	// send command
	var resp common.PauseCmdResp
	err := rpcClient.Call(
		"NewIpcService.Pause",
		common.PauseCmd{Jobs: jobs},
		&resp,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// handle response
	fmt.Printf("Paused %v jobs.\n", resp.AmtPaused)
}

func doResumeCmd(args []string, rpcClient *rpc.Client) {
	// parse flags
	flagSet := flag.NewFlagSet(ResumeCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(ResumeCmdStr, "[JOBS...]", flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	}

	// get jobs
	var jobs []string = flagSet.Args()

	var resp common.ResumeCmdResp
	err := rpcClient.Call(
		"NewIpcService.Resume",
		common.ResumeCmd{Jobs: jobs},
		&resp,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// handle response
	fmt.Printf("Resumed %v jobs.\n", resp.AmtResumed)
}
