package jobber

import (
    "time"
    "log"
    "io"
    "sync"
    "fmt"
    "strings"
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
    Job     *Job
    Time    time.Time
    Result  JobStatus
}

func (e *RunLogEntry) String() string {
    return fmt.Sprintf("%v %v %v", e.Job.Name, e.Time, e.Result)
}

type JobManager struct {
    jobs            []*Job
    runLog          []RunLogEntry
    doneChan        <-chan bool
    Shell           string
}

func (m *JobManager) jobsToRun(now time.Time) []*Job {
    jobs := make([]*Job, 0)
    for _, job := range m.jobs {
        if job.ShouldRun(now) {
            jobs = append(jobs, job)
        } else {
        }
    }
    return jobs
}

func (m *JobManager) LoadJobs(r io.Reader) error {
    jobs, err := ReadJobFile(r)
    m.jobs = jobs
    return err
}

func (m *JobManager) Launch(cmdChan <-chan ICmd) {
    // make main thread
    m.doneChan = m.runMainThread(cmdChan)
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
                    cancel()
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
                fmt.Printf("JobManager: processing cmd.\n")
                if !ok {
                    /* Channel is closed. */
                    log.Panicln("JobManager: Cmd channel closed unexpectedly.")
                    break Loop
                } else {
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
        m.LoadJobs(reloadCmd.JobFile)
        
        // send response
        reloadCmd.RespChan() <- &SuccessCmdResp{}
    
    case *ListJobsCmd:
        strs := make([]string, 0, len(m.jobs))
        for _, job := range m.jobs {
            strs = append(strs, job.String())
        }
        cmd.RespChan() <- &SuccessCmdResp{strings.Join(strs, "\n")}
    
    case *ListHistoryCmd:
        // send response
        strs := make([]string, 0, len(m.runLog))
        for _, entry := range m.runLog {
            strs = append(strs, entry.String())
        }
        cmd.RespChan() <- &SuccessCmdResp{strings.Join(strs, "\n")}
    
    case *StopCmd:
        // stop
        cancel()
        
        // send response
        cmd.RespChan() <- &SuccessCmdResp{}
    
    default:
        cmd.RespChan() <- &ErrorCmdResp{&JobberError{What: "Unknown command."}}
    }
}
