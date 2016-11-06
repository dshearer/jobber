package main

import (
	"github.com/dshearer/jobber/common"
	"io"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"syscall"
)

var g_err_logger, _ = syslog.NewLogger(syslog.LOG_ERR|syslog.LOG_CRON, 0)

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

func main() {
	var err error

	// make loggers
	infoSyslogWriter, _ := syslog.New(syslog.LOG_NOTICE|syslog.LOG_CRON, "")
	errSyslogWriter, _ := syslog.New(syslog.LOG_ERR|syslog.LOG_CRON, "")
	if infoSyslogWriter != nil {
		common.Logger = log.New(io.MultiWriter(infoSyslogWriter, os.Stdout), "", 0)
		common.ErrLogger = log.New(io.MultiWriter(errSyslogWriter, os.Stderr), "", 0)
	}

	// run them
	manager, err := NewJobManager()
	if err != nil {
		if g_err_logger != nil {
			g_err_logger.Printf("Error: %v\n", err)
		}
		os.Exit(1)
	}
	cmdChan, err := manager.Launch()
	if err != nil {
		if g_err_logger != nil {
			g_err_logger.Printf("Error: %v\n", err)
		}
		os.Exit(1)
	}

	// make IPC server
	ipcServer := NewIpcServer(cmdChan)
	go stopServerOnSignal(ipcServer, manager)
	err = ipcServer.Launch()
	if err != nil {
		if g_err_logger != nil {
			g_err_logger.Printf("Error: %v\n", err)
		}
		os.Exit(1)
	}
	defer ipcServer.Stop()

	manager.Wait()
}
