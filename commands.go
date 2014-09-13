package jobber

import (
    "io"
)

type ICmd interface {
    RespChan() chan ICmdResp
}

type ICmdResp interface {
    IsError() bool
}

type SuccessCmdResp struct {
    Details string
}

func (r SuccessCmdResp) IsError() bool {
    return false
}

type ErrorCmdResp struct {
    Error error
}

func (r ErrorCmdResp) IsError() bool {
    return true
}

type ReloadCmd struct {
    JobFile io.Reader
    respChan chan ICmdResp
}

func (c *ReloadCmd) RespChan() chan ICmdResp {
    return c.respChan
}

func (c ReloadCmd) String() string {
    return "ReloadCmd"
}

type ListJobsCmd struct {
    respChan chan ICmdResp
}

func (c *ListJobsCmd) RespChan() chan ICmdResp {
    return c.respChan
}

func (c ListJobsCmd) String() string {
    return "ListJobsCmd"
}

type ListHistoryCmd struct {
    respChan chan ICmdResp
}

func (c *ListHistoryCmd) RespChan() chan ICmdResp {
    return c.respChan
}

func (c ListHistoryCmd) String() string {
    return "ListHistoryCmd"
}

type StopCmd struct {
    respChan chan ICmdResp
}

func (c *StopCmd) RespChan() chan ICmdResp {
    return c.respChan
}

func (c StopCmd) String() string {
    return "StopCmd"
}
