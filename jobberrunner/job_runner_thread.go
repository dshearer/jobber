package main

import (
	"context"
	"os/exec"
	"sync"
	"time"

	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
)

type JobRunnerThread struct {
	Running            bool
	runRecChan         chan *jobfile.RunRec
	mainThreadDoneChan chan interface{}
	ctxCancel          context.CancelFunc
}

func NewJobRunnerThread() *JobRunnerThread {
	jr := JobRunnerThread{
		Running:    false,
		runRecChan: make(chan *jobfile.RunRec),
	}
	close(jr.runRecChan)
	return &jr
}

func (self *JobRunnerThread) RunRecChan() <-chan *jobfile.RunRec {
	return self.runRecChan
}

func (self *JobRunnerThread) Start(jobs map[string]*jobfile.Job, shell string) {

	if self.Running {
		panic("JobRunnerThread already running.")
	}
	self.Running = true

	self.mainThreadDoneChan = make(chan interface{})

	// make subcontext
	ctx, cancel := context.WithCancel(context.Background())
	self.ctxCancel = cancel

	self.runRecChan = make(chan *jobfile.RunRec)

	var jobQ JobQueue
	jobQ.SetJobs(time.Now(), jobs)

	go func() {
		defer close(self.mainThreadDoneChan)

		var jobThreadWaitGroup sync.WaitGroup

		for {
			var job *jobfile.Job = jobQ.Pop(ctx, time.Now()) // sleeps

			if job != nil && !job.Paused {
				// launch thread to run this job
				common.Logger.Printf("%v: %v\n", job.User, job.Cmd)
				jobThreadWaitGroup.Add(1)
				go func(job *jobfile.Job) {
					defer jobThreadWaitGroup.Done()
					self.runRecChan <- RunJob(job, shell, false)
				}(job)

			} else if job == nil {
				/* We were canceled. */
				break
			}
		}

		// wait for run threads to stop
		jobThreadWaitGroup.Wait()

		// close run-rec channel
		close(self.runRecChan)
	}()
}

func (self *JobRunnerThread) Cancel() {
	if self.ctxCancel != nil {
		self.ctxCancel()
		self.Running = false
	}
}

func (self *JobRunnerThread) Wait() {
	if self.mainThreadDoneChan != nil {
		<-self.mainThreadDoneChan
	}
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
		rec.Err = err
		return rec
	}

	// update run rec
	rec.Succeeded = execResult.Succeeded
	rec.NewStatus = jobfile.JobGood
	rec.Stdout = &execResult.Stdout
	rec.Stderr = &execResult.Stderr

	if testing {
		return rec
	}

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

	// write output to disk
	job.StdoutHandler.WriteOutput(execResult.Stdout, job.Name, rec.RunTime)
	job.StderrHandler.WriteOutput(execResult.Stderr, job.Name, rec.RunTime)

	return rec
}
