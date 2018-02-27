package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dshearer/jobber/common"
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
	InitCmdStr   = "init"
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
	InitCmdStr,
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
	InitCmdStr:   doInitCmd,
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
