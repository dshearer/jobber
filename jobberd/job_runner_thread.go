package main

import (
    "github.com/dshearer/jobber/common"
    "github.com/dshearer/jobber/jobfile"
    "github.com/dshearer/jobber/Godeps/_workspace/src/golang.org/x/net/context"
    "time"
)

type JobRunnerThread struct {
    running     bool
    runRecChan  chan *jobfile.RunRec
    ctx         *JobberContext
    ctl         JobberCtl
}

func NewJobRunnerThread() *JobRunnerThread {
    return &JobRunnerThread{
        running: false,
    }
}

func (t *JobRunnerThread) RunRecChan() <-chan *jobfile.RunRec {
    return t.runRecChan
}

func (t *JobRunnerThread) Start(jobs []*jobfile.Job, shell string, ctx *JobberContext) {
    if t.running {
        panic("JobRunnerThread already running.")
    }
    t.running = true
    
    t.runRecChan = make(chan *jobfile.RunRec)
    var jobQ JobQueue
    jobQ.SetJobs(time.Now(), jobs)
    
    // make subcontext
    t.ctx, t.ctl = NewJobberContext(ctx)
    //Logger.Printf("Job Runner thread context: %v\n", t.ctx.Name)
    
    go func() {
        for {
            var job *jobfile.Job = jobQ.Pop(time.Now(), t.ctx) // sleeps
        
            if job != nil && !job.Paused {
                // launch thread to run this job
                common.Logger.Printf("%v: %v\n", job.User, job.Cmd)
                subsubctx, _ := NewJobberContext(t.ctx)
                go func(job *jobfile.Job) {
                    t.runRecChan <- RunJob(job, subsubctx, shell, false)
                    subsubctx.Finish()
                }(job)
            
            } else if job == nil {
                /* We were canceled. */
                //Logger.Printf("Run thread got 'stop'\n")
                break
            }
        }
        
        // wait for run threads to stop
        //Logger.Printf("JobRunner: cleaning up...\n")
        t.ctx.Finish()
        
        // close run-rec channel
        close(t.runRecChan)
        //Logger.Printf("JobRunner done\n")
    }()
}

func (t *JobRunnerThread) Cancel() {
    if t.running {
        t.ctl.Cancel()
        t.running = false
    }
}

func (t *JobRunnerThread) Wait() {
    t.ctl.Wait()
}

func RunJob(job *jobfile.Job, ctx context.Context, shell string, testing bool) *jobfile.RunRec {
	rec := &jobfile.RunRec{Job: job, RunTime: time.Now()}

	// run
	var sudoResult *common.SudoResult
	sudoResult, err := common.Sudo(job.User, job.Cmd, shell, nil)

	if err != nil {
		/* unexpected error while trying to run job */
		rec.Err = err
		return rec
	}

	// update run rec
	rec.Succeeded = sudoResult.Succeeded
	rec.NewStatus = jobfile.JobGood
	rec.Stdout = &sudoResult.Stdout
	rec.Stderr = &sudoResult.Stderr

	if !testing {
		// update job
		if sudoResult.Succeeded {
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