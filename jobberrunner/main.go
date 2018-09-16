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
var gIpcServer *IpcServer
var gJobManager *JobManager

func quit(exitCode int) {
	common.Logger.Printf("Jobberrunner Quitting")
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

func quitOnJobbermasterDiscon() {
	/*
	   Jobbermaster launched us and gave us a path to a Unix socket.
	   When jobbermaster wants us to quit, it will close that socket.
	*/

	common.Logger.Printf(
		"jobbermaster quit socket: %v",
		common.QuitSocketPath(gUser),
	)

	// open socket
	addr, err := net.ResolveUnixAddr("unix",
		common.QuitSocketPath(gUser))
	if err != nil {
		common.ErrLogger.Printf(
			"ResolveUnixAddr failed on %v: %v",
			common.QuitSocketPath(gUser),
			err,
		)
		quit(1)
		return
	}
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		common.ErrLogger.Printf(
			"DialUnix failed on %v: %v",
			common.QuitSocketPath(gUser),
			err,
		)
		quit(1)
		return
	}
	defer conn.Close()

	// read from it -- this will block until jobbermaster close the cxn
	var buf [1]byte
	_, err = conn.Read(buf[:])
	common.Logger.Printf("Finished reading from quit socket")
	if err != nil {
		common.Logger.Printf("read: %v", err)
	}

	/*
	   As conn.Read has returned, jobbermaster must have closed the cxn.
	*/
	quit(0)
}

func usage() {
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
	flag.Usage = usage
	var helpFlag_p = flag.Bool("h", false, "help")
	var versionFlag_p = flag.Bool("v", false, "version")
	var childFlag_p = flag.Bool("c",
		false,
		"run as child of jobbermaster",
	)
	flag.Parse()

	// handle flags
	if *helpFlag_p {
		usage()
		quit(0)
	} else if *versionFlag_p {
		fmt.Printf("%v\n", common.LongVersionStr())
		quit(0)
	}

	// get args
	if len(flag.Args()) != 1 {
		usage()
		quit(1)
	}
	jobfilePath := flag.Args()[0]

	// get current user
	var err error
	gUser, err = user.Current()
	if err != nil {
		common.ErrLogger.Printf("Failed to get current user: %v", err)
		quit(1)
	}

	// record PID in file (for jobbermaster)
	recordPid(gUser)

	// run job manager
	gJobManager = NewJobManager(jobfilePath)
	if err := gJobManager.Launch(); err != nil {
		common.ErrLogger.Printf("Error: %v\n", err)
		quit(1)
	}

	// make IPC server
	gIpcServer = NewIpcServer(
		common.CmdSocketPath(gUser),
		gJobManager.CmdChan,
		gJobManager.CmdRespChan,
	)
	if err := gIpcServer.Launch(); err != nil {
		common.ErrLogger.Printf("Error: %v", err)
		quit(1)
	}
	common.Logger.Printf(
		"Listening for commands on %v",
		common.CmdSocketPath(gUser),
	)

	if *childFlag_p {
		// listen for jobbermaster to tell us to quit
		go quitOnJobbermasterDiscon()
	}

	// listen for signals
	go quitOnSignal()

	// wait for job manager
	gJobManager.Wait()

	quit(0)
}
