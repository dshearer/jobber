package main

import (
	"github.com/dshearer/jobber/common"
)

func (self *JobManager) doCatCmd(cmd common.CatCmd) {
	defer close(cmd.RespChan)

	// find job
	job := self.findJob(cmd.Job)
	if job == nil {
		cmd.RespChan <- &common.CatCmdResp{
			Err: &common.Error{What: "No such job."},
		}
		return
	}

	// make response
	cmd.RespChan <- &common.CatCmdResp{Result: job.Cmd}
}
