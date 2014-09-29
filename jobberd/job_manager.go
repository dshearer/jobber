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
    return s.entries[i].Time.Before(s.entries[j].Time)
}

func (e *RunLogEntry) String() string {
    return fmt.Sprintf("%v\t%v\t%v\t%v\t%v", e.Time, e.Job.Name, e.Job.User, e.Succeeded, e.Result)
}

type JobManager struct {
    jobs            []*Job
    loadedJobs      bool
    runLog          []RunLogEntry
    cmdChan         chan ICmd
    doneChan        <-chan bool
    jobWaitGroup    sync.WaitGroup
    logger          *log.Logger
    errorLogger     *log.Logger
    Shell           string
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
        err := m.LoadAllJobs()
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
    m.cmdChan <- &StopCmd{"root", make(chan ICmdResp, 1)}
    close(m.cmdChan)
    m.Wait()
}

func (m *JobManager) runMainThread() <-chan bool {
    doneChan := make(chan bool, 1)
    go func() {
        /*
         All modifications to the job manager's state occur here.
        */
    
        // make main context
        ctx, cancel := context.WithCancel(context.Background())
    
        // make job-runner thread
        runRecChan := m.runJobRunnerThread(ctx)
    
        Loop: for {
            select {
            case rec, ok := <-runRecChan:
                if !ok {
                    /* Channel is closed. */
                    //fmt.Printf("JobManager: Run rec channel closed.\n")
                    break Loop
                } else {
                    //fmt.Printf("JobManager: processing run rec.\n")
                
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
                        bod := fmt.Sprintf("Job \"%v\" failed.  New status: %v.\r\n\r\nStdout:\r\n%v\r\n\r\nStderr:\r\n%v", rec.Job.Name, rec.Job.Status, rec.Stdout, rec.Stderr)
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
                    m.doCmd(cmd, cancel)
                }
            }
        }
        
        // clean up
        doneChan <- true
        close(doneChan)
    }()
    
    return doneChan
}

func (m *JobManager) runJobRunnerThread(ctx context.Context) <-chan *RunRec {
    runRecChan := make(chan *RunRec)
    
    go func() {
        ticker := time.Tick(time.Duration(1 * time.Second))
        Loop: for {
            select {
            case now := <-ticker:
                jobsToRun := m.jobsToRun(now)
                for _, j := range jobsToRun {
                    m.logger.Printf("%v: %v\n", j.User, j.Cmd)
                    m.jobWaitGroup.Add(1)
                    go func(job *Job) {
                        runRecChan <- job.Run(ctx, m.Shell)
                        m.jobWaitGroup.Done()
                    }(j)
                }
        
            case <-ctx.Done():
                break Loop
            }
        }
    
        // clean up
        //fmt.Printf("JobRunner: cleaning up...\n")
        m.jobWaitGroup.Wait()
        close(runRecChan)
        //fmt.Printf("JobRunner: done cleaning up.\n")
    }()
    
    return runRecChan
}

func (m *JobManager) doCmd(cmd ICmd, cancel context.CancelFunc) {
    //fmt.Printf("Got command: %v.\n", cmd);
    
    switch cmd.(type) {
    case *ReloadCmd:
        // load jobs
        var err error
        if cmd.(*ReloadCmd).ForAllUsers {
            m.logger.Printf("Reloading jobs for all users.\n")
            err = m.ReloadAllJobs()
        } else {
            m.logger.Printf("Reloading jobs for %v.\n", cmd.RequestingUser())
            err = m.ReloadJobsForUser(cmd.RequestingUser())
        }
        
        // send response
        if err != nil {
            m.errorLogger.Printf("Failed to load jobs: %v.\n", err)
            cmd.RespChan() <- &ErrorCmdResp{err}
        } else {
            cmd.RespChan() <- &SuccessCmdResp{}
        }
    
    case *ListJobsCmd:
        // get jobs
        var jobs []*Job
        if cmd.(*ListJobsCmd).ForAllUsers {
            jobs = m.jobs
        } else {
            jobs = m.jobsForUser(cmd.RequestingUser()) 
        }
        
        // send response
        strs := make([]string, 0, len(m.jobs))
        for _, job := range jobs {
            strs = append(strs, job.String())
        }
        cmd.RespChan() <- &SuccessCmdResp{strings.Join(strs, "\n")}
    
    case *ListHistoryCmd:
        // get log entries
        var entries []RunLogEntry
        if cmd.(*ListHistoryCmd).ForAllUsers {
            entries = m.runLog
        } else {
            entries = m.runLogEntriesForUser(cmd.RequestingUser()) 
        }
        sort.Sort(&runLogEntrySorter{entries})
    
        // send response
        strs := make([]string, 0, len(m.runLog))
        for _, entry := range entries {
            strs = append(strs, entry.String())
        }
        cmd.RespChan() <- &SuccessCmdResp{strings.Join(strs, "\n")}
    
    case *StopCmd:
        if cmd.RequestingUser() != "root" {
            cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "You must be root."}}
        } else {
            // stop
            m.logger.Println("Stopping.")
            
            cancel()
            m.jobWaitGroup.Wait()
        
            // send response
            cmd.RespChan() <- &SuccessCmdResp{}
        }
    
    default:
        cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "Unknown command."}}
    }
}
