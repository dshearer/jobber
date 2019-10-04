package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"os/user"
	"syscall"

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
	if gUser != nil {
		os.Remove(common.RunnerPidFilePath(gUser))
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

func recordPid(usr *user.User) {
	// write to temp file
	pidTmp, err := ioutil.TempFile(common.PerUserDirPath(usr), "temp-")
	if err != nil {
		common.ErrLogger.Printf("Failed to make temp file: %v", err)
		quit(1)
	}
	_, err = pidTmp.WriteString(fmt.Sprintf("%v", os.Getpid()))
	if err != nil {
		common.ErrLogger.Printf("Failed to write to %v: %v", pidTmp.Name(),
			err)
		pidTmp.Close()
		os.Remove(pidTmp.Name())
		quit(1)
	}
	pidTmp.Close()

	// rename it
	pidPath := common.RunnerPidFilePath(usr)
	if err = os.Rename(pidTmp.Name(), pidPath); err != nil {
		common.ErrLogger.Printf(
			"Failed to rename %v to %v: %v",
			pidTmp.Name(),
			pidPath,
			err,
		)
		os.Remove(pidTmp.Name())
		quit(1)
	}
}

func main() {
	// parse args
	flag.Usage = func() { usage("") }
	helpFlag_p := flag.Bool("h", false, "help")
	versionFlag_p := flag.Bool("v", false, "version")
	quickSockPath_p := flag.String("q", "", "quit socket path")
	udsSockPath_p := flag.String("u", "",
		"path to Unix socket on which to receive commands")
	inetPort_p := flag.Uint("p", 0,
		"path TCP socket on which to receive commands")
	standaloneFlag_p := flag.Bool("s", false, 
		"standalone mode - use jobberrunner without master process/ipc")
	flag.Parse()

	// handle generic version + help flags
	if *helpFlag_p {
		usage("")
		quit(0)
	} else if *versionFlag_p {
		fmt.Printf("%v\n", common.LongVersionStr())
		quit(0)
	}

	// get args
	if len(flag.Args()) != 1 {
		usage("")
		quit(1)
	}
	jobfilePath := flag.Args()[0]

	// sockets params provided ?
	haveUdsParam := udsSockPath_p != nil && len(*udsSockPath_p) > 0
	haveInetParam := inetPort_p != nil && *inetPort_p > 0

	// regular ipc mode ?
	if *standaloneFlag_p == false {
		// check for errors (tcp or unix socket param has to be set)
		if (!haveUdsParam && !haveInetParam) || (haveUdsParam && haveInetParam) {
			usage("Must specify exactly one of -u and -p")
			quit(1)
		}

		// get current user
		var err error
		gUser, err = user.Current()
		if err != nil {
			common.ErrLogger.Printf("Failed to get current user: %v", err)
			quit(1)
		}

		// record PID in file (for jobbermaster)
		recordPid(gUser)
	}

	// run job manager
	gJobManager = NewJobManager(jobfilePath)
	if err := gJobManager.Launch(); err != nil {
		common.ErrLogger.Printf("Error: %v\n", err)
		quit(1)
	}

	// regular ipc mode ?
	if *standaloneFlag_p == false {
		// make IPC server
		if haveUdsParam {
			gIpcServer = NewUdsIpcServer(*udsSockPath_p, gJobManager.CmdChan)
			common.Logger.Printf("Listening for commands on %v", *udsSockPath_p)
		} else {
			gIpcServer = NewInetIpcServer(*inetPort_p, gJobManager.CmdChan)
			common.Logger.Printf("Listening for commands on :%v", *inetPort_p)
		}
		if err := gIpcServer.Launch(); err != nil {
			common.ErrLogger.Printf("Error: %v", err)
			quit(1)
		}

		if quickSockPath_p != nil && len(*quickSockPath_p) > 0 {
			// listen for jobbermaster to tell us to quit
			go quitOnJobbermasterDiscon(*quickSockPath_p)
		}
	}

	// listen for signals
	go quitOnSignal()

	// wait for job manager
	gJobManager.Wait()

	quit(0)
}
