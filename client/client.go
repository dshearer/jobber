package main

import (
    "github.com/dshearer/jobber"
    "net"
    "net/rpc"
    "fmt"
    "os"
    "flag"
)

const (
    ListCmdStr = "list"
    LogCmdStr  = "log"
    StopCmdStr = "stop"
)

func usage() {
    fmt.Printf("\nUsage: %v [flags] (%v|%v|%v)\nFlags:\n", os.Args[0], ListCmdStr, LogCmdStr, StopCmdStr)
    flag.PrintDefaults()
}

func main() {
    flag.Usage = usage
    
    var helpFlag_p = flag.Bool("h", false, "help")
    flag.Parse()
    
    if *helpFlag_p {
        usage()
        os.Exit(0)
    } else {
        if len(flag.Args()) == 0 {
            fmt.Fprintf(os.Stderr, "Specify a command.\n")
            flag.Usage()
            os.Exit(1)
        }
        
        // connect to daemon
        addr, err := net.ResolveUnixAddr("unix", jobber.DaemonSocketAddr)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Couldn't resolve Unix addr: %v\n", err)
            os.Exit(1)
        }
        conn, err := net.DialUnix("unix", nil, addr)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Couldn't connect to daemon: %v\n", err)
            os.Exit(1)
        }
        defer conn.Close()
        rpcClient := rpc.NewClient(conn)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Couldn't make RPC client: %v\n", err)
            os.Exit(1)
        }
        
        // do command
        switch flag.Arg(0) {
        case ListCmdStr:
            var result string
            err = rpcClient.Call("RealIpcServer.ListJobs", 1, &result)
            if err != nil {
                fmt.Fprintf(os.Stderr, "RPC failed: %v\n", err)
                os.Exit(1)
            }
            fmt.Printf("%s\n", result)
        
        case LogCmdStr:
            var result string
            err = rpcClient.Call("RealIpcServer.ListHistory", 1, &result)
            if err != nil {
                fmt.Fprintf(os.Stderr, "RPC failed: %v\n", err)
                os.Exit(1)
            }
            fmt.Printf("%s\n", result)
        
        case StopCmdStr:
            var result string
            err = rpcClient.Call("RealIpcServer.Stop", 1, &result)
            if err != nil {
                fmt.Fprintf(os.Stderr, "RPC failed: %v\n", err)
                os.Exit(1)
            }
        
        default:
            fmt.Fprintf(os.Stderr, "Invalid command: \"%v\".\n", flag.Arg(0))
            flag.Usage()
            os.Exit(1)
        }
    }
}
