package main

import (
	"flag"
	"fmt"
	"github.com/dshearer/jobber/common"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"syscall"
)

func stopServerOnSignal(server *IpcServer, jm *JobManager) {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, os.Kill)

	// Block until a signal is received.
	<-c
	server.Stop()
	jm.Cancel()
	jm.Wait()
	os.Exit(0)
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
		os.Exit(1)
	}
	_, err = pidTmp.WriteString(fmt.Sprintf("%v", os.Getpid()))
	if err != nil {
		common.ErrLogger.Printf("Failed to write to %v: %v", pidTmp.Name(),
			err)
		pidTmp.Close()
		os.Remove(pidTmp.Name())
		os.Exit(1)
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
		os.Exit(1)
	}
}

func daemonMain(usr *user.User) int {
	/*
	   IMPORTANT: Do not use os.Exit in here (or any called functions).
	   There's cleanup to do in main.
	*/

	// get args
	if len(flag.Args()) != 1 {
		usage()
		return 1
	}
	jobfilePath := flag.Args()[0]

	// run job manager
	manager := NewJobManager(jobfilePath)
	if err := manager.Launch(); err != nil {
		common.ErrLogger.Printf("Error: %v\n", err)
		return 1
	}

	// make IPC server
	ipcServer := NewIpcServer(
		common.SocketPath(usr),
		manager.CmdChan,
		manager.CmdRespChan)
	if err := ipcServer.Launch(); err != nil {
		common.ErrLogger.Printf("Error: %v", err)
		return 1
	}
	common.Logger.Printf("Listening on %v", common.SocketPath(usr))

	go stopServerOnSignal(ipcServer, manager)

	manager.Wait()
	ipcServer.Stop()
	return 0
}

func main() {

	// parse args
	flag.Usage = usage
	var helpFlag_p = flag.Bool("h", false, "help")
	var versionFlag_p = flag.Bool("v", false, "version")
	flag.Parse()

	// handle flags
	if *helpFlag_p {
		usage()
		os.Exit(0)
	} else if *versionFlag_p {
		fmt.Printf("%v\n", common.LongVersionStr())
		os.Exit(0)
	}

	// set umask
	syscall.Umask(0177)

	// get current user
	usr, err := user.Current()
	if err != nil {
		common.ErrLogger.Printf("Failed to get current user: %v", err)
		os.Exit(1)
	}

	// record PID in file (for jobbermaster)
	recordPid(usr)

	// run main logic
	retcode := daemonMain(usr)

	// remove PID file
	os.Remove(common.RunnerPidFilePath(usr))

	os.Exit(retcode)
}
