package main

import (
	"github.com/dshearer/jobber/common"
)

func (self *JobManager) doReloadCmd(cmd common.ReloadCmd) common.ICmdResp {
	common.Logger.Printf("Got cmd 'reload'\n")

	// read job file
	if err := self.loadJobfile(); err != nil {
		return common.NewErrorCmdResp(err)
	}

	common.Logger.Printf("%v", self.jfile.Prefs.String())
	return common.ReloadCmdResp{NumJobs: len(self.jfile.Jobs)}
}
