package main

import (
)

type ICmd interface {
    RequestingUser() string
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
