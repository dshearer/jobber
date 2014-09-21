package main

import (
    "github.com/dshearer/jobber"
    "net"
    "net/rpc"
    "fmt"
    "os"
    "os/user"
    "syscall"
    "flag"
)

const (
    ListCmdStr   = "list"
    LogCmdStr    = "log"
    ReloadCmdStr = "reload"
    StopCmdStr   = "stop"
)

func usage() {
    fmt.Printf("Usage: %v [flags] (%v|%v|%v|%v)\nFlags:\n", os.Args[0], ListCmdStr, LogCmdStr, ReloadCmdStr, StopCmdStr)
    flag.PrintDefaults()
}

func subcmdUsage(subcmd string, flagSet *flag.FlagSet) func() {
    return func() {
        fmt.Printf("\nUsage: %v %v [flags]\nFlags:\n", os.Args[0], subcmd)
        flagSet.PrintDefaults()
    }
}

func failIfNotRoot(user *user.User) {
    if user.Uid != "0" {
        fmt.Fprintf(os.Stderr, "You must be root.\n")
        os.Exit(1)
    }
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
            fmt.Fprintf(os.Stderr, "Specify a command, asshole.\n\n")
            flag.Usage()
            os.Exit(1)
        }
        
        // make sure the daemon is running
        if _, err := os.Stat(jobber.DaemonSocketAddr); os.IsNotExist(err) {
            fmt.Fprintf(os.Stderr, "jobberd isn't running.\n")
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
        
        // drop privileges
        err = syscall.Setreuid(syscall.Getuid(), syscall.Getuid())
        if err != nil {
            fmt.Fprintf(os.Stderr, "Couldn't drop privileges: %v\n", err)
            os.Exit(1)
        }
        
        // get current username
        user, err := user.Current()
        if err != nil {
            fmt.Fprintf(os.Stderr, "Couldn't get current user: %v\n", err)
            os.Exit(1)
        }
        
        // do command
        switch flag.Arg(0) {
        case ListCmdStr:
            doListCmd(flag.Args()[1:], rpcClient, user)
        
        case LogCmdStr:
            doLogCmd(flag.Args()[1:], rpcClient, user)
        
        case ReloadCmdStr:
            doReloadCmd(flag.Args()[1:], rpcClient, user)
        
        case StopCmdStr:
            var result string
            arg := jobber.IpcArg{user.Username, false}
            err = rpcClient.Call("RealIpcServer.Stop", arg, &result)
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

func doListCmd(args []string, rpcClient *rpc.Client, user *user.User) {
    // parse flags
    flagSet := flag.NewFlagSet("list", flag.ExitOnError)
    flagSet.Usage = subcmdUsage("list", flagSet)
    var help_p = flagSet.Bool("h", false, "help")
    var allUsers_p = flagSet.Bool("a", false, "all-users")
    flagSet.Parse(args)
    
    if *help_p {
        flagSet.Usage()
        os.Exit(0)
    } else {
        if *allUsers_p {
            failIfNotRoot(user)
        }
        
        var result string
        arg := jobber.IpcArg{user.Username, *allUsers_p}
        err := rpcClient.Call("RealIpcServer.ListJobs", arg, &result)
        if err != nil {
            fmt.Fprintf(os.Stderr, "RPC failed: %v\n", err)
            os.Exit(1)
        }
        
        // print result
        fmt.Printf("%s\n", result)
    }
}

func doLogCmd(args []string, rpcClient *rpc.Client, user *user.User) {
    // parse flags
    flagSet := flag.NewFlagSet("log", flag.ExitOnError)
    flagSet.Usage = subcmdUsage("log", flagSet)
    var help_p = flagSet.Bool("h", false, "help")
    var allUsers_p = flagSet.Bool("a", false, "all-users")
    flagSet.Parse(args)
    
    if *help_p {
        flagSet.Usage()
        os.Exit(0)
    } else {
        if *allUsers_p {
            failIfNotRoot(user)
        }
        
        var result string
        arg := jobber.IpcArg{user.Username, *allUsers_p}
        err := rpcClient.Call("RealIpcServer.ListHistory", arg, &result)
        if err != nil {
            fmt.Fprintf(os.Stderr, "RPC failed: %v\n", err)
            os.Exit(1)
        }
        
        // print result
        fmt.Printf("%s\n", result)
    }
}

func doReloadCmd(args []string, rpcClient *rpc.Client, user *user.User) {
    // parse flags
    flagSet := flag.NewFlagSet("reload", flag.ExitOnError)
    flagSet.Usage = subcmdUsage("reload", flagSet)
    var help_p = flagSet.Bool("h", false, "help")
    var allUsers_p = flagSet.Bool("a", false, "all-users")
    flagSet.Parse(args)
    
    if *help_p {
        flagSet.Usage()
        os.Exit(0)
    } else {
        if *allUsers_p {
            failIfNotRoot(user)
        }
        
        var result string
        arg := jobber.IpcArg{user.Username, *allUsers_p}
        err := rpcClient.Call("RealIpcServer.Reload", arg, &result)
        if err != nil {
            fmt.Fprintf(os.Stderr, "RPC failed: %v\n", err)
            os.Exit(1)
        }
    }
}
