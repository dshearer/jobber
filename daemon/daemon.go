package main

import (
    "github.com/dshearer/jobber"
    "os"
    "os/signal"
    "fmt"
)

func stopServerOnSigint(server *jobber.IpcServer) {
    // Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	<-c
	server.Stop()
}

func main() {
    var err error
    
    // run them
	cmdChan := make(chan jobber.ICmd)
	manager := jobber.NewJobManager()
	err = manager.Launch(cmdChan)
	if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    // make IPC server
    ipcServer := jobber.NewIpcServer(cmdChan)
    go stopServerOnSigint(ipcServer)
    err = ipcServer.Launch()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    defer ipcServer.Stop()
    
    manager.Wait()
}
