package common

import ()

type ICmd interface{}
type ICmdResp interface{}

type ReloadCmd struct{}
type ReloadCmdResp struct {
	NumJobs int
	Err     error
}
