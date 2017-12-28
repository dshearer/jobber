package main

import (
	"github.com/dshearer/jobber/common"
)

func (self *JobManager) doLogCmd(cmd common.LogCmd) {
	defer close(cmd.RespChan)

	// make log list
	var logDescs []common.LogDesc
	entries, err := self.jfile.Prefs.RunLog.GetFromIndex()
	if err != nil {
		cmd.RespChan <- &common.LogCmdResp{Err: err}
		close(cmd.RespChan)
		return
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
	cmd.RespChan <- &common.LogCmdResp{Logs: logDescs}
}
