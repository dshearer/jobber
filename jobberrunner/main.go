package main

import (
	"flag"
	"fmt"
	"github.com/dshearer/jobber/common"
	"io"
	"log"
	"log/syslog"
	"os"
	"os/signal"
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
	fmt.Printf("Usage: %v [flags] SOCKET_PATH JOBFILE_PATH\n\n", os.Args[0])

	fmt.Printf("Flags:\n")
	flag.PrintDefaults()
	fmt.Printf("\n")
}

func main() {
	// parse args
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

	if len(flag.Args()) != 2 {
		usage()
		os.Exit(1)
	}
	sockPath, jobfilePath := flag.Args()[0], flag.Args()[1]

	// make loggers
	infoSyslogWriter, _ := syslog.New(syslog.LOG_NOTICE|syslog.LOG_CRON, "")
	errSyslogWriter, _ := syslog.New(syslog.LOG_ERR|syslog.LOG_CRON, "")
	if infoSyslogWriter != nil {
		common.Logger = log.New(io.MultiWriter(infoSyslogWriter, os.Stdout), "", 0)
		common.ErrLogger = log.New(io.MultiWriter(errSyslogWriter, os.Stderr), "", 0)
	}

	// run job manager
	manager := NewJobManager(jobfilePath)
	if err := manager.Launch(); err != nil {
		common.ErrLogger.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// make IPC server
	ipcServer := NewIpcServer(
		sockPath,
		manager.CmdChan,
		manager.CmdRespChan)
	if err := ipcServer.Launch(); err != nil {
		common.ErrLogger.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	go stopServerOnSignal(ipcServer, manager)

	manager.Wait()
	ipcServer.Stop()
}
