package jobfile

import (
	"fmt"

	"github.com/dshearer/jobber/common"
)

const (
	ErrorHandlerStopName     = "Stop"
	ErrorHandlerBackoffName  = "Backoff"
	ErrorHandlerContinueName = "Continue"

	MaxBackoffWait = 8
)

type ErrorHandler interface {
	Handle(job *Job)
	fmt.Stringer
}

type ContinueErrorHandler struct{}

func (self ContinueErrorHandler) Handle(job *Job) {
	job.Status = JobGood
}

func (self ContinueErrorHandler) String() string {
	return ErrorHandlerContinueName
}

type StopErrorHandler struct{}

func (self StopErrorHandler) Handle(job *Job) {
	job.Status = JobFailed
}

func (self StopErrorHandler) String() string {
	return ErrorHandlerStopName
}

type BackoffErrorHandler struct{}

func (self BackoffErrorHandler) Handle(job *Job) {
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
}

func (self BackoffErrorHandler) String() string {
	return ErrorHandlerBackoffName
}

func GetErrorHandler(name string) (ErrorHandler, error) {
	switch name {
	case ErrorHandlerStopName:
		return StopErrorHandler{}, nil
	case ErrorHandlerBackoffName:
		return BackoffErrorHandler{}, nil
	case ErrorHandlerContinueName:
		return ContinueErrorHandler{}, nil
	default:
		return nil, &common.Error{What: "Invalid error handler: " + name}
	}
}
