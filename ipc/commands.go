package ipc

import (
	"time"

	"github.com/dshearer/jobber/jobfile"
)

type ICmd interface{}

type ICmdResp interface {
	Error() error
}

type errorCmdResp struct {
	err error
}

func (self errorCmdResp) Error() error {
	return self.err
}

func NewErrorCmdResp(err error) ICmdResp {
	return errorCmdResp{err: err}
}

type nonErrorCmdResp struct{}

func (self nonErrorCmdResp) Error() error {
	return nil
}

type ReloadCmd struct{}

type ReloadCmdResp struct {
	NumJobs int `json:"numJobs"`
	nonErrorCmdResp
}

type JobDesc struct {
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	Schedule        string     `json:"schedule"`
	NextRunTime     *time.Time `json:"nextRunTime"`
	NotifyOnSuccess string     `json:"notifyOnSuccess"`
	NotifyOnErr     string     `json:"notifyOnError"`
	NotifyOnFail    string     `json:"notifyOnFailure"`
	ErrHandler      string     `json:"errHandler"`
}

type ListJobsCmd struct{}

type ListJobsCmdResp struct {
	Jobs []JobDesc `json:"jobs"`
	nonErrorCmdResp
}

type LogDesc struct {
	Time      time.Time     `json:"time"`
	Job       string        `json:"job"`
	Succeeded bool          `json:"succeeded"` // deprecated
	Fate      string        `json:"fate"`
	ExecTime  time.Duration `json:"exectime"`
	Result    string        `json:"result"`
}

type LogCmd struct{}

type LogCmdResp struct {
	Logs []LogDesc `json:"logs"`
	nonErrorCmdResp
}

type TestCmd struct {
	Job string `json:"job"`
}

type TestCmdResp struct {
	UnixSocketPath string `json:"unixSocketPath"`
	nonErrorCmdResp
}

type CatCmd struct {
	Job string `json:"job"`
}

type CatCmdResp struct {
	Result string `json:"result"`
	nonErrorCmdResp
}

type PauseCmd struct {
	Jobs []string `json:"jobs"`
}

type PauseCmdResp struct {
	NumPaused int `json:"numPaused"`
	nonErrorCmdResp
}

type ResumeCmd struct {
	Jobs []string `json:"jobs"`
}

type ResumeCmdResp struct {
	NumResumed int `json:"numResumed"`
	nonErrorCmdResp
}

type InitCmd struct{}

type InitCmdResp struct {
	JobfilePath string `json:"jobfilePath"`
	nonErrorCmdResp
}

type JobV3RawWithName struct {
	jobfile.JobV3Raw
	Name string
}

type SetJobCmd struct {
	Job JobV3RawWithName `json:"job"`
}

type SetJobCmdResp struct {
	Ok bool `json:"ok"` // just to make IPC work
	nonErrorCmdResp
}

type DeleteJobCmd struct {
	Job string `json:"job"`
}

type DeleteJobCmdResp struct {
	Ok bool `json:"ok"` // just to make IPC work
	nonErrorCmdResp
}
