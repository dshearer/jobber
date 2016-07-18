package main

/* ICmd */

type ICmd interface {
    RequestingUser() string
    RespChan() chan ICmdResp
}

/* ICmdResp */

type ICmdResp interface {
    IsError() bool
}

/* SuccessCmdResp */

type SuccessCmdResp struct {
    Details string
}

func (r SuccessCmdResp) IsError() bool {
    return false
}

/* ErrorCmdResp */

type ErrorCmdResp struct {
    Error error
}

func (r ErrorCmdResp) IsError() bool {
    return true
}

/* ReloadCmd */

type ReloadCmd struct {
    user string
    respChan chan ICmdResp
    ForAllUsers bool
}

func (c *ReloadCmd) RespChan() chan ICmdResp {
    return c.respChan
}

func (c *ReloadCmd) RequestingUser() string {
    return c.user
}

func (c ReloadCmd) String() string {
    return "ReloadCmd"
}

/* ListJobsCmd */

type ListJobsCmd struct {
    user string
    respChan chan ICmdResp
    ForAllUsers bool
}

func (c *ListJobsCmd) RespChan() chan ICmdResp {
    return c.respChan
}

func (c *ListJobsCmd) RequestingUser() string {
    return c.user
}

func (c ListJobsCmd) String() string {
    return "ListJobsCmd"
}

/* ListHistoryCmd */

type ListHistoryCmd struct {
    user string
    respChan chan ICmdResp
    ForAllUsers bool
}

func (c *ListHistoryCmd) RespChan() chan ICmdResp {
    return c.respChan
}

func (c *ListHistoryCmd) RequestingUser() string {
    return c.user
}

func (c ListHistoryCmd) String() string {
    return "ListHistoryCmd"
}

/* StopCmd */

type StopCmd struct {
    user string
    respChan chan ICmdResp
}

func (c *StopCmd) RespChan() chan ICmdResp {
    return c.respChan
}

func (c *StopCmd) RequestingUser() string {
    return c.user
}

func (c StopCmd) String() string {
    return "StopCmd"
}

/* TestCmd */

type TestCmd struct {
    user string
    respChan chan ICmdResp
    job string
    jobUser string
}

func (c *TestCmd) RespChan() chan ICmdResp {
    return c.respChan
}

func (c *TestCmd) RequestingUser() string {
    return c.user
}

func (c TestCmd) String() string {
    return "TestCmd"
}

/* CatCmd */

type CatCmd struct {
    user string
    respChan chan ICmdResp
    job string
    jobUser string
}

func (c *CatCmd) RespChan() chan ICmdResp {
    return c.respChan
}

func (c *CatCmd) RequestingUser() string {
    return c.user
}

func (c CatCmd) String() string {
    return "CatCmd"
}

/* PauseCmd */

type PauseCmd struct {
    user string
    respChan chan ICmdResp
    jobs []string
}

func (c *PauseCmd) RespChan() chan ICmdResp {
    return c.respChan
}

func (c *PauseCmd) RequestingUser() string {
    return c.user
}

func (c PauseCmd) String() string {
    return "PauseCmd"
}

/* ResumeCmd */

type ResumeCmd struct {
    user string
    respChan chan ICmdResp
    jobs []string
}

func (c *ResumeCmd) RespChan() chan ICmdResp {
    return c.respChan
}

func (c *ResumeCmd) RequestingUser() string {
    return c.user
}

func (c ResumeCmd) String() string {
    return "ResumeCmd"
}
