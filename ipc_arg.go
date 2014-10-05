package jobber

type IpcArg struct {
    User string
    ForAllUsers bool
    
    // For TestCmd:
    Job string
    JobUser string
}
