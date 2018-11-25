package main

import (
	"fmt"

	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/ipc"
	"github.com/dshearer/jobber/jobfile"
)

func (self *JobManager) doResumeCmd(cmd ipc.ResumeCmd) ipc.ICmdResp {
	// look up jobs to resume
	var jobsToResume []*jobfile.Job
	if len(cmd.Jobs) == 0 {
		for _, job := range self.jfile.Jobs {
			jobsToResume = append(jobsToResume, job)
		}
	} else {
		for _, jobName := range cmd.Jobs {
			job, ok := self.jfile.Jobs[jobName]
			if !ok {
				msg := fmt.Sprintf("No such job: %v", jobName)
				return ipc.NewErrorCmdResp(&common.Error{What: msg})
			}
			jobsToResume = append(jobsToResume, job)
		}
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
	return ipc.ResumeCmdResp{NumResumed: numResumed}
}
