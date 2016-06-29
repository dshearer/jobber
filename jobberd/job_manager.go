package main

import (
    "time"
    "fmt"
    "strings"
    "sort"
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
    userPrefs             map[string]UserPrefs
    jobs                  []*Job
    loadedJobs            bool
    runLog                []RunLogEntry
    cmdChan               chan ICmd
    mainThreadCtx         *JobberContext
    mainThreadCtl         JobberCtl
    jobRunner             *JobRunnerThread
    Shell                 string
}

func NewJobManager() (*JobManager, error) {
    jm := JobManager{Shell: "/bin/sh"}
    jm.userPrefs = make(map[string]UserPrefs)
    jm.loadedJobs = false
    jm.jobRunner = NewJobRunnerThread()
    return &jm, nil
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
    if m.mainThreadCtx != nil {
        return nil, &JobberError{"Already launched.", nil}
    }
    
    Logger.Println("Launching.")
    if !m.loadedJobs {
        _, err := m.LoadAllJobs()
        if err != nil {
            ErrLogger.Printf("Failed to load jobs: %v.\n", err)
            return nil, err
        }
    }
    
    // make main thread
    m.cmdChan = make(chan ICmd)
    m.runMainThread()
    return m.cmdChan, nil
}

func (m *JobManager) Cancel() {
    if m.mainThreadCtl.Cancel != nil {
        Logger.Printf("JobManager canceling\n")
        m.mainThreadCtl.Cancel()
    }
}

func (m *JobManager) Wait() {
    if m.mainThreadCtl.Wait != nil {
        m.mainThreadCtl.Wait()
    }
}

func (m *JobManager) handleRunRec(rec *RunRec) {
    if rec.Err != nil {
        ErrLogger.Panicln(rec.Err)
    }
    
    m.runLog = append(m.runLog, 
              RunLogEntry{rec.Job, rec.RunTime, rec.Succeeded, rec.NewStatus})
    
    /* NOTE: error-handler was already applied by the job, if necessary. */
    
    if (!rec.Succeeded && rec.Job.NotifyOnError) ||
        (rec.Job.NotifyOnFailure && rec.NewStatus == JobFailed) {
        // notify user
        m.userPrefs[rec.Job.User].Notifier(rec);
    }
}

func (m *JobManager) runMainThread() {
    m.mainThreadCtx, m.mainThreadCtl = NewJobberContext(BackgroundJobberContext())
    Logger.Printf("Main thread context: %v\n", m.mainThreadCtx.Name)
    
    go func() {
        /*
         All modifications to the job manager's state occur here.
        */
    
        // start job-runner thread
        m.jobRunner.Start(m.jobs, m.Shell, m.mainThreadCtx)
    
        Loop: for {
            select {
            case <-m.mainThreadCtx.Done():
                Logger.Printf("Main thread got 'stop'\n")
                break Loop
                
            case rec, ok := <-m.jobRunner.RunRecChan():
                if ok {
                    m.handleRunRec(rec)
                } else {
                    ErrLogger.Printf("Job-runner thread ended prematurely.\n")
                    break Loop
                }
    
            case cmd, ok := <-m.cmdChan:
                if ok {
                    //fmt.Printf("JobManager: processing cmd.\n")
                    shouldStop := m.doCmd(cmd)
                    if shouldStop {
                        break Loop
                    }
                } else {
                    ErrLogger.Printf("Command channel was closed.\n")
                    break Loop
                }
            }
        }
        
        // cancel main thread
        m.mainThreadCtl.Cancel()
        
        // consume all run-records
        for rec := range m.jobRunner.RunRecChan() {
            m.handleRunRec(rec)
        }
        
        // finish up (and wait for job-runner thread to finish)
        m.mainThreadCtx.Finish()
        
        Logger.Printf("Main Thread done.\n")
    }()
}

func (m *JobManager) doCmd(cmd ICmd) bool {  // runs in main thread
    
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
            
            Logger.Printf("Reloading jobs for all users.\n")
            amt, err = m.ReloadAllJobs()
        } else {
            Logger.Printf("Reloading jobs for %v.\n", cmd.RequestingUser())
            amt, err = m.ReloadJobsForUser(cmd.RequestingUser())
        }
        
        // send response
        if err != nil {
            ErrLogger.Printf("Failed to load jobs: %v.\n", err)
            cmd.RespChan() <- &ErrorCmdResp{err}
        } else {
            cmd.RespChan() <- &SuccessCmdResp{fmt.Sprintf("Loaded %v jobs.", amt)}
        }
        
        return false
    
    case *CatCmd:
        /* Policy: Only root can cat other users' jobs. */
        
        var catCmd *CatCmd = cmd.(*CatCmd)
        
        // enfore policy
        if catCmd.jobUser != catCmd.RequestingUser() && catCmd.RequestingUser() != "root" {
            cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "You must be root."}}
            break
        }
    
        // find job to cat
        var job_p *Job
        for _, job := range m.jobsForUser(catCmd.jobUser) {
            if job.Name == catCmd.job {
                job_p = job
                break
            }
        }
        if job_p == nil {
            msg := fmt.Sprintf("No job named \"%v\".", catCmd.job)
            cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: msg}}
            break
        }
        
        // make and send response
        cmd.RespChan() <- &SuccessCmdResp{job_p.Cmd}
        
        return false
    
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
        fmt.Fprintf(writer, "NAME\tSTATUS\tSEC/MIN/HR/MDAY/MTH/WDAY\tNEXT RUN TIME\tNOTIFY ON ERR\tNOTIFY ON FAIL\tERR HANDLER\n")
        strs := make([]string, 0, len(m.jobs))
        for _, j := range jobs {
            schedStr := fmt.Sprintf("%v %v %v %v %v %v", 
                                    j.FullTimeSpec.Sec,
                                    j.FullTimeSpec.Min,
                                    j.FullTimeSpec.Hour,
                                    j.FullTimeSpec.Mday,
                                    j.FullTimeSpec.Mon,
                                    j.FullTimeSpec.Wday)
            var runTimeStr string = "none"
            if j.NextRunTime != nil {
                runTimeStr = j.NextRunTime.Format("Jan _2 15:04:05 2006")
            }
            s := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v",
                               j.Name,
                               j.Status,
                               schedStr,
                               runTimeStr,
                               j.NotifyOnError,
                               j.NotifyOnFailure,
                               j.ErrorHandler)
            strs = append(strs, s)
        }
        fmt.Fprintf(writer, "%v", strings.Join(strs, "\n"))
        writer.Flush()
        
        // send response
        cmd.RespChan() <- &SuccessCmdResp{buffer.String()}
        
        return false
    
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
        
        return false
    
    case *StopCmd:
        /* Policy: Only root can stop jobberd. */
        
        if cmd.RequestingUser() != "root" {
            cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "You must be root."}}
            break
        }
        
        Logger.Println("Stopping.")
        return true
    
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
        
        return false
    
    default:
        cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "Unknown command."}}
        return false
    }
    
    return false
}
