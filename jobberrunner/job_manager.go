package main

import (
	"fmt"
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
	"os/user"
	"strings"
)

type JobManager struct {
	jobfilePath   string
	launched      bool
	jfile         *jobfile.JobFile
	runLog        RunLog
	CmdChan       chan common.ICmd
	CmdRespChan   chan common.ICmdResp
	mainThreadCtx *common.NewContext
	jobRunner     *JobRunnerThread
	Shell         string
}

func NewJobManager(jobfilePath string) *JobManager {
	jm := JobManager{Shell: "/bin/sh"}
	jm.jobfilePath = jobfilePath
	jm.jobRunner = NewJobRunnerThread()
	jm.runLog = NewMemOnlyRunLog(100)
	return &jm
}

func (self *JobManager) Launch() error {
	if self.launched {
		return &common.Error{"Already launched.", nil}
	}

	// run main thread
	self.CmdChan = make(chan common.ICmd)
	self.CmdRespChan = make(chan common.ICmdResp)
	self.runMainThread()

	self.launched = true
	return nil
}

func (self *JobManager) Cancel() {
	self.mainThreadCtx.Cancel()
}

func (self *JobManager) Wait() {
	common.Logger.Printf("Waiting")
	self.mainThreadCtx.Wait()
}

func (self *JobManager) findJob(name string) *jobfile.Job {
	for _, job := range self.jfile.Jobs {
		if job.Name == name {
			return job
		}
	}
	return nil
}

func (self *JobManager) findJobs(names []string) ([]*jobfile.Job, error) {
	foundJobs := make([]*jobfile.Job, 0, len(names))
	missingJobNames := make([]string, 0)
	for _, name := range names {
		job := self.findJob(name)
		if job != nil {
			foundJobs = append(foundJobs, job)
		} else {
			missingJobNames = append(missingJobNames, name)
		}
	}

	if len(missingJobNames) > 0 {
		msg := fmt.Sprintf(
			"No such jobs: %v",
			strings.Join(missingJobNames, ", "),
		)
		return foundJobs, &common.Error{What: msg}
	} else {
		return foundJobs, nil
	}
}

func (self *JobManager) loadJobFile() (*jobfile.JobFile, error) {
	// get current user
	user, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current user.", err))
	}

	// read jobfile
	jfile, err := jobfile.LoadJobFile(self.jobfilePath, user.Username)
	if err == nil {
		return jfile, nil
	} else {
		return nil, &common.Error{
			What: fmt.Sprintf(
				"Failed to read jobfile %v",
				self.jobfilePath,
			),
			Cause: err,
		}
	}
}

func (self *JobManager) handleRunRec(rec *jobfile.RunRec) {
	if rec.Err != nil {
		common.ErrLogger.Panicln(rec.Err)
	}

	// record in run log
	newRunLogEntry := RunLogEntry{
		rec.Job.Name, rec.RunTime, rec.Succeeded, rec.NewStatus,
	}
	self.runLog.Put(newRunLogEntry)

	/* NOTE: error-handler was already applied by the job, if necessary. */

	if (!rec.Succeeded && rec.Job.NotifyOnError) ||
		(rec.Job.NotifyOnFailure && rec.NewStatus == jobfile.JobFailed) {
		// notify user
		self.jfile.Prefs.Notifier(rec)
	}
}

func (self *JobManager) runMainThread() {
	self.mainThreadCtx = common.BackgroundContext().MakeChild()

	go func() {
		/*
		   All modifications to the job manager's state occur here.
		*/
		common.Logger.Println("In job manager main thread")

		// load job file
		jfile, err := self.loadJobFile()
		if err == nil {
			self.jfile = jfile
		} else {
			common.ErrLogger.Printf("%v", err)
			self.jfile = jobfile.NewEmptyJobFile()
		}

		// start job-runner thread
		self.jobRunner.Start(self.jfile.Jobs, self.Shell, self.mainThreadCtx)

	Loop:
		for {
			select {
			case <-self.mainThreadCtx.CancelledChan():
				common.Logger.Println("Main thread cancelled")
				break Loop

			case rec, ok := <-self.jobRunner.RunRecChan():
				if ok {
					self.handleRunRec(rec)
				} else {
					common.ErrLogger.Println("jobfile.Job-runner thread ended prematurely.")
					break Loop
				}

			case cmd, ok := <-self.CmdChan:
				if ok {
					var shouldExit bool
					self.doCmd(cmd, &shouldExit)
					if shouldExit {
						break Loop
					}
				} else {
					common.ErrLogger.Println("Command channel was closed.")
					break Loop
				}
			}
		}

		// cancel main thread
		self.mainThreadCtx.Cancel()

		// consume all run-records
		for rec := range self.jobRunner.RunRecChan() {
			self.handleRunRec(rec)
		}

		// wait for job-runner thread to finish
		self.mainThreadCtx.Finish()
	}()
}

func (self *JobManager) doCmd(
	tmpCmd common.ICmd,
	shouldExit *bool) { // runs in main thread

	*shouldExit = false

	switch cmd := tmpCmd.(type) {
	case common.ReloadCmd:
		// read job file
		newJfile, err := self.loadJobFile()
		if err != nil {
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
		self.jobRunner.Start(self.jfile.Jobs, self.Shell, self.mainThreadCtx)

		// make response
		cmd.RespChan <- &common.ReloadCmdResp{NumJobs: len(self.jfile.Jobs)}
		close(cmd.RespChan)
		return

	case common.ListJobsCmd:
		// make job list
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
				NextRunTime:  j.NextRunTime,
				NotifyOnErr:  j.NotifyOnError,
				NotifyOnFail: j.NotifyOnFailure,
				ErrHandler:   j.ErrorHandler.String(),
			}
			if j.Paused {
				jobDesc.Status += " (Paused)"
				jobDesc.NextRunTime = nil
			}

			jobDescs = append(jobDescs, jobDesc)
		}

		// make response
		cmd.RespChan <- &common.ListJobsCmdResp{Jobs: jobDescs}
		close(cmd.RespChan)
		return

	case common.LogCmd:
		// make log list
		logDescs := make([]common.LogDesc, 0)
		entries, err := self.runLog.GetFromIndex()
		if err != nil {
			cmd.RespChan <- &common.LogCmdResp{Err: err}
			close(cmd.RespChan)
			return
		}
		for _, l := range entries {
			logDesc := common.LogDesc{
				l.Time,
				l.JobName,
				l.Succeeded,
				l.Result.String(),
			}
			logDescs = append(logDescs, logDesc)
		}

		// make response
		cmd.RespChan <- &common.LogCmdResp{Logs: logDescs}
		close(cmd.RespChan)
		return

	case common.TestCmd:
		// find job
		job := self.findJob(cmd.Job)
		if job == nil {
			cmd.RespChan <- &common.TestCmdResp{
				Err: &common.Error{What: "No such job."},
			}
			close(cmd.RespChan)
			return
		}

		// run the job in this thread
		runRec := RunJob(job, self.Shell, true)

		// make response
		if runRec.Err == nil {
			cmd.RespChan <- &common.TestCmdResp{Result: runRec.Describe()}
		} else {
			cmd.RespChan <- &common.TestCmdResp{Err: runRec.Err}
		}
		close(cmd.RespChan)
		return

	case common.CatCmd:
		// find job
		job := self.findJob(cmd.Job)
		if job == nil {
			cmd.RespChan <- &common.CatCmdResp{
				Err: &common.Error{What: "No such job."},
			}
			close(cmd.RespChan)
			return
		}

		// make response
		cmd.RespChan <- &common.CatCmdResp{Result: job.Cmd}
		close(cmd.RespChan)
		return

	case common.PauseCmd:
		// look up jobs to pause
		var jobsToPause []*jobfile.Job
		if len(cmd.Jobs) > 0 {
			var err error
			jobsToPause, err = self.findJobs(cmd.Jobs)
			if err != nil {
				cmd.RespChan <- &common.PauseCmdResp{Err: err}
				close(cmd.RespChan)
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
		close(cmd.RespChan)
		return

	case common.ResumeCmd:
		// look up jobs to resume
		var jobsToResume []*jobfile.Job
		if len(cmd.Jobs) > 0 {
			var err error
			jobsToResume, err = self.findJobs(cmd.Jobs)
			if err != nil {
				cmd.RespChan <- &common.ResumeCmdResp{Err: err}
				close(cmd.RespChan)
				return
			}
		} else {
			jobsToResume = self.jfile.Jobs
		}

		// pause them
		amtResumed := 0
		for _, job := range jobsToResume {
			if job.Paused {
				job.Paused = false
				amtResumed += 1
			}
		}

		// make response
		cmd.RespChan <- &common.ResumeCmdResp{AmtResumed: amtResumed}
		close(cmd.RespChan)
		return

	default:
		common.ErrLogger.Printf("Unknown command: %v", cmd)
		return
	}
}
