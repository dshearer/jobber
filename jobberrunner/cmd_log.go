package main

import (
	"github.com/dshearer/jobber/common"
)

func (self *JobManager) doLogCmd(cmd common.LogCmd) common.ICmdResp {
	common.Logger.Printf("Got cmd 'log'\n")

	// make log list
	var logDescs []common.LogDesc
	entries, err := self.jfile.Prefs.RunLog.GetAll()
	if err != nil {
		return common.NewErrorCmdResp(err)
	}
	for _, l := range entries {
		logDesc := common.LogDesc{
			Time:      l.Time,
			Job:       l.JobName,
			Succeeded: l.Succeeded,
			Result:    l.Result.String(),
		}
		logDescs = append(logDescs, logDesc)
	}

	// make response
	return common.LogCmdResp{Logs: logDescs}
}
