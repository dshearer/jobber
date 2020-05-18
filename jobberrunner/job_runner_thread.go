package main

import (
	"context"
	"sync"
	"time"

	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
)

// JobRunnerThread implements the thread that schedules and runs all the jobs.
//
// There should be only one instance of JobRunnerThread. It can be reused --
// that is, started and stopped multiple times.
//
// The general usage pattern is this:
// 1. Start the thread with method Start
// 2. Read RunRecs off of the channel returned by method RunRecChan
//
// When you want to stop the thread, do this:
// 1. Stop reading from the RunRec channel
// 2. Call method Cancel
// 3. Read from the RunRec channel until it is closed
//
// At any time, the RunRec channel is closed iff the thread is not running.
type JobRunnerThread struct {
	Running            bool
	runRecChan         chan *jobfile.RunRec
	mainThreadDoneChan chan interface{}
	ctxCancel          context.CancelFunc
}

// RunRecChan returns a channel on which records of completed jobs are written.
// If the job runner thread is not currently running, returns a closed channel.
//
// Never returns nil.
func (self *JobRunnerThread) RunRecChan() <-chan *jobfile.RunRec {
	if self.runRecChan == nil {
		tmpChan := make(chan *jobfile.RunRec)
		close(tmpChan)
		return tmpChan
	}
	return self.runRecChan
}

func (self *JobRunnerThread) Start(jobs map[string]*jobfile.Job, shell string) {
	if self.Running {
		panic("JobRunnerThread already running.")
	}

	// NOTE: order of these is important:
	self.Running = true
	self.runRecChan = make(chan *jobfile.RunRec)

	// make subcontext
	ctx, cancel := context.WithCancel(context.Background())
	self.ctxCancel = cancel

	// make job queue
	var jobQ JobQueue
	jobQ.SetJobs(time.Now(), jobs)

	go func() {
		// NOTE: order of these is important:
		defer close(self.runRecChan)
		defer func() { self.Running = false }()

		var jobThreadWaitGroup sync.WaitGroup
		for {
			var job *jobfile.Job = jobQ.Pop(ctx, time.Now()) // sleeps

			if job != nil && !job.Paused {
				// launch thread to run this job
				common.Logger.Printf("%v: %v\n", job.User, job.Cmd)
				jobThreadWaitGroup.Add(1)
				go func(job *jobfile.Job) {
					defer jobThreadWaitGroup.Done()
					self.runRecChan <- RunJob(ctx, job, shell, false)
				}(job)

			} else if job == nil {
				/* We were canceled. */
				break
			}
		}

		/*
			We've been told to stop. We will no longer start any jobs, but there
			may still be jobs running (and writing to the run rec chan). We need to
			wait for them to stop.
		*/

		// wait for run threads to stop
		jobThreadWaitGroup.Wait()

		// close run rec chan
	}()
}

// Cancel tells the thread to stop scheduling new jobs.
// Note that it returns immediately; the thread may still
// be running.
func (self *JobRunnerThread) Cancel() {
	if self.ctxCancel != nil {
		self.ctxCancel()
	}
}

func RunJob(
	ctx context.Context,
	job *jobfile.Job,
	shell string,
	testing bool) *jobfile.RunRec {

	rec := &jobfile.RunRec{Job: job, RunTime: time.Now()}

	// run
	var execResult *common.ExecResult
	common.Logger.Println("Running job...")
	execResult, err :=
		common.ExecAndWaitContext(ctx, []string{shell, "-c", job.Cmd}, nil)
	common.Logger.Println("Job done")

	if err != nil {
		/* unexpected error while trying to run job */
		common.ErrLogger.Printf("Unexpected error from ExecAndWaitContext: %v\n", err)
		rec.Err = err
		return rec
	}
	defer execResult.Close()

	// get output
	rec.Stdout, err = execResult.ReadStdout(jobfile.RunRecOutputMaxLen)
	if err != nil {
		common.ErrLogger.Printf("Failed to read job's stdout: %v\n", err)
		rec.Err = err
		return rec
	}
	rec.Stderr, err = execResult.ReadStderr(jobfile.RunRecOutputMaxLen)
	if err != nil {
		common.ErrLogger.Printf("Failed to read job's stderr: %v\n", err)
		rec.Err = err
		return rec
	}

	// update run rec
	rec.Fate = execResult.Fate
	rec.NewStatus = jobfile.JobGood
	rec.ExecTime = time.Since(rec.RunTime)

	if testing {
		return rec
	}

	// update job
	switch execResult.Fate {
	case common.SubprocFateSucceeded:
		job.Status = jobfile.JobGood
		break
	case common.SubprocFateFailed:
		/* job failed: apply error-handler (which sets job.Status) */
		job.ErrorHandler.Handle(job)
		break
	}
	job.LastRunTime = rec.RunTime

	// update rec.NewStatus
	rec.NewStatus = job.Status

	return rec
}
