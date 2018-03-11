package main

import (
	"fmt"
	"os/user"

	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/ipc"
	"github.com/dshearer/jobber/jobfile"
)

func (self *JobManager) doSetJobCmd(cmd ipc.SetJobCmd) ipc.ICmdResp {
	common.Logger.Println("Got command 'set job'")

	// get current user
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current user: %v", err))
	}

	// make job
	newJob, err := jobfile.JobRawToJob(cmd.Job, usr, self.jfile.Prefs)
	if err != nil {
		return ipc.NewErrorCmdResp(err)
	}

	// make new jobfile
	newJobs := make(map[string]*jobfile.Job)
	for currJobName, currJob := range self.jfile.Jobs {
		newJobs[currJobName] = currJob
	}
	newJobs[newJob.Name] = newJob
	newJobfile := jobfile.JobFile{
		Prefs: self.jfile.Prefs,
		Jobs:  newJobs,
	}

	// install it
	self.replaceCurrJobfile(&newJobfile)

	common.Logger.Println("Finished command 'set job'")
	return ipc.SetJobCmdResp{Ok: true}
}
