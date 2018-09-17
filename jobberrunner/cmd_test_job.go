package main

import (
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/ipc"
)

func (self *JobManager) doTestCmd(cmd ipc.TestCmd) ipc.ICmdResp {
	common.Logger.Printf("Got cmd 'test'\n")

	// find job
	job, ok := self.jfile.Jobs[cmd.Job]
	if !ok {
		return ipc.NewErrorCmdResp(&common.Error{What: "No such job."})
	}

	// run the job in this thread
	runRec := RunJob(job, self.Shell, true)

	// make response
	if runRec.Err != nil {
		return ipc.NewErrorCmdResp(runRec.Err)
	}

	return ipc.TestCmdResp{Result: runRec.Describe()}
}
