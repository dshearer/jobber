package main

import (
	"fmt"
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
	"os/user"
	"time"
)

type RunLogEntry struct {
	Job       *jobfile.Job
	Time      time.Time
	Succeeded bool
	Result    jobfile.JobStatus
}

/* For sorting RunLogEntries: */
type runLogEntrySorter struct {
	entries []RunLogEntry
}

/* For sorting RunLogEntries: */
func (s *runLogEntrySorter) Len() int {
	return len(s.entries)
}

/* For sorting RunLogEntries: */
func (s *runLogEntrySorter) Swap(i, j int) {
	s.entries[i], s.entries[j] = s.entries[j], s.entries[i]
}

/* For sorting RunLogEntries: */
func (s *runLogEntrySorter) Less(i, j int) bool {
	return s.entries[i].Time.After(s.entries[j].Time)
}

type JobManager struct {
	jobfilePath   string
	launched      bool
	jfile         *jobfile.JobFile
	runLog        []RunLogEntry
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
	return &jm
}

func (self *JobManager) Launch() error {
	if self.launched {
		return &common.Error{"Already launched.", nil}
	}

	if err := self.loadJobs(); err != nil {
		return &common.Error{
			fmt.Sprintf("Failed to read jobfile %v", self.jobfilePath),
			err,
		}
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

func (self *JobManager) loadJobs() error {
	// read jobfile
	user, err := user.Current()
	if err != nil {
		return &common.Error{"Failed to get current user.", err}
	}
	jfile, err := jobfile.LoadJobFile(self.jobfilePath, user.Username)
	self.jfile = jfile
	return err
}

func (self *JobManager) reloadJobs() error {
	// stop job-runner thread and wait for current runs to end
	self.jobRunner.Cancel()
	for rec := range self.jobRunner.RunRecChan() {
		self.handleRunRec(rec)
	}
	self.jobRunner.Wait()

	// reload jobs
	if err := self.loadJobs(); err != nil {
		return err
	}

	// restart job-runner thread
	self.jobRunner.Start(self.jfile.Jobs, self.Shell, self.mainThreadCtx)

	return nil
}

func (self *JobManager) handleRunRec(rec *jobfile.RunRec) {
	if rec.Err != nil {
		common.ErrLogger.Panicln(rec.Err)
	}

	self.runLog = append(self.runLog,
		RunLogEntry{rec.Job, rec.RunTime, rec.Succeeded, rec.NewStatus})

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
					shouldStop := self.doCmd(cmd)
					if shouldStop {
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

func (self *JobManager) doCmd(cmd common.ICmd) bool { // runs in main thread
	switch cmd.(type) {
	case common.ReloadCmd:
		// load jobs
		err := self.reloadJobs()

		// send response
		var resp common.ReloadCmdResp
		if err == nil {
			resp.NumJobs = len(self.jfile.Jobs)
		} else {
			resp.Err = err
		}
		self.CmdRespChan <- &resp

		return false

	default:
		resp := common.ReloadCmdResp{Err: &common.Error{What: "Unknown command."}}
		self.CmdRespChan <- &resp
		return false
	}
}
