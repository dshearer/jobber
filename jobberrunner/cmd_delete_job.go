package main

import (
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/ipc"
	"github.com/dshearer/jobber/jobfile"
)

func (self *JobManager) doDeleteJobCmd(cmd ipc.DeleteJobCmd) ipc.ICmdResp {
	common.Logger.Println("Got command 'delete job'")

	// make new jobfile
	newJobs := make(map[string]*jobfile.Job)
	for currJobName, currJob := range self.jfile.Jobs {
		if currJobName == cmd.Job {
			continue
		}
		newJobs[currJobName] = currJob
	}
	newJobfile := jobfile.JobFile{
		Prefs: self.jfile.Prefs,
		Jobs:  newJobs,
	}

	// install it
	self.replaceCurrJobfile(&newJobfile)

	return ipc.DeleteJobCmdResp{Ok: true}
}
