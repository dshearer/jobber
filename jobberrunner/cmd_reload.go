package main

import (
	"github.com/dshearer/jobber/common"
	"os"
)

func (self *JobManager) doReloadCmd(cmd common.ReloadCmd) {
	defer close(cmd.RespChan)

	// read job file
	newJfile, err := self.loadJobFile()
	if err != nil && !os.IsNotExist(err) {
		cmd.RespChan <- &common.ReloadCmdResp{Err: err}
		close(cmd.RespChan)
		return
	}

	// stop job-runner thread and wait for current runs to end
	self.jobRunner.Cancel()
	for rec := range self.jobRunner.RunRecChan() {
		self.handleRunRec(rec)
	}
	self.jobRunner.Wait()

	// set new job file
	self.jfile = newJfile

	// restart job-runner thread
	self.jobRunner.Start(self.jfile.Jobs, self.Shell,
		self.mainThreadCtx)

	// make response
	cmd.RespChan <- &common.ReloadCmdResp{
		NumJobs: len(self.jfile.Jobs),
	}
}
