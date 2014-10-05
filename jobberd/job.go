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

type TimePred struct {
    apply func(int) bool
    desc string
}

func (p TimePred) String() string {
    return p.desc
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
    Sec             TimePred
    Min             TimePred
    Hour            TimePred
    Mday            TimePred
    Mon             TimePred
    Wday            TimePred
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
    job.Sec = TimePred{func (i int) bool { return true }, "*"}
    job.Min = TimePred{func (i int) bool { return true }, "*"}
    job.Hour = TimePred{func (i int) bool { return true }, "*"}
    job.Mday = TimePred{func (i int) bool { return true }, "*"}
    job.Mon = TimePred{func (i int) bool { return true }, "*"}
    job.Wday = TimePred{func (i int) bool { return true }, "*"}
    job.ErrorHandler = &ErrorHandlerContinue
    job.NotifyOnError = false
    job.NotifyOnFailure = true
    return job
}

func monthToInt(m time.Month) int {
    switch m {
        case time.January : return 1
        case time.February : return 2
        case time.March : return 3
        case time.April : return 4
        case time.May : return 5
        case time.June : return 6
        case time.July : return 7
        case time.August : return 8
        case time.September : return 9
        case time.October : return 10
        case time.November : return 11
        default : return 12
    }
}

func weekdayToInt(d time.Weekday) int {
    switch d {
        case time.Sunday: return 0
        case time.Monday: return 1
        case time.Tuesday: return 2
        case time.Wednesday: return 3
        case time.Thursday: return 4
        case time.Friday: return 5
        default: return 6
    }
}

func (job *Job) ShouldRun(now time.Time) bool {
    if job.Status == JobFailed {
        return false
    } else if job.shouldRun_time(now) {
        if job.Status == JobBackoff {
            job.backoffTillNextTry--
            return job.backoffTillNextTry <= 0
        } else {
            return true
        }
    } else {
        return false
    }
}

func (job *Job) shouldRun_time(now time.Time) bool {
    if !job.Sec.apply(now.Second()) {
        return false
    } else if !job.Min.apply(now.Minute()) {
        return false
    } else if !job.Hour.apply(now.Hour()) {
        return false
    } else if !job.Mday.apply(now.Day()) {
        return false
    } else if !job.Mon.apply(monthToInt(now.Month())) {
        return false
    } else if !job.Wday.apply(weekdayToInt(now.Weekday())) {
        return false
    } else {
        return true
    }
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

