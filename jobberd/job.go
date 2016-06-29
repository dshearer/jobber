package main

import (
	"fmt"
	"github.com/dshearer/jobber/Godeps/_workspace/src/golang.org/x/net/context"
	"log"
	"time"
)

const (
	MaxBackoffWait = 8
)

type JobStatus uint8

const (
	JobGood    JobStatus = 0
	JobFailed            = 1
	JobBackoff           = 2
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

type TimeSpec interface {
	String() string
	Satisfied(int) bool
}

type FullTimeSpec struct {
	Sec  TimeSpec
	Min  TimeSpec
	Hour TimeSpec
	Mday TimeSpec
	Mon  TimeSpec
	Wday TimeSpec
}

func (self FullTimeSpec) String() string {
    return fmt.Sprintf("%v %v %v %v %v %v",
                       self.Sec,
                       self.Min,
                       self.Hour,
                       self.Mday,
                       self.Mon,
                       self.Wday)
}

const (
	ErrorHandlerStopName     = "Stop"
	ErrorHandlerBackoffName  = "Backoff"
	ErrorHandlerContinueName = "Continue"
)

type ErrorHandler struct {
	apply func(*Job)
	desc  string
}

func (h ErrorHandler) String() string {
	return h.desc
}

var ErrorHandlerStop = ErrorHandler{
	apply: func(job *Job) { job.Status = JobFailed },
	desc:  ErrorHandlerStopName,
}

var ErrorHandlerBackoff = ErrorHandler{
	apply: func(job *Job) {
		/*
		   The job has just had an error.  We'll handle
		   it by skipping the next N consecutive chances
		   to run.  If this is the first time it has failed
		   (i.e., job.Status == JobGood), then N will be 1;
		   otherwise, N will be twice the amount of chances
		   previously skipped.  If N is greater than
		   MaxBackoffWait, however, we mark this job as
		   "Failed" and don't run it again.

		   We use two variables: backoffLevel and skipsLeft.
		   backoffLevel is the amount of chances to skip
		   at this time.  skipsLeft is the amount of chances
		   we have skipped so far.  When a job is in state
		   JobBackoff, Job.ShouldRun decrements skipsLeft
		   and returns false if skipsLeft > 0, true otherwise.
		*/

		if job.Status == JobGood {
			job.Status = JobBackoff
			job.backoffLevel = 1
		} else {
			job.backoffLevel *= 2
		}
		if job.backoffLevel > MaxBackoffWait {
			// give up
			job.Status = JobFailed
			job.backoffLevel = 0
			job.skipsLeft = 0
		} else {
			job.skipsLeft = job.backoffLevel
		}
	},
	desc: ErrorHandlerBackoffName,
}

var ErrorHandlerContinue = ErrorHandler{
	apply: func(job *Job) { job.Status = JobGood },
	desc:  ErrorHandlerContinueName,
}

type Job struct {
	// params
	Name            string
	Cmd             string
	FullTimeSpec    FullTimeSpec
	User            string
	ErrorHandler    *ErrorHandler
	NotifyOnError   bool
	NotifyOnFailure bool
	NextRunTime     *time.Time

	// other params
	stdoutLogger *log.Logger
	stderrLogger *log.Logger

	// dynamic shit
	Status      JobStatus
	LastRunTime time.Time

	// backoff after errors
	backoffLevel int
	skipsLeft    int
}

func (j *Job) String() string {
	return j.Name
}

func NewJob(name string, cmd string, username string) *Job {
	job := &Job{Name: name, Cmd: cmd, Status: JobGood, User: username}
	job.ErrorHandler = &ErrorHandlerContinue
	job.NotifyOnError = false
	job.NotifyOnFailure = true
	return job
}

type RunRec struct {
	Job       *Job
	RunTime   time.Time
	NewStatus JobStatus
	Stdout    *[]byte
	Stderr    *[]byte
	Succeeded bool
	Err       *JobberError
}

func (rec *RunRec) Describe() string {
	var summary string
	if rec.Succeeded {
		summary = fmt.Sprintf("Job \"%v\" succeeded.", rec.Job.Name)
	} else {
		summary = fmt.Sprintf("Job \"%v\" failed.", rec.Job.Name)
	}
	stdoutStr, _ := SafeBytesToStr(*rec.Stdout)
	stderrStr, _ := SafeBytesToStr(*rec.Stderr)
	return fmt.Sprintf("%v\r\nNew status: %v.\r\n\r\nStdout:\r\n%v\r\n\r\nStderr:\r\n%v",
		summary, rec.Job.Status, stdoutStr, stderrStr)
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
	rec.Stdout = &sudoResult.Stdout
	rec.Stderr = &sudoResult.Stderr

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

func (job *Job) ShouldRun() bool {
	switch job.Status {
	case JobFailed:
		return false

	case JobBackoff:
		job.skipsLeft--
		return job.skipsLeft <= 0

	default:
		return true
	}
}
