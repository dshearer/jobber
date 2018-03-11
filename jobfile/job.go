package jobfile

import (
	"fmt"
	"time"

	"github.com/dshearer/jobber/common"
)

type JobStatus uint8

const (
	JobGood JobStatus = iota
	JobFailed
	JobBackoff
)

var JobStatuses = [...]JobStatus{
	JobGood,
	JobFailed,
	JobBackoff,
}

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

type Job struct {
	// params
	Name            string
	Cmd             string
	FullTimeSpec    FullTimeSpec
	User            string
	ErrorHandler    ErrorHandler
	NotifyOnError   RunRecNotifier
	NotifyOnFailure RunRecNotifier
	NotifyOnSuccess RunRecNotifier
	NextRunTime     *time.Time

	// output handling
	StdoutHandler JobOutputHandler
	StderrHandler JobOutputHandler

	// backoff after errors
	backoffLevel int
	skipsLeft    int

	// other dynamic stuff
	Status      JobStatus
	LastRunTime time.Time
	Paused      bool
}

func (j *Job) String() string {
	return j.Name
}

func NewJob(name string, cmd string, username string) *Job {
	return &Job{
		Name:            name,
		Cmd:             cmd,
		Status:          JobGood,
		User:            username,
		ErrorHandler:    ContinueErrorHandler{},
		NotifyOnError:   NopRunRecNotifier{},
		NotifyOnFailure: MailRunRecNotifier{},
		NotifyOnSuccess: NopRunRecNotifier{},
	}
}

type RunRec struct {
	Job       *Job
	RunTime   time.Time
	NewStatus JobStatus
	Stdout    *[]byte
	Stderr    *[]byte
	Succeeded bool
	Err       *common.Error
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
