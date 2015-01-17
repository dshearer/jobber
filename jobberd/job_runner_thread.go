package main

import (
    "time"
)

type JobRunnerThread struct {
    running     bool
    runRecChan  chan *RunRec
    ctx         *JobberContext
    ctl         JobberCtl
}

func NewJobRunnerThread() *JobRunnerThread {
    return &JobRunnerThread{
        running: false,
    }
}

func (t *JobRunnerThread) RunRecChan() <-chan *RunRec {
    return t.runRecChan
}

func (t *JobRunnerThread) Start(jobs []*Job, shell string, ctx *JobberContext) {
    if t.running {
        panic("JobRunnerThread already running.")
    }
    t.running = true
    
    t.runRecChan = make(chan *RunRec)
    var jobQ JobQueue
    jobQ.SetJobs(time.Now(), jobs)
    
    // make subcontext
    t.ctx, t.ctl = NewJobberContext(ctx)
    //Logger.Printf("Job Runner thread context: %v\n", t.ctx.Name)
    
    go func() {
        for {
            var job *Job = jobQ.Pop(time.Now(), t.ctx) // sleeps
        
            if job != nil {
                // launch thread to run this job
                Logger.Printf("%v: %v\n", job.User, job.Cmd)
                subsubctx, _ := NewJobberContext(t.ctx)
                go func(job *Job) {
                    t.runRecChan <- job.Run(subsubctx, shell, false)
                    subsubctx.Finish()
                }(job)
            
            } else {
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
        Logger.Printf("JobRunnerThread: Canceling\n")
        t.ctl.Cancel()
        t.running = false
    }
}

func (t *JobRunnerThread) Wait() {
    t.ctl.Wait()
}