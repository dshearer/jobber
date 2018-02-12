package main

import (
	"context"
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
	"os/exec"
	"sync"
	"time"
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

func (self *JobRunnerThread) Start(jobs []*jobfile.Job, shell string) {

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

	common.Logger.Println("Launching job runner thread")
	go func() {
		defer close(self.mainThreadDoneChan)

		var jobThreadWaitGroup sync.WaitGroup

		for {
			common.Logger.Println("Calling pop")
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
				common.Logger.Printf("Run thread got 'stop'\n")
				break
			}
		}

		// wait for run threads to stop
		common.Logger.Printf("JobRunner: cleaning up...")
		jobThreadWaitGroup.Wait()

		// close run-rec channel
		close(self.runRecChan)
		common.Logger.Println("JobRunner done")
	}()
}

func (self *JobRunnerThread) Cancel() {
	common.Logger.Println("JobRunner: cancelling")
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
