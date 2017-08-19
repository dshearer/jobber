package common

import (
	"time"
)

type ICmd interface{}
type ICmdResp interface{}

type ReloadCmd struct {
	Dummy    int // this is here just to make RPC work
	RespChan chan *ReloadCmdResp
}
type ReloadCmdResp struct {
	NumJobs int
	Err     error
}

type JobDesc struct {
	Name         string
	Status       string
	Schedule     string
	NextRunTime  *time.Time
	NotifyOnErr  bool
	NotifyOnFail bool
	ErrHandler   string
}

type ListJobsCmd struct {
	Dummy    int // this is here just to make RPC work
	RespChan chan *ListJobsCmdResp
}
type ListJobsCmdResp struct {
	Jobs []JobDesc
	Err  error
}

type LogDesc struct {
	Time      time.Time
	Job       string
	Succeeded bool
	Result    string
}

type LogCmd struct {
	Dummy    int // this is here just to make RPC work
	RespChan chan *LogCmdResp
}
type LogCmdResp struct {
	Logs []LogDesc
	Err  error
}

type TestCmd struct {
	Job      string
	RespChan chan *TestCmdResp
}
type TestCmdResp struct {
	Result string
	Err    error
}

type CatCmd struct {
	Job      string
	RespChan chan *CatCmdResp
}
type CatCmdResp struct {
	Result string
	Err    error
}

type PauseCmd struct {
	Jobs     []string
	RespChan chan *PauseCmdResp
}
type PauseCmdResp struct {
	AmtPaused int
	Err       error
}

type ResumeCmd struct {
	Jobs     []string
	RespChan chan *ResumeCmdResp
}
type ResumeCmdResp struct {
	AmtResumed int
	Err        error
}
