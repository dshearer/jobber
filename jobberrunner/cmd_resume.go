package main

import (
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
)

func (self *JobManager) doResumeCmd(cmd common.ResumeCmd) {
	common.Logger.Printf("Got cmd 'resume'\n")

	defer close(cmd.RespChan)

	// look up jobs to resume
	var jobsToResume []*jobfile.Job
	if len(cmd.Jobs) > 0 {
		var err error
		jobsToResume, err = self.findJobs(cmd.Jobs)
		if err != nil {
			cmd.RespChan <- &common.ResumeCmdResp{Err: err}
			return
		}
	} else {
		jobsToResume = self.jfile.Jobs
	}

	// pause them
	amtResumed := 0
	for _, job := range jobsToResume {
		if job.Paused {
			job.Paused = false
			amtResumed += 1
		}
	}

	// make response
	cmd.RespChan <- &common.ResumeCmdResp{AmtResumed: amtResumed}
}
