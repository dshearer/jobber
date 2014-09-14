package jobber

import (
    "time"
    "log"
    "os"
    "os/user"
    "path/filepath"
    "sync"
    "fmt"
    "strings"
    "sort"
    "code.google.com/p/go.net/context"
)

const (
    HomeDirRoot    = "/home"
    JobberFileName = ".jobber"
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
    Job     *Job
    Time    time.Time
    Result  JobStatus
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
    return fmt.Sprintf("%v\t%v\t%v\t%v", e.Time, e.Job.Name, e.Job.User, e.Result)
}

type JobManager struct {
    jobs            []*Job
    loadedJobs      bool
    runLog          []RunLogEntry
    doneChan        <-chan bool
    Shell           string
}

func NewJobManager() *JobManager {
    jm := JobManager{Shell: "/bin/sh"}
    jm.loadedJobs = false
    return &jm
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

func (m *JobManager) loadJobs() error {
    err := filepath.Walk(HomeDirRoot, m.procHomeFile)
    fmt.Printf("Loaded %v jobs.\n", len(m.jobs))
    return err
}

func (m *JobManager) procHomeFile(path string, info os.FileInfo, err error) error {
    if err != nil {
        return err
    } else if path == HomeDirRoot {
        return nil
    } else if info.IsDir() {
        username := filepath.Base(path)
        
        // check whether this user exists
        _, err = user.Lookup(username)
        if err == nil {
            /* User exists. */
            
            jobberFilePath := filepath.Join(path, JobberFileName)
            f, err := os.Open(jobberFilePath)
            if err != nil {
                if os.IsNotExist(err) {
                    return filepath.SkipDir
                } else {
                    return err
                }
            }
            defer f.Close()
            newJobs, err := ReadJobFile(f, username)
            m.jobs = append(m.jobs, newJobs...)
        }
        
        return filepath.SkipDir
    } else {
        return nil
    }
}

func (m *JobManager) reloadJobs(username string) error {
    // remove user's jobs
    newJobList := make([]*Job, 0)
    for _, job := range m.jobs {
        if job.User != username {
            newJobList = append(newJobList, job)
        }
    }
    fmt.Printf("Removed %v jobs.\n", len(m.jobs) - len(newJobList))
    m.jobs = newJobList
    
    // reload user's .jobber file
    jobberFilePath := filepath.Join(HomeDirRoot, username, JobberFileName)
    f, err := os.Open(jobberFilePath)
    if err != nil {
        if os.IsNotExist(err) {
            return nil
        } else {
            return err
        }
    }
    defer f.Close()
    newJobs, err := ReadJobFile(f, username)
    m.jobs = append(m.jobs, newJobs...)
    fmt.Printf("Loaded %v new jobs.\n", len(newJobs))
    
    return nil
}

func (m *JobManager) Launch(cmdChan <-chan ICmd) error {
    if !m.loadedJobs {
        err := m.loadJobs()
        if err != nil {
            return err
        }
    }
    
    // make main thread
    m.doneChan = m.runMainThread(cmdChan)
    return nil
}

func (m *JobManager) Wait() {
    <-m.doneChan
}

func (m *JobManager) runMainThread(cmdChan <-chan ICmd) <-chan bool {
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
                    fmt.Printf("JobManager: Run rec channel closed.\n")
                    break Loop
                } else {
                    fmt.Printf("JobManager: processing run rec.\n")
                
                    if len(rec.Stdout) > 0 {
                        rec.Job.stdoutLogger.Println(rec.Stdout)
                    }
                    if len(rec.Stderr) > 0 {
                        rec.Job.stderrLogger.Println(rec.Stderr)
                    }
                    if rec.Err != nil {
                        log.Panicln(rec.Err)
                    }
                    rec.Job.Status = rec.NewStatus
                    rec.Job.LastRunTime = rec.RunTime
                    m.runLog = append(m.runLog, RunLogEntry{rec.Job, rec.RunTime, rec.Job.Status})
                }
    
            case cmd, ok := <-cmdChan:
                if ok {
                    fmt.Printf("JobManager: processing cmd.\n")
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
        var jobWaitGroup sync.WaitGroup
        ticker := time.Tick(time.Duration(5 * time.Second))
        Loop: for {
            select {
            case now := <-ticker:
                for _, j := range m.jobsToRun(now) {
                    jobWaitGroup.Add(1)
                    go func(job *Job) {
                        runRecChan <- job.Run(ctx, m.Shell)
                        jobWaitGroup.Done()
                    }(j)
                }
        
            case <-ctx.Done():
                break Loop
            }
        }
    
        // clean up
        fmt.Printf("JobRunner: cleaning up...\n")
        jobWaitGroup.Wait()
        close(runRecChan)
        fmt.Printf("JobRunner: done cleaning up.\n")
    }()
    
    return runRecChan
}

func (m *JobManager) doCmd(cmd ICmd, cancel context.CancelFunc) {
    fmt.Printf("Got command: %v.\n", cmd);
    
    switch cmd.(type) {
    case *ReloadCmd:
        reloadCmd := cmd.(*ReloadCmd)
        
        // load new job config
        m.reloadJobs(cmd.RequestingUser())
        
        // send response
        reloadCmd.RespChan() <- &SuccessCmdResp{}
    
    case *ListJobsCmd:
        strs := make([]string, 0, len(m.jobs))
        for _, job := range m.jobsForUser(cmd.RequestingUser()) {
            strs = append(strs, job.String())
        }
        cmd.RespChan() <- &SuccessCmdResp{strings.Join(strs, "\n")}
    
    case *ListHistoryCmd:
        // get run-log entries
        entries := m.runLogEntriesForUser(cmd.RequestingUser())
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
            cancel()
        
            // send response
            cmd.RespChan() <- &SuccessCmdResp{}
        }
    
    default:
        cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "Unknown command."}}
    }
}
