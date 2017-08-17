package main

import (
	"flag"
	"fmt"
	"github.com/dshearer/jobber/common"
	"net"
	"net/rpc"
	"os"
	"os/user"
	"syscall"
)

const (
	ListCmdStr   = "list"
	LogCmdStr    = "log"
	ReloadCmdStr = "reload"
	StopCmdStr   = "stop"
	TestCmdStr   = "test"
	CatCmdStr    = "cat"
	PauseCmdStr  = "pause"
	ResumeCmdStr = "resume"
)

var CmdStrs = [...]string{
	ListCmdStr,
	LogCmdStr,
	ReloadCmdStr,
	StopCmdStr,
	TestCmdStr,
	CatCmdStr,
	PauseCmdStr,
	ResumeCmdStr,
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

func subcmdUsage(subcmd string, flagSet *flag.FlagSet) func() {
	return func() {
		fmt.Printf("\nUsage: %v %v [flags]\nFlags:\n", os.Args[0], subcmd)
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
	} else {
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
		if _, err := os.Stat(common.RunnerSockPath(user)); os.IsNotExist(err) {
			if flag.Arg(0) == StopCmdStr {
				os.Exit(0)
			} else {
				fmt.Fprintf(os.Stderr, "jobberd isn't running.\n")
				os.Exit(1)
			}
		}

		// connect to daemon
		addr, err := net.ResolveUnixAddr("unix", common.RunnerSockPath(user))
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

		// drop privileges
		err = syscall.Setreuid(syscall.Getuid(), syscall.Getuid())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't drop privileges: %v\n", err)
			os.Exit(1)
		}

		// do command
		switch flag.Arg(0) {
		case ListCmdStr:
			doListCmd(flag.Args()[1:], rpcClient, user)

		case LogCmdStr:
			doLogCmd(flag.Args()[1:], rpcClient, user)

		case ReloadCmdStr:
			doReloadCmd(flag.Args()[1:], rpcClient, user)

		case StopCmdStr:
			var result string
			arg := common.IpcArg{User: user.Username, ForAllUsers: false}
			err = rpcClient.Call("RealIpcServer.Stop", arg, &result)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}

		case TestCmdStr:
			doTestCmd(flag.Args()[1:], rpcClient, user)

		case CatCmdStr:
			doCatCmd(flag.Args()[1:], rpcClient, user)

		case PauseCmdStr:
			doPauseCmd(flag.Args()[1:], rpcClient, user)

		case ResumeCmdStr:
			doResumeCmd(flag.Args()[1:], rpcClient, user)

		default:
			fmt.Fprintf(os.Stderr, "Invalid command: \"%v\".\n", flag.Arg(0))
			flag.Usage()
			os.Exit(1)
		}
	}
}

func doListCmd(args []string, rpcClient *rpc.Client, user *user.User) {
	// parse flags
	flagSet := flag.NewFlagSet(ListCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(ListCmdStr, flagSet)
	var help_p = flagSet.Bool("h", false, "help")
	var allUsers_p = flagSet.Bool("a", false, "all-users")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	} else {
		var result string
		arg := common.IpcArg{User: user.Username, ForAllUsers: *allUsers_p}
		err := rpcClient.Call("RealIpcServer.ListJobs", arg, &result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		// print result
		fmt.Printf("%s\n", result)
	}
}

func doLogCmd(args []string, rpcClient *rpc.Client, user *user.User) {
	// parse flags
	flagSet := flag.NewFlagSet(LogCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(LogCmdStr, flagSet)
	var help_p = flagSet.Bool("h", false, "help")
	var allUsers_p = flagSet.Bool("a", false, "all-users")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	} else {
		var result string
		arg := common.IpcArg{User: user.Username, ForAllUsers: *allUsers_p}
		err := rpcClient.Call("RealIpcServer.ListHistory", arg, &result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		// print result
		fmt.Printf("%s\n", result)
	}
}

func doReloadCmd(args []string, rpcClient *rpc.Client, user *user.User) {
	// parse flags
	flagSet := flag.NewFlagSet(ReloadCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(ReloadCmdStr, flagSet)
	var help_p = flagSet.Bool("h", false, "help")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	} else {
		var result common.ReloadCmdResp
		err := rpcClient.Call(
			"NewIpcService.Reload",
			common.ReloadNewCmd{},
			&result,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if result.Err == nil {
			fmt.Printf("Loaded %v jobs.\n", result.NumJobs)
		} else {
			fmt.Fprintf(os.Stderr, "%v", result.Err)
			os.Exit(1)
		}
	}
}

func doTestCmd(args []string, rpcClient *rpc.Client, user *user.User) {
	// parse flags
	flagSet := flag.NewFlagSet(TestCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(TestCmdStr, flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	var jobUser_p *string = flagSet.String("u", user.Username, "user")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	} else {
		// get job to test
		if len(flagSet.Args()) == 0 {
			fmt.Fprintf(os.Stderr, "You must specify a job.\n")
			os.Exit(1)
		}
		var job string = flagSet.Args()[0]

		// check "-u" opt
		if *jobUser_p == "" {
			fmt.Fprintf(os.Stderr, "Option requires an argument: \"-u\"\n")
			os.Exit(1)
		}

		var result string
		fmt.Printf("Running job \"%v\" for user \"%v\"...\n", job, *jobUser_p)
		arg := common.IpcArg{User: user.Username, Job: job, JobUser: *jobUser_p}
		err := rpcClient.Call("RealIpcServer.Test", arg, &result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%v\n", result)
	}
}

func doCatCmd(args []string, rpcClient *rpc.Client, user *user.User) {
	// parse flags
	flagSet := flag.NewFlagSet(CatCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(CatCmdStr, flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	var jobUser_p *string = flagSet.String("u", user.Username, "user")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	} else {
		// get job to cat
		if len(flagSet.Args()) == 0 {
			fmt.Fprintf(os.Stderr, "You must specify a job.\n")
			os.Exit(1)
		}
		var job string = flagSet.Args()[0]

		// check "-u" opt
		if *jobUser_p == "" {
			fmt.Fprintf(os.Stderr, "Option requires an argument: \"-u\"\n")
			os.Exit(1)
		}

		var result string
		arg := common.IpcArg{User: user.Username, Job: job, JobUser: *jobUser_p}
		err := rpcClient.Call("RealIpcServer.Cat", arg, &result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%v\n", result)
	}
}

func doPauseCmd(args []string, rpcClient *rpc.Client, user *user.User) {
	// parse flags
	flagSet := flag.NewFlagSet(PauseCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(PauseCmdStr, flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	} else {
		// get jobs
		var jobs []string = flagSet.Args()

		var result string
		arg := common.IpcArg{User: user.Username, Jobs: jobs}
		err := rpcClient.Call("RealIpcServer.Pause", arg, &result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%v\n", result)
	}
}

func doResumeCmd(args []string, rpcClient *rpc.Client, user *user.User) {
	// parse flags
	flagSet := flag.NewFlagSet(ResumeCmdStr, flag.ExitOnError)
	flagSet.Usage = subcmdUsage(ResumeCmdStr, flagSet)
	var help_p *bool = flagSet.Bool("h", false, "help")
	flagSet.Parse(args)

	if *help_p {
		flagSet.Usage()
		os.Exit(0)
	} else {
		// get jobs
		var jobs []string = flagSet.Args()

		var result string
		arg := common.IpcArg{User: user.Username, Jobs: jobs}
		err := rpcClient.Call("RealIpcServer.Resume", arg, &result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%v\n", result)
	}
}
