package main

import (
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/ipc"
)

func (self *JobManager) doCatCmd(cmd ipc.CatCmd) ipc.ICmdResp {
	common.Logger.Printf("Got cmd 'cat'\n")

	// find job
	job, ok := self.jfile.Jobs[cmd.Job]
	if !ok {
		return ipc.NewErrorCmdResp(&common.Error{What: "No such job."})
	}

	// make response
	return ipc.CatCmdResp{Result: job.Cmd}
}
