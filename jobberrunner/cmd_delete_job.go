package main

import (
	"github.com/dshearer/jobber/ipc"
	"github.com/dshearer/jobber/jobfile"
)

func (self *JobManager) doDeleteJobCmd(cmd ipc.DeleteJobCmd) ipc.ICmdResp {
	// make new raw jobfile
	newJobs := make(map[string]jobfile.JobRaw)
	for currJobName, currJob := range self.jfile.Raw.Jobs {
		if currJobName == cmd.Job {
			continue
		}
		newJobs[currJobName] = currJob
	}
	newJobfile := jobfile.JobFileRaw{
		Prefs: self.jfile.Raw.Prefs,
		Jobs:  newJobs,
	}

	// install it
	self.replaceCurrJobfile(&newJobfile)

	return ipc.DeleteJobCmdResp{Ok: true}
}
