package main

import (
    "log"
    "time"
    "fmt"
    "os/exec"
    "io/ioutil"
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

type Job struct {
    // params
    Name        string
    Sec         TimePred
    Min         TimePred
    Hour        TimePred
    Mday        TimePred
    Mon         TimePred
    Wday        TimePred
    Cmd         string
    User        string
    
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
    return fmt.Sprintf("%v\t%v\t\t%v\t%v\t%v\t%v\t%v\t%v\t\"%v\"\t",
                       j.Name,
                       j.Status,
                       j.Sec,
                       j.Min,
                       j.Hour,
                       j.Mday,
                       j.Mon,
                       j.Wday,
                       j.Cmd)
}

func NewJob(name string, cmd string, username string) *Job {
    job := &Job{Name: name, Cmd: cmd, Status: JobGood, User: username}
    job.Sec = TimePred{func (i int) bool { return true }, "*"}
    job.Min = TimePred{func (i int) bool { return true }, "*"}
    job.Hour = TimePred{func (i int) bool { return true }, "*"}
    job.Mday = TimePred{func (i int) bool { return true }, "*"}
    job.Mon = TimePred{func (i int) bool { return true }, "*"}
    job.Wday = TimePred{func (i int) bool { return true }, "*"}
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
    } else if job.Status == JobBackoff {
        job.backoffTillNextTry--
        if job.backoffTillNextTry == 0 {
            return job.shouldRun_time(now)
        } else {
            return false
        }
    } else {
        return job.shouldRun_time(now)
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
    Err         *JobberError
}

func (job *Job) Run(ctx context.Context, shell string) *RunRec {
    //log.Println("Running " + job.Name)
    rec := &RunRec{Job: job, RunTime: time.Now(), NewStatus: JobGood}
    
    /*
    var cmd *exec.Cmd = exec.Command("sudo", 
                                     "-u", job.User,
                                     "-n", // non-interactive
                                     "-H", // set "HOME" env var to user's home dir
                                     shell, "-c", job.Cmd) // run user's shell and pass job.Cmd to it
    */
    var cmd *exec.Cmd = exec.Command("su",
                                     "--login", // login shell
                                     "--shell", shell,
                                     "--command", job.Cmd,
                                     job.User)
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        rec.Err = &JobberError{"Failed to get pipe to stdout.", err}
        return rec
    }
    stderr, err := cmd.StderrPipe()
    if err != nil {
        rec.Err = &JobberError{"Failed to get pipe to stderr.", err}
        return rec
    }
    
    // start cmd
    if err := cmd.Start(); err != nil {
        /* Failed to start command. */
        rec.Stderr = "Failed to run: " + err.Error()
        rec.NewStatus = JobFailed
        return rec
    }
    
    // read output
    stdoutBytes, err := ioutil.ReadAll(stdout)
    if err != nil {
        rec.Err = &JobberError{"Failed to read stdout.", err}
        return rec
    }
    rec.Stdout = string(stdoutBytes)
    stderrBytes, err := ioutil.ReadAll(stderr)
    if err != nil {
        rec.Err = &JobberError{"Failed to read stderr.", err}
        return rec
    }
    rec.Stderr = string(stderrBytes)
    
    // finish execution
    err = cmd.Wait()
    if err != nil {
        job.expBackoff()
        rec.NewStatus = job.Status
        
        switch err.(type) {
        case *exec.ExitError: 
            return rec
        
        default:
            rec.Err = &JobberError{"Error", err}
            return rec
        }
    } else {
        return rec
    }
}

func (job *Job) expBackoff() {
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
}

