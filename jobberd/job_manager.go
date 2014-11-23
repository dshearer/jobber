package main

import (
    "time"
    "log"
    "log/syslog"
    "sync"
    "fmt"
    "strings"
    "sort"
    "code.google.com/p/go.net/context"
    "text/tabwriter"
    "bytes"
)

type JobberError struct {
    What  string
    Cause error
}

func (e *JobberError) Error() string {
    if e.Cause == nil {
        return e.What
    } else {
        return e.What + ":" + e.Cause.Error()
    }
}

type RunLogEntry struct {
    Job        *Job
    Time       time.Time
    Succeeded  bool
    Result     JobStatus
}

/* For sorting RunLogEntries: */
type runLogEntrySorter struct {
    entries []RunLogEntry
}

/* For sorting RunLogEntries: */
func (s *runLogEntrySorter) Len() int {
    return len(s.entries)
}

/* For sorting RunLogEntries: */
func (s *runLogEntrySorter) Swap(i, j int) {
    s.entries[i], s.entries[j] = s.entries[j], s.entries[i]
}

/* For sorting RunLogEntries: */
func (s *runLogEntrySorter) Less(i, j int) bool {
    return s.entries[i].Time.After(s.entries[j].Time)
}

type JobManager struct {
    jobs                  []*Job
    loadedJobs            bool
    runLog                []RunLogEntry
    cmdChan               chan ICmd
    mainThreadCtx         context.Context
    doneChan              <-chan bool
    jobRunThreadDoneChan  chan bool
    runRecChan            chan *RunRec
    stopMainThread        func()
    stopJobRunnerThread   func()
    logger                *log.Logger
    errorLogger           *log.Logger
    Shell                 string
}

func NewJobManager() (*JobManager, error) {
    var err error
    jm := JobManager{Shell: "/bin/sh"}
    jm.logger, err = syslog.NewLogger(syslog.LOG_NOTICE | syslog.LOG_CRON, 0)
    if err != nil {
        return nil, &JobberError{What: "Couldn't make Syslog logger.", Cause: err}
    }
    jm.errorLogger, err = syslog.NewLogger(syslog.LOG_ERR | syslog.LOG_CRON, 0)
    if err != nil {
        return nil, &JobberError{What: "Couldn't make Syslog logger.", Cause: err}
    }
    jm.loadedJobs = false
    return &jm, nil
}

func (m *JobManager) jobsToRun(now time.Time) []*Job {
    jobs := make([]*Job, 0)
    for _, job := range m.jobs {
        if job.ShouldRun(now) {
            jobs = append(jobs, job)
        }
    }
    return jobs
}

func (m *JobManager) jobsForUser(username string) []*Job {
    jobs := make([]*Job, 0)
    for _, job := range m.jobs {
        if username == job.User {
            jobs = append(jobs, job)
        }
    }
    return jobs
}

func (m *JobManager) runLogEntriesForUser(username string) []RunLogEntry {
    entries := make([]RunLogEntry, 0)
    for _, entry := range m.runLog {
        if username == entry.Job.User {
            entries = append(entries, entry)
        }
    }
    return entries
}

func (m *JobManager) Launch() (chan<- ICmd, error) {
    m.logger.Println("Launching.")
    if !m.loadedJobs {
        _, err := m.LoadAllJobs()
        if err != nil {
            m.errorLogger.Printf("Failed to load jobs: %v.\n", err)
            return nil, err
        }
    }
    
    // make main thread
    m.cmdChan = make(chan ICmd)
    m.doneChan = m.runMainThread()
    return m.cmdChan, nil
}

func (m *JobManager) Wait() {
    <-m.doneChan
}

func (m *JobManager) Stop() {
    close(m.cmdChan)
    m.stopMainThread()
}

func (m *JobManager) runMainThread() <-chan bool {
    doneChan := make(chan bool, 1)
    
    ctx, cancel := context.WithCancel(context.Background())
    m.mainThreadCtx = ctx
    m.stopMainThread = func() {
        cancel()
        <-doneChan
    }
    
    go func() {
        /*
         All modifications to the job manager's state occur here.
        */
    
        // make job-runner thread
        m.runJobRunnerThread()
    
        Loop: for {
            select {
            case <-ctx.Done():
                break Loop
                
            case rec, ok := <-m.runRecChan:
                if ok {
                    if len(rec.Stdout) > 0 {
                        m.logger.Println(rec.Stdout)
                    }
                    if len(rec.Stderr) > 0 {
                        m.errorLogger.Println(rec.Stderr)
                    }
                    if rec.Err != nil {
                        m.errorLogger.Panicln(rec.Err)
                    }
                    
                    m.runLog = append(m.runLog, RunLogEntry{rec.Job, rec.RunTime, rec.Succeeded, rec.NewStatus})
                    
                    /* NOTE: error-handler was already applied by the job, if necessary. */
                    
                    if (!rec.Succeeded && rec.Job.NotifyOnError) ||
                        (rec.Job.NotifyOnFailure && rec.NewStatus == JobFailed) {
                        // notify user
                        headers := fmt.Sprintf("To: %v\r\nFrom: %v\r\nSubject: \"%v\" failed.", rec.Job.User, rec.Job.User, rec.Job.Name)
                        bod := rec.Describe()
                        msg := fmt.Sprintf("%s\r\n\r\n%s.\r\n", headers, bod)
                        sendmailCmd := fmt.Sprintf("sendmail %v", rec.Job.User)
                        sudoResult, err := sudo(rec.Job.User, sendmailCmd, "/bin/sh", &msg)
                        if err != nil {
                            m.errorLogger.Println("Failed to send mail: %v", err)
                        } else if !sudoResult.Succeeded {
                            m.errorLogger.Println("Failed to send mail: %v", sudoResult.Stderr)
                        }
                    }
                }
    
            case cmd, ok := <-m.cmdChan:
                if ok {
                    //fmt.Printf("JobManager: processing cmd.\n")
                    m.doCmd(cmd)
                }
            }
        }
        
        // wait for job-running thread to quit
        <-m.jobRunThreadDoneChan
        
        // clean up
        doneChan <- true
        close(doneChan)
    }()
    
    return doneChan
}

func (m *JobManager) runJobRunnerThread() {
    m.runRecChan = make(chan *RunRec)
    m.jobRunThreadDoneChan = make(chan bool)
    
    subctx, cancel := context.WithCancel(m.mainThreadCtx)
    m.stopJobRunnerThread = func() {
        cancel()
        <-m.jobRunThreadDoneChan
    }
    
    go func() {
        var jobWaitGroup sync.WaitGroup
        ticker := time.Tick(time.Duration(1 * time.Second))
        Loop: for {
            select {
            case now := <-ticker:
                jobsToRun := m.jobsToRun(now)
                for _, j := range jobsToRun {
                    m.logger.Printf("%v: %v\n", j.User, j.Cmd)
                    jobWaitGroup.Add(1)
                    go func(job *Job) {
                        m.runRecChan <- job.Run(subctx, m.Shell, false)
                        jobWaitGroup.Done()
                    }(j)
                }
        
            case <-subctx.Done():
                break Loop
            }
        }
    
        // clean up
        //fmt.Printf("JobRunner: cleaning up...\n")
        jobWaitGroup.Wait()
        m.jobRunThreadDoneChan <- true
        close(m.jobRunThreadDoneChan)
        //m.logger.Printf("JobRunner: done cleaning up.\n")
    }()
}

func (m *JobManager) doCmd(cmd ICmd) {
    
    /*
    Security:
    
    It is jobberd's responsibility to enforce the security policy.
    
    It does so by assuming that cmd.RequestingUser() is truly the name
    of the requesting user.
    */
    
    switch cmd.(type) {
    case *ReloadCmd:
        /* Policy: Only root can reload other users' jobfiles. */
        
        // load jobs
        var err error
        var amt int
        if cmd.(*ReloadCmd).ForAllUsers {
            if cmd.RequestingUser() != "root" {
                cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "You must be root."}}
                break
            }
            
            m.logger.Printf("Reloading jobs for all users.\n")
            amt, err = m.ReloadAllJobs()
        } else {
            m.logger.Printf("Reloading jobs for %v.\n", cmd.RequestingUser())
            amt, err = m.ReloadJobsForUser(cmd.RequestingUser())
        }
        
        // send response
        if err != nil {
            m.errorLogger.Printf("Failed to load jobs: %v.\n", err)
            cmd.RespChan() <- &ErrorCmdResp{err}
        } else {
            cmd.RespChan() <- &SuccessCmdResp{fmt.Sprintf("Loaded %v jobs.", amt)}
        }
    
    case *ListJobsCmd:
        /* Policy: Only root can list other users' jobs. */
        
        // get jobs
        var jobs []*Job
        if cmd.(*ListJobsCmd).ForAllUsers {
            if cmd.RequestingUser() != "root" {
                cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "You must be root."}}
                break
            }
            
            jobs = m.jobs
        } else {
            jobs = m.jobsForUser(cmd.RequestingUser()) 
        }
        
        // make response
        var buffer bytes.Buffer
        var writer *tabwriter.Writer = tabwriter.NewWriter(&buffer, 5, 0, 2, ' ', 0)
        fmt.Fprintf(writer, "NAME\tSTATUS\tSEC\tMIN\tHOUR\tMDAY\tMONTH\tWDAY\tCOMMAND\tNOTIFY ON ERROR\tNOTIFY ON FAILURE\tERROR HANDLER\t\n")
        strs := make([]string, 0, len(m.jobs))
        for _, j := range jobs {
            s := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t\"%v\"\t%v\t%v\t%v\t",
                               j.Name,
                               j.Status,
                               j.Sec,
                               j.Min,
                               j.Hour,
                               j.Mday,
                               j.Mon,
                               j.Wday,
                               j.Cmd,
                               j.NotifyOnError,
                               j.NotifyOnFailure,
                               j.ErrorHandler)
            strs = append(strs, s)
        }
        fmt.Fprintf(writer, "%v", strings.Join(strs, "\n"))
        writer.Flush()
        
        // send response
        cmd.RespChan() <- &SuccessCmdResp{buffer.String()}
    
    case *ListHistoryCmd:
        /* Policy: Only root can see the histories of other users' jobs. */
        
        // get log entries
        var entries []RunLogEntry
        if cmd.(*ListHistoryCmd).ForAllUsers {
            if cmd.RequestingUser() != "root" {
                cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "You must be root."}}
                break
            }
            
            entries = m.runLog
        } else {
            entries = m.runLogEntriesForUser(cmd.RequestingUser()) 
        }
        sort.Sort(&runLogEntrySorter{entries})
        
        // make response
        var buffer bytes.Buffer
        var writer *tabwriter.Writer = tabwriter.NewWriter(&buffer, 5, 0, 2, ' ', 0)
        fmt.Fprintf(writer, "TIME\tJOB\tUSER\tSUCCEEDED\tRESULT\t\n")
        strs := make([]string, 0, len(m.jobs))
        for _, e := range entries {
            s := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t", e.Time, e.Job.Name, e.Job.User, e.Succeeded, e.Result)
            strs = append(strs, s)
        }
        fmt.Fprintf(writer, "%v", strings.Join(strs, "\n"))
        writer.Flush()
    
        // send response
        cmd.RespChan() <- &SuccessCmdResp{buffer.String()}
    
    case *StopCmd:
        /* Policy: Only root can stop jobberd. */
        
        if cmd.RequestingUser() != "root" {
            cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "You must be root."}}
            break
        }
        
        m.logger.Println("Stopping.")
        
        // stop job-runner thread
        m.stopJobRunnerThread()
    
        // send response
        cmd.RespChan() <- &SuccessCmdResp{}
    
    case *TestCmd:
        /* Policy: Only root can test other users' jobs. */
        
        var testCmd *TestCmd = cmd.(*TestCmd)
        
        // enfore policy
        if testCmd.jobUser != testCmd.RequestingUser() && testCmd.RequestingUser() != "root" {
            cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "You must be root."}}
            break
        }
    
        // find job to test
        var job_p *Job
        for _, job := range m.jobsForUser(testCmd.jobUser) {
            if job.Name == testCmd.job {
                job_p = job
                break
            }
        }
        if job_p == nil {
            msg := fmt.Sprintf("No job named \"%v\".", testCmd.job)
            cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: msg}}
            break
        }
        
        // run the job in this thread
        runRec := job_p.Run(nil, m.Shell, true)
        
        // send response
        if runRec.Err != nil {
            cmd.RespChan() <- &ErrorCmdResp{runRec.Err}
            break
        }
        cmd.RespChan() <- &SuccessCmdResp{Details: runRec.Describe()}
    
    default:
        cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "Unknown command."}}
    }
}
