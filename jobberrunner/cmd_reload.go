package main

import (
	"github.com/dshearer/jobber/ipc"
)

func (self *JobManager) doReloadCmd(cmd ipc.ReloadCmd) ipc.ICmdResp {
	// read job file
	if err := self.loadJobfile(); err != nil {
		return ipc.NewErrorCmdResp(err)
	}

	return ipc.ReloadCmdResp{NumJobs: len(self.jfile.Jobs)}
}
