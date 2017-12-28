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
	jobfilePath         string
	launched            bool
	jfile               *jobfile.JobFile
	CmdChan             chan common.ICmd
	CmdRespChan         chan common.ICmdResp
	mainThreadCtx       context.Context
	mainThreadCtxCancel context.CancelFunc
	mainThreadDoneChan  chan interface{}
	jobRunner           *JobRunnerThread
	Shell               string
}

func NewJobManager(jobfilePath string) *JobManager {
	jm := JobManager{Shell: "/bin/sh"}
	jm.jobfilePath = jobfilePath
	jm.jobRunner = NewJobRunnerThread()
	jm.jfile = jobfile.NewEmptyJobFile()
	common.LogToStdoutStderr()
	return &jm
}

func (self *JobManager) Launch() error {
	if self.launched {
		return &common.Error{What: "Already launched."}
	}
	self.runMainThread()

	self.launched = true
	return nil
}

func (self *JobManager) Cancel() {
	common.Logger.Println("JobManager cancelling...")
	self.mainThreadCtxCancel()
	common.Logger.Println("done")
}

func (self *JobManager) Wait() {
	common.Logger.Printf("JobManager Waiting...")
	<-self.mainThreadDoneChan
	common.Logger.Println("done")
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

/*
Replaces in-memory jobfile with the current version on disk.  If there
is no jobfile on disk, sets in-memory jobfile to an empty jobfile.  In
both cases, restarts the job-runner thread and sets the loggers.

If an error happens when trying to read the on-disk jobfile, does not
change the in-memory jobfile, and returns that error.
*/
func (self *JobManager) loadJobfile() error {

	// get current user
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current user: %v", err))
	}

	var jfile *jobfile.JobFile

	// open jobfile
	f, err := os.Open(self.jobfilePath)
	if err != nil && os.IsNotExist(err) {
		/* This is not an error. */
		jfile = jobfile.NewEmptyJobFile()
	} else {
		defer f.Close()

		// check jobfile
		flag, err := jobfile.ShouldLoadJobfile(f, usr)
		if !flag {
			/* This is an error. */
			msg := fmt.Sprintf("Failed to read jobfile %v",
				self.jobfilePath)
			return &common.Error{What: msg, Cause: err}
		}

		// read jobfile
		jfile, err = jobfile.LoadJobfile(f, usr)
		if err != nil {
			/* This is an error */
			msg := fmt.Sprintf("Failed to read jobfile %v",
				self.jobfilePath)
			return &common.Error{What: msg, Cause: err}
		}
	}

	// stop job-runner thread and wait for current runs to end
	common.Logger.Println("Stopping job-runner thread...")
	self.jobRunner.Cancel()
	for rec := range self.jobRunner.RunRecChan() {
		self.handleRunRec(rec)
	}
	self.jobRunner.Wait()
	common.Logger.Println("done")

	// set jobfile
	self.jfile = jfile

	// set loggers
	if len(jfile.Prefs.LogPath) > 0 {
		common.SetLogFile(jfile.Prefs.LogPath)
	} else {
		common.LogToStdoutStderr()
	}

	// restart job-runner thread
	common.Logger.Println("Starting job-runner thread...")
	self.jobRunner.Start(self.jfile.Jobs, self.Shell)
	common.Logger.Println("done")

	return nil
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
	ctx, cancel :=
		context.WithCancel(context.Background())
	self.mainThreadCtxCancel = cancel

	self.CmdChan = make(chan common.ICmd)
	self.CmdRespChan = make(chan common.ICmdResp)
	self.mainThreadDoneChan = make(chan interface{})

	go func() {
		/*
		   All modifications to the job manager's state occur here.
		*/
		common.Logger.Println("In job manager main thread")
		defer close(self.mainThreadDoneChan)
		defer close(self.CmdRespChan)
		defer close(self.CmdChan)

		// load job file & start job-runner thread
		err := self.loadJobfile()
		if err != nil {
			common.ErrLogger.Printf("%v", err)
		}

	Loop:
		for {
			select {
			case <-ctx.Done():
				common.Logger.Println("Main thread cancelled")
				break Loop

			case rec, ok := <-self.jobRunner.RunRecChan():
				if !ok {
					common.ErrLogger.Panic("Job-runner thread " +
						"ended prematurely.")
				}
				self.handleRunRec(rec)

			case cmd, ok := <-self.CmdChan:
				if ok {
					common.Logger.Println("Got command")
					var shouldExit bool
					self.doCmd(cmd, &shouldExit)
					if shouldExit {
						self.mainThreadCtxCancel()
						break Loop
					}
				} else {
					common.ErrLogger.Println("Command channel was " +
						"closed.")
					self.mainThreadCtxCancel()
					break Loop
				}
			}
		}

		// cancel job runner
		self.jobRunner.Cancel()

		// consume all run-records
		common.Logger.Println("Consuming remaining run recs...")
		for rec := range self.jobRunner.RunRecChan() {
			self.handleRunRec(rec)
		}
		common.Logger.Println("Done onsuming remaining run recs")

		// wait for job runner to fully stop
		self.jobRunner.Wait()
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
