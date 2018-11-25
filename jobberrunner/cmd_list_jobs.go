package main

import (
	"fmt"
	"strings"

	"github.com/dshearer/jobber/ipc"
	"github.com/dshearer/jobber/jobfile"
)

func resultSinksString(sinks []jobfile.ResultSink) string {
	var strs []string
	for _, sink := range sinks {
		strs = append(strs, sink.String())
	}
	return strings.Join(strs, ",")
}

func (self *JobManager) doListJobsCmd(cmd ipc.ListJobsCmd) ipc.ICmdResp {
	// make job list
	jobDescs := make([]ipc.JobDesc, 0)
	for _, j := range self.jfile.Jobs {
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
			NotifyOnSuccess: resultSinksString(j.NotifyOnSuccess),
			NotifyOnErr:     resultSinksString(j.NotifyOnError),
			NotifyOnFail:    resultSinksString(j.NotifyOnFailure),
			ErrHandler:      j.ErrorHandler.String(),
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
