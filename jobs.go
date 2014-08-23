package jobber

import (
    "time"
    "log"
    "os/exec"
    "os"
    "io/ioutil"
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

type JobStatus uint8
const (
    JobGood     JobStatus = 0
    JobFailed             = 1
)

type TimePred func(int) bool

type Job struct {
    // params
    Name        string
    Min         TimePred
    Hour        TimePred
    Mday        TimePred
    Mon         TimePred
    Wday        TimePred
    Cmd         string
    
    // other params
    stdoutLogger *log.Logger
    stderrLogger *log.Logger
    
    // dynamic shit
    Status      JobStatus
    LastRunTime time.Time
}

func NewJob(name string, cmd string) *Job {
    job := &Job{Name: name, Cmd: cmd, Status: JobGood}
    job.Min = func (i int) bool { return true }
    job.Hour = func (i int) bool { return true }
    job.Mday = func (i int) bool { return true }
    job.Mon = func (i int) bool { return true }
    job.Wday = func (i int) bool { return true }
    job.stdoutLogger = log.New(os.Stdout, name + " ", log.LstdFlags)
    job.stderrLogger = log.New(os.Stderr, name + " ", log.LstdFlags)
    return job
}

type RunLogEntry struct {
    Job     *Job
    Time    time.Time
    Result  JobStatus
}

type JobManager struct {
    Jobs        []*Job
    RunLog      []RunLogEntry
    Shell       string
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

func shouldRun(now time.Time, job *Job) bool {
    // match minute
    if !job.Min(now.Minute()) {
        return false
    } else if !job.Hour(now.Hour()) {
        return false
    } else if !job.Mday(now.Day()) {
        return false
    } else if !job.Mon(monthToInt(now.Month())) {
        return false
    } else if !job.Wday(weekdayToInt(now.Weekday())) {
        return false
    } else {
        return true
    }
}

func (m *JobManager) JobsToRun() []*Job {
    jobs := make([]*Job, 0)
    now := time.Now()
    for _, job := range m.Jobs {
        if job.Status == JobGood && shouldRun(now, job) {
            jobs = append(jobs, job)
        }
    }
    return jobs
}

func (m *JobManager) Go() {
    // make updater thread
    resultChan := make(chan *RunRec)
    go m.UpdaterThread(resultChan)
    
    // run jobs
    period := time.Duration(5 * time.Second)
    for {
        jobs := m.JobsToRun()
        for _, job := range jobs {
            go m.Run(job, resultChan)
        }
        
        time.Sleep(period)
    }
    
    close(resultChan)
}

type RunRec struct {
    Job         *Job
    RunTime     time.Time
    NewStatus   JobStatus
    Stdout      string
    Stderr      string
    Err         *JobberError
}

func (m *JobManager) Run(job *Job, c chan *RunRec) {
    log.Println("Running " + job.Name)
    rec := &RunRec{Job: job, RunTime: time.Now(), NewStatus: JobGood}
    
    var cmd *exec.Cmd = exec.Command(m.Shell, "-c", job.Cmd)
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        rec.Err = &JobberError{"Failed to get pipe to stdout.", err}
        c <- rec
        return
    }
    stderr, err := cmd.StderrPipe()
    if err != nil {
        rec.Err = &JobberError{"Failed to get pipe to stderr.", err}
        c <- rec
        return
    }
    
    // start cmd
    if err := cmd.Start(); err != nil {
        /* Failed to start command. */
        rec.Stderr = "Failed to run: " + err.Error()
        rec.NewStatus = JobFailed
        c <- rec
        return
    }
    
    // read output
    stdoutBytes, err := ioutil.ReadAll(stdout)
    if err != nil {
        rec.Err = &JobberError{"Failed to read stdout.", err}
        c <- rec
        return
    }
    rec.Stdout = string(stdoutBytes)
    stderrBytes, err := ioutil.ReadAll(stderr)
    if err != nil {
        rec.Err = &JobberError{"Failed to read stderr.", err}
        c <- rec
        return
    }
    rec.Stderr = string(stderrBytes)
    
    // finish execution
    err = cmd.Wait()
    if err != nil {
        switch err := err.(type) {
            case *exec.ExitError: 
                rec.NewStatus = JobFailed
                c <- rec
                return
            
            default:
                rec.Err = &JobberError{"Error", err}
                rec.NewStatus = JobFailed
                c <- rec
                return
        }
    } else {
        c <- rec
        return
    }
}

func (m *JobManager) UpdaterThread(c chan *RunRec) {
    for {
        var rec *RunRec
        rec, ok := <-c
        if !ok {
            /* Channel is closed. */
            return
        } else {
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
            m.RunLog = append(m.RunLog, RunLogEntry{rec.Job, rec.RunTime, rec.Job.Status})
        }
    }
}
