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
    
    // read jobs
    f, err := os.Open("/home/dylan/go_workspace/src/github.com/dshearer/jobber/example.json")
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
    
    // run them
	cmdChan := make(chan jobber.ICmd)
	manager := jobber.JobManager{Shell: "/bin/sh"}
	manager.LoadJobs(f)
	manager.Launch(cmdChan)
    
    // make IPC server
    ipcServer := jobber.NewIpcServer(cmdChan)
    go stopServerOnSigint(ipcServer)
    err = ipcServer.Launch()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
    defer ipcServer.Stop()
    
    manager.Wait()
}
