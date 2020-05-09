package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"os/user"
	"syscall"

	arg "github.com/alexflint/go-arg"
	"github.com/dshearer/jobber/common"
)

var gUser *user.User
var gIpcServer IpcServer
var gJobManager *JobManager

func quit(exitCode int) {
	if gIpcServer != nil {
		gIpcServer.Stop()
	}
	if gJobManager != nil {
		gJobManager.Cancel()
		gJobManager.Wait()
	}
	os.Exit(exitCode)
}

func quitOnSignal() {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, os.Kill)

	// Block until a signal is received.
	<-c
	quit(0)
}

func quitOnJobbermasterDiscon(quickSockPath string) {
	/*
	   Jobbermaster launched us and gave us a path to a Unix socket.
	   When jobbermaster wants us to quit, it will close that socket.
	*/

	common.Logger.Printf(
		"jobbermaster quit socket: %v",
		quickSockPath,
	)

	// open socket
	addr, err := net.ResolveUnixAddr("unix", quickSockPath)
	if err != nil {
		common.ErrLogger.Printf(
			"ResolveUnixAddr failed on %v: %v",
			quickSockPath,
			err,
		)
		quit(1)
		return
	}
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		common.ErrLogger.Printf(
			"DialUnix failed on %v: %v",
			quickSockPath,
			err,
		)
		quit(1)
		return
	}
	defer conn.Close()

	// read from it -- this will block until jobbermaster close the cxn
	var buf [1]byte
	_, _ = conn.Read(buf[:])

	/*
	   As conn.Read has returned, jobbermaster must have closed the cxn.
	*/
	quit(0)
}

func usage(extraMsg string) {
	if len(extraMsg) > 0 {
		fmt.Printf("%v\n\n", extraMsg)
	}

	fmt.Printf("Usage: %v [flags] JOBFILE_PATH\n\n", os.Args[0])

	fmt.Printf("Flags:\n")
	flag.PrintDefaults()
	fmt.Printf("\n")
}

type argsS struct {
	QuitSocket  *string `arg:"-q" help:"path to quit socket (used by jobbermaster to tell us to quit)"`
	UnixSocket  *string `arg:"-u" help:"path to Unix socket on which to receive commands"`
	TcpPort     *uint   `arg:"-p" help:"TCP port on which to receive commands"`
	TempDir     *string `arg:"-t" help:"Path to dir to use as temp dir"`
	JobfilePath string  `arg:"positional,required"`
	Debug       bool    `arg:"-d" default:"false"`
}

func (argsS) Version() string {
	return common.LongVersionStr()
}

func main() {

	// parse args
	var args argsS
	arg.MustParse(&args)

	// check for errors
	if args.UnixSocket != nil && args.TcpPort != nil {
		fmt.Fprintf(os.Stderr, "Must specify at most one of --unixsocket or --tcpport\n")
		quit(1)
	}
	if args.UnixSocket != nil && len(*args.UnixSocket) == 0 {
		fmt.Fprintf(os.Stderr, "Unix socket path cannot be empty\n")
		quit(1)
	}
	if args.QuitSocket != nil && len(*args.QuitSocket) == 0 {
		fmt.Fprintf(os.Stderr, "Quit socket path cannot be empty\n")
		quit(1)
	}
	if args.TempDir != nil && len(*args.TempDir) == 0 {
		fmt.Fprintf(os.Stderr, "Temp dir path cannot be empty\n")
		quit(1)
	}
	if args.TcpPort != nil && *args.TcpPort == 0 {
		fmt.Fprintf(os.Stderr, "TCP port cannot be zero\n")
		quit(1)
	}
	if len(args.JobfilePath) == 0 {
		fmt.Fprintf(os.Stderr, "Jobfile path cannot be empty")
		quit(1)
	}

	// init settings
	err := common.InitSettings(common.InitSettingsParams{
		TempDir: args.TempDir,
	})
	if err != nil {
		common.ErrLogger.Print(err)
		os.Exit(1)
	}

	if args.Debug {
		common.PrintPaths()
		os.Exit(0)
	}

	// get current user
	gUser, err = user.Current()
	if err != nil {
		common.ErrLogger.Printf("Failed to get current user: %v", err)
		quit(1)
	}

	// run job manager
	gJobManager = NewJobManager(args.JobfilePath)
	if err := gJobManager.Launch(); err != nil {
		common.ErrLogger.Printf("Error: %v\n", err)
		quit(1)
	}

	// make IPC server
	if args.UnixSocket != nil {
		gIpcServer = NewUdsIpcServer(*args.UnixSocket, gJobManager.CmdChan)
		common.Logger.Printf("Listening for commands on %v", *args.UnixSocket)
	} else if args.TcpPort != nil {
		gIpcServer = NewInetIpcServer(*args.TcpPort, gJobManager.CmdChan)
		common.Logger.Printf("Listening for commands on :%v", *args.TcpPort)
	}
	if gIpcServer != nil {
		if err := gIpcServer.Launch(); err != nil {
			common.ErrLogger.Printf("Error: %v", err)
			quit(1)
		}
	}

	if args.QuitSocket != nil {
		// listen for jobbermaster to tell us to quit
		go quitOnJobbermasterDiscon(*args.QuitSocket)
	}

	// listen for signals
	go quitOnSignal()

	// wait for job manager
	gJobManager.Wait()

	quit(0)
}
