package main

import (
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
)

func (self *JobManager) doPauseCmd(cmd common.PauseCmd) {
	common.Logger.Printf("Got cmd 'pause'\n")

	defer close(cmd.RespChan)

	// look up jobs to pause
	var jobsToPause []*jobfile.Job
	if len(cmd.Jobs) > 0 {
		var err error
		jobsToPause, err = self.findJobs(cmd.Jobs)
		if err != nil {
			cmd.RespChan <- &common.PauseCmdResp{Err: err}
			return
		}
	} else {
		jobsToPause = self.jfile.Jobs
	}

	// pause them
	amtPaused := 0
	for _, job := range jobsToPause {
		if !job.Paused {
			job.Paused = true
			amtPaused += 1
		}
	}

	// make response
	cmd.RespChan <- &common.PauseCmdResp{AmtPaused: amtPaused}
}
