package main

import (
	"fmt"
	"os/user"

	"github.com/dshearer/jobber/ipc"
	"github.com/dshearer/jobber/jobfile"
)

func (self *JobManager) doSetJobCmd(cmd ipc.SetJobCmd) ipc.ICmdResp {
	// get current user
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current user: %v", err))
	}

	// make job
	newJob := jobfile.NewJob()
	if err := cmd.Job.ToJob(usr, &newJob); err != nil {
		return ipc.NewErrorCmdResp(err)
	}
	newJob.Name = cmd.Job.Name

	// make new jobfile
	newJobs := make(map[string]*jobfile.Job)
	for currJobName, currJob := range self.jfile.Jobs {
		newJobs[currJobName] = currJob
	}
	newJobs[newJob.Name] = &newJob
	newJobfile := jobfile.JobFile{
		Prefs: self.jfile.Prefs,
		Jobs:  newJobs,
	}

	// install it
	self.replaceCurrJobfile(&newJobfile)

	return ipc.SetJobCmdResp{Ok: true}
}
