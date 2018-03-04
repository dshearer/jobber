package main

import (
	"github.com/dshearer/jobber/common"
)

func (self *JobManager) doTestCmd(cmd common.TestCmd) common.ICmdResp {
	common.Logger.Printf("Got cmd 'test'\n")

	// find job
	job := self.findJob(cmd.Job)
	if job == nil {
		return common.NewErrorCmdResp(&common.Error{What: "No such job."})
	}

	// run the job in this thread
	runRec := RunJob(job, self.Shell, true)

	// make response
	if runRec.Err != nil {
		return common.NewErrorCmdResp(runRec.Err)
	}

	return common.TestCmdResp{Result: runRec.Describe()}
}
