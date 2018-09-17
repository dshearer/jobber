package main

import (
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/ipc"
)

func (self *JobManager) doReloadCmd(cmd ipc.ReloadCmd) ipc.ICmdResp {
	common.Logger.Printf("Got cmd 'reload'\n")

	// read job file
	if err := self.loadJobfile(); err != nil {
		return ipc.NewErrorCmdResp(err)
	}

	common.Logger.Printf("%v", self.jfile.Prefs.String())
	return ipc.ReloadCmdResp{NumJobs: len(self.jfile.Jobs)}
}
