package main

import (
	"context"
	"fmt"
	"os"
	"os/user"

	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/ipc"
	"github.com/dshearer/jobber/jobfile"
)

type CmdContainer struct {
	Cmd      ipc.ICmd
	RespChan chan<- ipc.ICmdResp
}

type JobManager struct {
	jobfilePath         string
	launched            bool
	jfile               *jobfile.JobFile
	CmdChan             chan CmdContainer
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
	tmp := jobfile.NewEmptyJobFile()
	jm.jfile = &tmp
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
	self.mainThreadCtxCancel()
}

func (self *JobManager) Wait() {
	<-self.mainThreadDoneChan
}

/*
Stop the job-runner thread, replace the current jobfile with the given
one, then start the job-runner thread.
*/
func (self *JobManager) replaceCurrJobfile(jfile *jobfile.JobFile) {
	if self.jobRunner.Running {
		// stop job-runner thread and wait for current runs to end
		self.jobRunner.Cancel()
		for rec := range self.jobRunner.RunRecChan() {
			self.handleRunRec(rec)
		}
		self.jobRunner.Wait()
	}

	// set jobfile
	self.jfile = jfile

	// set loggers
	if len(jfile.Prefs.LogPath) > 0 {
		common.SetLogFile(jfile.Prefs.LogPath)
	} else {
		common.LogToStdoutStderr()
	}

	// start job-runner thread
	self.jobRunner.Start(self.jfile.Jobs, self.Shell)
}

func (self *JobManager) openJobfile(path string,
	usr *user.User) (*jobfile.JobFile, error) {

	// open jobfile
	f, err := os.Open(self.jobfilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// check jobfile
	jobfileGood, err := jobfile.ShouldLoadJobfile(f, usr)
	if !jobfileGood {
		if os.IsNotExist(err) {
			return nil, &common.Error{What: "Problem with jobfile", Cause: err}
		} else {
			return nil, err
		}
	}

	// read jobfile
	jobfile, err := jobfile.LoadJobfile(f, usr)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &common.Error{What: "Problem with jobfile", Cause: err}
		} else {
			return nil, err
		}
	}
	return jobfile, nil
}

/*
Replaces in-memory jobfile with the current version on disk.  If there
is no jobfile on disk, sets in-memory jobfile to an empty jobfile.  In
both cases, restarts the job-runner thread and sets the loggers.

If an error happens when trying to read the on-disk jobfile, does not
change the in-memory jobfile, and returns that error.
*/
func (self *JobManager) loadJobfile() error {

	/*
			   If there is no jobfile:
			       1. Stop job-runner thread
			       2. Replace current jobfile with empty one
			       3. Start job-runner thread

			   If there is a jobfile with no errors:
			       1. Stop job-runner thread
			       2. Replace current jobfile with new one
			       3. Start job-runner thread

			   If there is a jobfile but it has an error:
		        	   1. If job-runner thread is not running, start it
		        	   2. Return the error
	*/

	// get current user
	usr, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current user: %v", err))
	}

	// open jobfile
	jfile, err := self.openJobfile(self.jobfilePath, usr)

	if err == nil || os.IsNotExist(err) {
		if jfile == nil {
			tmp := jobfile.NewEmptyJobFile()
			jfile = &tmp
		}
		self.replaceCurrJobfile(jfile)
		return nil

	} else {
		if !self.jobRunner.Running {
			// start job-runner thread
			self.jobRunner.Start(self.jfile.Jobs, self.Shell)
		}

		// report error
		return err
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

	if rec.Succeeded {
		rec.Job.NotifyOnSuccess.Notify(rec)
	} else {
		rec.Job.NotifyOnError.Notify(rec)
	}
	if rec.NewStatus == jobfile.JobFailed {
		rec.Job.NotifyOnFailure.Notify(rec)
	}
}

func (self *JobManager) runMainThread() {
	ctx, cancel :=
		context.WithCancel(context.Background())
	self.mainThreadCtxCancel = cancel

	self.CmdChan = make(chan CmdContainer)
	self.mainThreadDoneChan = make(chan interface{})

	go func() {
		/*
		   All modifications to the job manager's state occur here.
		*/
		defer close(self.mainThreadDoneChan)
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
				break Loop

			case rec, ok := <-self.jobRunner.RunRecChan():
				if !ok {
					common.ErrLogger.Panic("Job-runner thread " +
						"ended prematurely.")
				}
				self.handleRunRec(rec)

			case cmd, ok := <-self.CmdChan:
				if ok {
					var shouldExit bool
					cmd.RespChan <- self.doCmd(cmd.Cmd, &shouldExit)
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
		} // for

		// cancel job runner
		self.jobRunner.Cancel()

		// consume all run-records
		for rec := range self.jobRunner.RunRecChan() {
			self.handleRunRec(rec)
		}

		// wait for job runner to fully stop
		self.jobRunner.Wait()
	}()
}

func (self *JobManager) doCmd(
	tmpCmd ipc.ICmd,
	shouldExit *bool) ipc.ICmdResp { // runs in main thread

	*shouldExit = false

	switch cmd := tmpCmd.(type) {
	case ipc.ReloadCmd:
		return self.doReloadCmd(cmd)

	case ipc.ListJobsCmd:
		return self.doListJobsCmd(cmd)

	case ipc.LogCmd:
		return self.doLogCmd(cmd)

	case ipc.TestCmd:
		return self.doTestCmd(cmd)

	case ipc.CatCmd:
		return self.doCatCmd(cmd)

	case ipc.PauseCmd:
		return self.doPauseCmd(cmd)

	case ipc.ResumeCmd:
		return self.doResumeCmd(cmd)

	case ipc.InitCmd:
		return self.doInitCmd(cmd)

	case ipc.SetJobCmd:
		return self.doSetJobCmd(cmd)

	case ipc.DeleteJobCmd:
		return self.doDeleteJobCmd(cmd)

	default:
		return ipc.NewErrorCmdResp(
			&common.Error{What: fmt.Sprintf("Unknown command: %v", cmd)},
		)
	}
}
