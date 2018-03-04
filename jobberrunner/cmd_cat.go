package main

import (
	"github.com/dshearer/jobber/common"
)

func (self *JobManager) doCatCmd(cmd common.CatCmd) common.ICmdResp {
	common.Logger.Printf("Got cmd 'cat'\n")

	// find job
	job := self.findJob(cmd.Job)
	if job == nil {
		return common.NewErrorCmdResp(&common.Error{What: "No such job."})
	}

	// make response
	return common.CatCmdResp{Result: job.Cmd}
}
