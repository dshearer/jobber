package main

import (
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/ipc"
)

func (self *JobManager) doTestCmd(cmd ipc.TestCmd) ipc.ICmdResp {
	// find job
	job, ok := self.jfile.Jobs[cmd.Job]
	if !ok {
		return ipc.NewErrorCmdResp(&common.Error{What: "No such job."})
	}

	common.Logger.Printf("Trying job %v\n", job.Name)
	sockPath, err := self.testJobServer.Launch(job)
	if err != nil {
		return ipc.NewErrorCmdResp(err)
	}

	return ipc.TestCmdResp{UnixSocketPath: *sockPath}
}
