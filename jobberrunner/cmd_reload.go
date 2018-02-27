package main

import (
	"github.com/dshearer/jobber/common"
)

func (self *JobManager) doReloadCmd(cmd common.ReloadCmd) {
	common.Logger.Printf("Got cmd 'reload'\n")

	defer close(cmd.RespChan)

	// read job file
	err := self.loadJobfile()
	if err != nil {
		cmd.RespChan <- &common.ReloadCmdResp{Err: err}
	} else {
		common.Logger.Printf("%v", self.jfile.Prefs.String())
		cmd.RespChan <- &common.ReloadCmdResp{
			NumJobs: len(self.jfile.Jobs),
		}
	}
}
