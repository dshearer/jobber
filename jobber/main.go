package main

import (
	"flag"
	"fmt"
	"github.com/dshearer/jobber/common"
	"net"
	"net/rpc"
	"os"
	"os/user"
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

type CmdHandler func([]string) int

var CmdHandlers = map[string]CmdHandler{
	ListCmdStr:   doListCmd,
	LogCmdStr:    doLogCmd,
	ReloadCmdStr: doReloadCmd,
	TestCmdStr:   doTestCmd,
	CatCmdStr:    doCatCmd,
	PauseCmdStr:  doPauseCmd,
	ResumeCmdStr: doResumeCmd,
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

func connectToDaemon(socketPath string) (net.Conn, error) {
	// make sure the daemon is running
	_, err := os.Stat(socketPath)
	if os.IsNotExist(err) {
		msg := fmt.Sprintf(
			"jobberrunner isn't running (%v)",
			socketPath,
		)
		return nil, &common.Error{What: msg, Cause: err}
	} else if err != nil {
		return nil, err
	}

	// connect to daemon
	addr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return nil, err
	}
	return net.DialUnix("unix", nil, addr)
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
	os.Exit(handler(flag.Args()[1:]))
}

func doTestCmd(args []string) int {
	// parse flags
	flagSet := flag.NewFlagSet(TestCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(TestCmdStr, "JOB", flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	//	var jobUser_p *string = flagSet.String("u", user.Username, "user")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		return 0
	}

	// get job to test
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

	// connect to user's daemon
	daemonConn, err := connectToDaemon(common.CmdSocketPath(usr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	defer daemonConn.Close()
	daemonClient := rpc.NewClient(daemonConn)

	// send command
	var resp common.TestCmdResp
	fmt.Printf("Running job \"%v\"...\n", job)
	err = daemonClient.Call(
		"NewIpcService.Test",
		common.TestCmd{Job: job},
		&resp,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// handle response
	fmt.Printf("%v\n", resp.Result)
	return 0
}

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

	// connect to user's daemon
	daemonConn, err := connectToDaemon(common.CmdSocketPath(usr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	defer daemonConn.Close()
	daemonClient := rpc.NewClient(daemonConn)

	// send command
	var resp common.CatCmdResp
	err = daemonClient.Call(
		"NewIpcService.Cat",
		common.CatCmd{Job: job},
		&resp,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// handle response
	fmt.Printf("%v\n", resp.Result)
	return 0
}

func doPauseCmd(args []string) int {
	// parse flags
	flagSet := flag.NewFlagSet(PauseCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(PauseCmdStr, "[JOBS...]", flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		return 0
	}

	// get jobs
	var jobs []string = flagSet.Args()

	// get current user
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to get current user: %v\n", err,
		)
		return 1
	}

	// connect to user's daemon
	daemonConn, err := connectToDaemon(common.CmdSocketPath(usr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	defer daemonConn.Close()
	daemonClient := rpc.NewClient(daemonConn)

	// send command
	var resp common.PauseCmdResp
	err = daemonClient.Call(
		"NewIpcService.Pause",
		common.PauseCmd{Jobs: jobs},
		&resp,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// handle response
	fmt.Printf("Paused %v jobs.\n", resp.AmtPaused)
	return 0
}

func doResumeCmd(args []string) int {
	// parse flags
	flagSet := flag.NewFlagSet(ResumeCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(ResumeCmdStr, "[JOBS...]", flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		return 0
	}

	// get jobs
	var jobs []string = flagSet.Args()

	// get current user
	usr, err := user.Current()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to get current user: %v\n", err,
		)
		return 1
	}

	// connect to user's daemon
	daemonConn, err := connectToDaemon(common.CmdSocketPath(usr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	defer daemonConn.Close()
	daemonClient := rpc.NewClient(daemonConn)

	// send command
	var resp common.ResumeCmdResp
	err = daemonClient.Call(
		"NewIpcService.Resume",
		common.ResumeCmd{Jobs: jobs},
		&resp,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// handle response
	fmt.Printf("Resumed %v jobs.\n", resp.AmtResumed)
	return 0
}
