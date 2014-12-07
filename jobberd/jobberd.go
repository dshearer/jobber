package main

import (
    "os"
    "os/signal"
    "syscall"
    "log/syslog"
)

var g_logger, _ = syslog.NewLogger(syslog.LOG_NOTICE | syslog.LOG_CRON, 0)
var g_err_logger, _ = syslog.NewLogger(syslog.LOG_ERR | syslog.LOG_CRON, 0)

func stopServerOnSignal(server *IpcServer, jm *JobManager) {
    // Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, os.Kill)

	// Block until a signal is received.
	<-c
	g_logger.Printf("Got signal.\n")
	server.Stop()
	jm.Stop()
	os.Exit(0)
}

func main() {
    var err error
    
    // run them
	manager, err := NewJobManager()
	if err != nil {
        g_err_logger.Printf("Error: %v\n", err)
        os.Exit(1)
    }
	cmdChan, err := manager.Launch()
	if err != nil {
        g_err_logger.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    
    // make IPC server
    ipcServer := NewIpcServer(cmdChan)
    go stopServerOnSignal(ipcServer, manager)
    err = ipcServer.Launch()
    if err != nil {
        g_err_logger.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    defer ipcServer.Stop()
    
    manager.Wait()
    g_logger.Printf("Job manager stopped.\n")
}
