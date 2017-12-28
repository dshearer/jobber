package main

import (
	"github.com/dshearer/jobber/common"
)

func (self *JobManager) doReloadCmd(cmd common.ReloadCmd) {
	common.Logger.Println("Got reload cmd")
	defer close(cmd.RespChan)

	// read job file
	err := self.loadJobfile()
	if err != nil {
		cmd.RespChan <- &common.ReloadCmdResp{Err: err}
	} else {
		cmd.RespChan <- &common.ReloadCmdResp{
			NumJobs: len(self.jfile.Jobs),
		}
	}
}
