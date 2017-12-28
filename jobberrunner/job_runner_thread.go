package main

import (
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
	"os/exec"
	"time"
)

type JobRunnerThread struct {
	running    bool
	runRecChan chan *jobfile.RunRec
	ctx        common.BetterContext
	ctxCtl     common.BetterContextCtl
}

func NewJobRunnerThread() *JobRunnerThread {
	return &JobRunnerThread{
		running: false,
	}
}

func (self *JobRunnerThread) RunRecChan() <-chan *jobfile.RunRec {
	return self.runRecChan
}

func (self *JobRunnerThread) Start(
	jobs []*jobfile.Job,
	shell string,
	ctx common.BetterContext) {

	if self.running {
		panic("JobRunnerThread already running.")
	}
	self.running = true

	self.runRecChan = make(chan *jobfile.RunRec)
	var jobQ JobQueue
	jobQ.SetJobs(time.Now(), jobs)

	// make subcontext
	self.ctx, self.ctxCtl = common.MakeChildContext(ctx)

	go func() {
		defer self.ctx.Finish()

		for {
			var job *jobfile.Job = jobQ.Pop(self.ctx, time.Now()) // sleeps

			if job != nil && !job.Paused {
				// launch thread to run this job
				common.Logger.Printf("%v: %v\n", job.User, job.Cmd)
				subctx, _ := common.MakeChildContext(self.ctx)
				go func(job *jobfile.Job) {
					defer subctx.Finish()
					self.runRecChan <- RunJob(job, shell, false)
				}(job)

			} else if job == nil {
				/* We were canceled. */
				//Logger.Printf("Run thread got 'stop'\n")
				break
			}
		}

		// wait for run threads to stop
		//Logger.Printf("JobRunner: cleaning up...\n")
		self.ctx.WaitForChildren()

		// close run-rec channel
		close(self.runRecChan)
		//Logger.Printf("JobRunner done\n")
	}()
}

func (self *JobRunnerThread) Cancel() {
	self.ctxCtl.Cancel()
	self.running = false
}

func (self *JobRunnerThread) Wait() {
	self.ctxCtl.WaitForFinish()
}

func RunJob(
	job *jobfile.Job,
	shell string,
	testing bool) *jobfile.RunRec {

	rec := &jobfile.RunRec{Job: job, RunTime: time.Now()}

	// run
	var execResult *common.ExecResult
	execResult, err :=
		common.ExecAndWait(exec.Command(shell, "-c", job.Cmd), nil)

	if err != nil {
		/* unexpected error while trying to run job */
		common.Logger.Printf("RunJob: %v", err)
		rec.Err = err
		return rec
	}

	// update run rec
	rec.Succeeded = execResult.Succeeded
	rec.NewStatus = jobfile.JobGood
	rec.Stdout = &execResult.Stdout
	rec.Stderr = &execResult.Stderr

	if !testing {
		// update job
		if execResult.Succeeded {
			/* job succeeded */
			job.Status = jobfile.JobGood
		} else {
			/* job failed: apply error-handler (which sets job.Status) */
			job.ErrorHandler.Apply(job)
		}
		job.LastRunTime = rec.RunTime

		// update rec.NewStatus
		rec.NewStatus = job.Status
	}

	return rec
}
