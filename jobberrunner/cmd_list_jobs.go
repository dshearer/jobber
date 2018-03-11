package main

import (
	"fmt"

	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/ipc"
	"github.com/dshearer/jobber/jobfile"
)

func (self *JobManager) doListJobsCmd(cmd ipc.ListJobsCmd) ipc.ICmdResp {
	common.Logger.Printf("Got cmd 'list'\n")

	// make job list
	jobDescs := make([]ipc.JobDesc, 0)
	for _, j := range self.jfile.Jobs {
		var stdoutDir *string
		var stderrDir *string
		if handler, ok := j.StdoutHandler.(jobfile.FileJobOutputHandler); ok {
			stdoutDir = &handler.Where
		}
		if handler, ok := j.StderrHandler.(jobfile.FileJobOutputHandler); ok {
			stderrDir = &handler.Where
		}

		jobDesc := ipc.JobDesc{
			Name:   j.Name,
			Status: j.Status.String(),
			Schedule: fmt.Sprintf(
				"%v %v %v %v %v %v",
				j.FullTimeSpec.Sec,
				j.FullTimeSpec.Min,
				j.FullTimeSpec.Hour,
				j.FullTimeSpec.Mday,
				j.FullTimeSpec.Mon,
				j.FullTimeSpec.Wday),
			NextRunTime:     j.NextRunTime,
			NotifyOnSuccess: j.NotifyOnSuccess.String(),
			NotifyOnErr:     j.NotifyOnError.String(),
			NotifyOnFail:    j.NotifyOnFailure.String(),
			ErrHandler:      j.ErrorHandler.String(),
			StdoutDir:       stdoutDir,
			StderrDir:       stderrDir,
		}
		if j.Paused {
			jobDesc.Status += " (Paused)"
			jobDesc.NextRunTime = nil
		}

		jobDescs = append(jobDescs, jobDesc)
	}

	// make response
	return ipc.ListJobsCmdResp{Jobs: jobDescs}
}
