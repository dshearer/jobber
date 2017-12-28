package main

import (
	"fmt"
	"github.com/dshearer/jobber/common"
)

func (self *JobManager) doListJobsCmd(cmd common.ListJobsCmd) {
	defer close(cmd.RespChan)

	// make job list
	common.Logger.Printf("Got list jobs cmd\n")
	jobDescs := make([]common.JobDesc, 0)
	for _, j := range self.jfile.Jobs {
		jobDesc := common.JobDesc{
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
			NotifyOnSuccess: j.NotifyOnSuccess,
			NotifyOnErr:     j.NotifyOnError,
			NotifyOnFail:    j.NotifyOnFailure,
			ErrHandler:      j.ErrorHandler.String(),
		}
		if j.Paused {
			jobDesc.Status += " (Paused)"
			jobDesc.NextRunTime = nil
		}

		jobDescs = append(jobDescs, jobDesc)
	}

	// make response
	cmd.RespChan <- &common.ListJobsCmdResp{Jobs: jobDescs}
}
