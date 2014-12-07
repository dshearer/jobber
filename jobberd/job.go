package main

import (
    "log"
    "time"
    "fmt"
    "code.google.com/p/go.net/context"
)

const (
    MaxBackoffWait = 8
)

type JobStatus uint8
const (
    JobGood     JobStatus = 0
    JobFailed             = 1
    JobBackoff            = 2
)

func (s JobStatus) String() string {
    switch s {
        case JobGood:
            return "Good"
            
        case JobBackoff:
            return "Backoff"
            
        default:
            return "Failed"
    }
}

type TimeSpec int

func (t TimeSpec) String() string {
    if t == -1 {
        return "*"
    } else {
        return fmt.Sprintf("%v", int(t))
    }
}

const (
    ErrorHandlerStopName       = "Stop"
    ErrorHandlerBackoffName    = "Backoff"
    ErrorHandlerContinueName   = "Continue"
)

type ErrorHandler struct {
    apply func(*Job)
    desc string
}

func (h ErrorHandler) String() string {
    return h.desc
}

var ErrorHandlerStop = ErrorHandler{
    apply : func(job *Job) { job.Status = JobFailed },
    desc : ErrorHandlerStopName,
}

var ErrorHandlerBackoff = ErrorHandler{
    apply : func(job *Job) { 
        if job.Status == JobGood {
            job.Status = JobBackoff
            job.backoffWait = 1
        } else {
            job.backoffWait *= 2
        }

        job.backoffTillNextTry = job.backoffWait
        if job.backoffWait > MaxBackoffWait {
            // give up
            job.Status = JobFailed
            job.backoffWait = 0
            job.backoffTillNextTry = 0
        }
    },
    desc : ErrorHandlerBackoffName,
}

var ErrorHandlerContinue = ErrorHandler{
    apply : func(job *Job) { job.Status = JobGood },
    desc : ErrorHandlerContinueName,
}

type Job struct {
    // params
    Name            string
    Sec             TimeSpec
    Min             TimeSpec
    Hour            TimeSpec
    Mday            TimeSpec
    Mon             TimeSpec
    Wday            TimeSpec
    Cmd             string
    User            string
    ErrorHandler   *ErrorHandler
    NotifyOnError   bool
    NotifyOnFailure bool
    
    // other params
    stdoutLogger *log.Logger
    stderrLogger *log.Logger
    
    // dynamic shit
    Status      JobStatus
    LastRunTime time.Time
    
    // backoff after errors
    backoffWait         int
    backoffTillNextTry   int
}

func (j *Job) String() string {
    return j.Name
}

func NewJob(name string, cmd string, username string) *Job {
    job := &Job{Name: name, Cmd: cmd, Status: JobGood, User: username}
    job.Sec = -1
    job.Min = -1
    job.Hour = -1
    job.Mday = -1
    job.Mon = -1
    job.Wday = -1
    job.ErrorHandler = &ErrorHandlerContinue
    job.NotifyOnError = false
    job.NotifyOnFailure = true
    return job
}

type RunRec struct {
    Job         *Job
    RunTime     time.Time
    NewStatus   JobStatus
    Stdout      string
    Stderr      string
    Succeeded   bool
    Err         *JobberError
}

func (rec *RunRec) Describe() string {
    var summary string
    if rec.Succeeded {
        summary = fmt.Sprintf("Job \"%v\" succeeded.", rec.Job.Name)
    } else {
        summary = fmt.Sprintf("Job \"%v\" failed.", rec.Job.Name)
    }
    return fmt.Sprintf("%v\r\nNew status: %v.\r\n\r\nStdout:\r\n%v\r\n\r\nStderr:\r\n%v", summary, rec.Job.Status, rec.Stdout, rec.Stderr)
}

func (job *Job) Run(ctx context.Context, shell string, testing bool) *RunRec {
    rec := &RunRec{Job: job, RunTime: time.Now()}
    
    // run
    var sudoResult *SudoResult
    sudoResult, err := sudo(job.User, job.Cmd, shell, nil)
    
    if err != nil {
        /* unexpected error while trying to run job */
        rec.Err = err
        return rec
    }
    
    // update run rec
    rec.Succeeded = sudoResult.Succeeded
    rec.NewStatus = JobGood
    rec.Stdout = sudoResult.Stdout
    rec.Stderr = sudoResult.Stderr
    
    if !testing {
        // update job
        if sudoResult.Succeeded {
            /* job succeeded */
            job.Status = JobGood
        } else {
            /* job failed: apply error-handler (which sets job.Status) */
            job.ErrorHandler.apply(job)
        }
        job.LastRunTime = rec.RunTime
        
        // update rec.NewStatus
        rec.NewStatus = job.Status
    }
    
    return rec
}

