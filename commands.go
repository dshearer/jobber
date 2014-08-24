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

type ListJobsCmd struct {
    respChan chan ICmdResp
}

func (c *ListJobsCmd) RespChan() chan ICmdResp {
    return c.respChan
}

type ListHistoryCmd struct {
    respChan chan ICmdResp
}

func (c *ListHistoryCmd) RespChan() chan ICmdResp {
    return c.respChan
}
