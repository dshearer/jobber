package main

import (
	"fmt"

	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/ipc"
	"github.com/dshearer/jobber/jobfile"
)

func (self *JobManager) doPauseCmd(cmd ipc.PauseCmd) ipc.ICmdResp {
	common.Logger.Printf("Got cmd 'pause'\n")

	// look up jobs to pause
	var jobsToPause []*jobfile.Job
	if len(cmd.Jobs) == 0 {
		for _, job := range self.jfile.Jobs {
			jobsToPause = append(jobsToPause, job)
		}
	} else {
		for _, jobName := range cmd.Jobs {
			job, ok := self.jfile.Jobs[jobName]
			if !ok {
				msg := fmt.Sprintf("No such job: %v", jobName)
				return ipc.NewErrorCmdResp(&common.Error{What: msg})
			}
			jobsToPause = append(jobsToPause, job)
		}
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
	return ipc.PauseCmdResp{NumPaused: numPaused}
}
