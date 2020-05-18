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
	NotifyOnError   []ResultSink
	NotifyOnFailure []ResultSink
	NotifyOnSuccess []ResultSink

	// backoff after errors
	backoffLevel int
	skipsLeft    int

	// other dynamic stuff
	NextRunTime *time.Time
	Status      JobStatus
	LastRunTime time.Time
	Paused      bool
}

func (j *Job) String() string {
	return j.Name
}

func NewJob() Job {
	return Job{
		Status:          JobGood,
		ErrorHandler:    ContinueErrorHandler{},
		NotifyOnError:   nil,
		NotifyOnFailure: nil,
		NotifyOnSuccess: nil,
	}
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

const RunRecOutputMaxLen = 1 << 20

type RunRec struct {
	Job       *Job
	RunTime   time.Time
	NewStatus JobStatus
	Stdout    []byte
	Stderr    []byte
	Fate      common.SubprocFate
	ExecTime  time.Duration
	Err       error
}

func (rec *RunRec) Describe() string {
	var summary string
	switch rec.Fate {
	case common.SubprocFateSucceeded:
		summary = fmt.Sprintf("Job \"%v\" succeeded.", rec.Job.Name)
		break
	case common.SubprocFateFailed:
		summary = fmt.Sprintf("Job \"%v\" failed.", rec.Job.Name)
		break
	case common.SubprocFateCancelled:
		summary = fmt.Sprintf("Job \"%v\" cancelled.", rec.Job.Name)
		break
	default:
		panic("Unknown subproc fate")
	}
	stdoutStr, _ := SafeBytesToStr(rec.Stdout)
	stderrStr, _ := SafeBytesToStr(rec.Stderr)
	return fmt.Sprintf("%v\r\nNew status: %v.\r\n\r\nStdout:\r\n%v\r\n\r\nStderr:\r\n%v",
		summary, rec.Job.Status, stdoutStr, stderrStr)
}
