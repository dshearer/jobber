package main

import (
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/ipc"
)

func (self *JobManager) doSetJobCmd(cmd ipc.SetJobCmd) ipc.ICmdResp {
	common.Logger.Println("doSetJobCmd")
	// set job
	rawDup := self.jfile.Raw.Dup()
	common.Logger.Printf("After dup: %v", rawDup)
	rawDup.Jobs[cmd.Job.Name] = cmd.Job.JobV3Raw

	// install it
	common.Logger.Println("Before replaceCurrJobfile")
	self.replaceCurrJobfile(rawDup)
	common.Logger.Println("After replaceCurrJobfile")

	return ipc.SetJobCmdResp{Ok: true}
}
