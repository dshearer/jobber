package main

import (
	"github.com/dshearer/jobber/common"
)

func (self *JobManager) doTestCmd(cmd common.TestCmd) {
	defer close(cmd.RespChan)

	// find job
	job := self.findJob(cmd.Job)
	if job == nil {
		cmd.RespChan <- &common.TestCmdResp{
			Err: &common.Error{What: "No such job."},
		}
		return
	}

	// run the job in this thread
	runRec := RunJob(job, self.Shell, true)

	// make response
	if runRec.Err == nil {
		cmd.RespChan <- &common.TestCmdResp{Result: runRec.Describe()}
	} else {
		cmd.RespChan <- &common.TestCmdResp{Err: runRec.Err}
	}
}
