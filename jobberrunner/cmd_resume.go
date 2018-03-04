package main

import (
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
)

func (self *JobManager) doResumeCmd(cmd common.ResumeCmd) common.ICmdResp {
	common.Logger.Printf("Got cmd 'resume'\n")

	// look up jobs to resume
	var jobsToResume []*jobfile.Job
	if len(cmd.Jobs) > 0 {
		var err error
		jobsToResume, err = self.findJobs(cmd.Jobs)
		if err != nil {
			return common.NewErrorCmdResp(err)
		}
	} else {
		jobsToResume = self.jfile.Jobs
	}

	// pause them
	numResumed := 0
	for _, job := range jobsToResume {
		if job.Paused {
			job.Paused = false
			numResumed += 1
		}
	}

	// make response
	return common.ResumeCmdResp{NumResumed: numResumed}
}
