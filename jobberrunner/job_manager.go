package main

import (
	"context"
	"fmt"
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
	"os"
	"os/user"
	"strings"
)

type JobManager struct {
	jobfilePath      string
	launched         bool
	jfile            *jobfile.JobFile
	CmdChan          chan common.ICmd
	CmdRespChan      chan common.ICmdResp
	mainThreadCtx    common.BetterContext
	mainThreadCtxCtl common.BetterContextCtl
	jobRunner        *JobRunnerThread
	Shell            string
}

func NewJobManager(jobfilePath string) *JobManager {
	jm := JobManager{Shell: "/bin/sh"}
	jm.jobfilePath = jobfilePath
	jm.jobRunner = NewJobRunnerThread()
	return &jm
}

func (self *JobManager) Launch() error {
	if self.launched {
		return &common.Error{What: "Already launched."}
	}

	self.mainThreadCtx, self.mainThreadCtxCtl =
		common.MakeChildContext(context.Background())

	// run main thread
	self.CmdChan = make(chan common.ICmd)
	self.CmdRespChan = make(chan common.ICmdResp)
	self.runMainThread()

	self.launched = true
	return nil
}

func (self *JobManager) Cancel() {
	self.mainThreadCtxCtl.Cancel()
}

func (self *JobManager) Wait() {
	common.Logger.Printf("JobManager Waiting")
	self.mainThreadCtxCtl.WaitForFinish()
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
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current user: %v", err))
	}

	// check jobfile
	flag, err := jobfile.ShouldLoadJobfile(self.jobfilePath, usr)
	if !flag {
		return jobfile.NewEmptyJobFile(), &common.Error{
			What: fmt.Sprintf(
				"Failed to read jobfile %v",
				self.jobfilePath,
			),
			Cause: err,
		}
	}

	// read jobfile
	jfile, err := jobfile.LoadJobFile(self.jobfilePath, usr)
	if err == nil {
		// set loggers
		if len(jfile.Prefs.LogPath) > 0 {
			common.SetLogFile(jfile.Prefs.LogPath)
		}

		return jfile, nil
	} else {
		if os.IsNotExist(err) {
			return jobfile.NewEmptyJobFile(), err
		} else {
			return jobfile.NewEmptyJobFile(), &common.Error{
				What: fmt.Sprintf(
					"Failed to read jobfile %v",
					self.jobfilePath,
				),
				Cause: err,
			}
		}
	}
}

func (self *JobManager) handleRunRec(rec *jobfile.RunRec) {
	if rec.Err != nil {
		common.ErrLogger.Panicln(rec.Err)
	}

	// record in run log
	newRunLogEntry := jobfile.RunLogEntry{
		JobName:   rec.Job.Name,
		Time:      rec.RunTime,
		Succeeded: rec.Succeeded,
		Result:    rec.NewStatus,
	}
	self.jfile.Prefs.RunLog.Put(newRunLogEntry)

	/* NOTE: error-handler was already applied by the job, if necessary. */

	shouldNotify := (!rec.Succeeded && rec.Job.NotifyOnError) ||
		(rec.NewStatus == jobfile.JobFailed && rec.Job.NotifyOnFailure) ||
		(rec.Succeeded && rec.Job.NotifyOnSuccess)
	if shouldNotify {
		// notify user
		self.jfile.Prefs.Notifier(rec)
	}
}

func (self *JobManager) runMainThread() {
	go func() {
		/*
		   All modifications to the job manager's state occur here.
		*/
		common.Logger.Println("In job manager main thread")
		defer self.mainThreadCtx.Finish()

		// load job file
		jfile, err := self.loadJobFile()
		self.jfile = jfile
		if err != nil && !os.IsNotExist(err) {
			common.ErrLogger.Printf("%v", err)
		}

		// start job-runner thread
		self.jobRunner.Start(
			self.jfile.Jobs,
			self.Shell,
			self.mainThreadCtx,
		)

	Loop:
		for {
			select {
			case <-self.mainThreadCtx.Done():
				common.Logger.Println("Main thread cancelled")
				break Loop

			case rec, ok := <-self.jobRunner.RunRecChan():
				if ok {
					self.handleRunRec(rec)
				} else {
					common.ErrLogger.Println("Job-runner thread ended prematurely.")
					self.mainThreadCtxCtl.Cancel()
					break Loop
				}

			case cmd, ok := <-self.CmdChan:
				if ok {
					var shouldExit bool
					self.doCmd(cmd, &shouldExit)
					if shouldExit {
						self.mainThreadCtxCtl.Cancel()
						break Loop
					}
				} else {
					common.ErrLogger.Println("Command channel was closed.")
					self.mainThreadCtxCtl.Cancel()
					break Loop
				}
			}
		}

		// consume all run-records
		for rec := range self.jobRunner.RunRecChan() {
			self.handleRunRec(rec)
		}
	}()
}

func (self *JobManager) doCmd(
	tmpCmd common.ICmd,
	shouldExit *bool) { // runs in main thread

	*shouldExit = false

	switch cmd := tmpCmd.(type) {
	case common.ReloadCmd:
		self.doReloadCmd(cmd)

	case common.ListJobsCmd:
		self.doListJobsCmd(cmd)

	case common.LogCmd:
		self.doLogCmd(cmd)

	case common.TestCmd:
		self.doTestCmd(cmd)

	case common.CatCmd:
		self.doCatCmd(cmd)

	case common.PauseCmd:
		self.doPauseCmd(cmd)

	case common.ResumeCmd:
		self.doResumeCmd(cmd)

	case common.InitCmd:
		self.doInitCmd(cmd)

	default:
		common.ErrLogger.Printf("Unknown command: %v", cmd)
	}
}
