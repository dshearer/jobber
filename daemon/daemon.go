package main

import (
    "github.com/dshearer/jobber"
    "os"
    "fmt"
)

func main() {
    var err error
    
    // read jobs
    f, err := os.Open("go_workspace/src/github.com/dshearer/jobber/example.json")
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
    
    // run them
	cmdChan := make(chan jobber.ICmd)
	manager := jobber.JobManager{Shell: "/bin/ksh"}
	manager.LoadJobs(f)
	manager.Launch(cmdChan)
    
    // make IPC server
    ipcServer := jobber.NewIpcServer(cmdChan)
    err = ipcServer.Launch()
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error:", err)
        os.Exit(1)
    }
    defer ipcServer.Stop()
    
    manager.Wait()
}
