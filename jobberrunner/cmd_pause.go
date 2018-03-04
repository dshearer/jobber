package main

import (
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
)

func (self *JobManager) doPauseCmd(cmd common.PauseCmd) common.ICmdResp {
	common.Logger.Printf("Got cmd 'pause'\n")

	// look up jobs to pause
	var jobsToPause []*jobfile.Job
	if len(cmd.Jobs) > 0 {
		var err error
		jobsToPause, err = self.findJobs(cmd.Jobs)
		if err != nil {
			return common.NewErrorCmdResp(err)
		}
	} else {
		jobsToPause = self.jfile.Jobs
	}

	// pause them
	numPaused := 0
	for _, job := range jobsToPause {
		if !job.Paused {
			job.Paused = true
			numPaused += 1
		}
	}

	// make response
	return common.PauseCmdResp{NumPaused: numPaused}
}
