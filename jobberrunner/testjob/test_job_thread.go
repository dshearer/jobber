package testjob

import (
	"context"
	"io"
	"os/exec"
	"time"

	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
)

// testJobThread represents a thread that is spawned when the user calls the
// "test" command. The thread runs one job one time. As the job runs, the
// thread writes its output to given writers. When the job finishes,
// the thread reports the result via a channel.
type testJobThread struct {
	Stdout io.Writer
	Stderr io.Writer

	runRecChan chan *jobfile.RunRec
	cancelChan chan interface{}
}

// Run spawns the thread that runs the job.
func (self *testJobThread) Run(ctx context.Context, job *jobfile.Job, shell string) error {
	// make channels
	self.runRecChan = make(chan *jobfile.RunRec)

	// make subproc
	cmd := exec.CommandContext(ctx, shell, "-c", job.Cmd)
	cmd.Stdout = self.Stdout
	cmd.Stderr = self.Stderr

	// launch subproc
	if err := cmd.Start(); err != nil {
		return err
	}
	go self.runThread(ctx, job, cmd)

	return nil
}

func (self *testJobThread) runThread(ctx context.Context, job *jobfile.Job, cmd *exec.Cmd) {
	rec := jobfile.RunRec{Job: job, RunTime: time.Now()}

	// clean up
	defer close(self.runRecChan)

	// launch subproc
	waitErrChan := make(chan error)
	go func() {
		defer close(waitErrChan)
		waitErrChan <- cmd.Wait()
	}()

	// wait for something to happen
	var waitErr error
	select {
	case waitErr = <-waitErrChan:
		/* Subproc finished */
		break

	case <-ctx.Done():
		// wait for subproc to finish
		waitErr = <-waitErrChan
		break
	}

	// report result
	if waitErr == nil {
		rec.Fate = common.SubprocFateSucceeded
	} else if ctx.Err() != nil {
		rec.Fate = common.SubprocFateCancelled
	} else {
		rec.Fate = common.SubprocFateFailed
	}
	rec.NewStatus = jobfile.JobGood
	rec.ExecTime = time.Since(rec.RunTime)
	self.runRecChan <- &rec
}

// ResultChan returns a channel on which the result is written.
// If the thread is not currently running, returns a closed channel.
//
// Never returns nil (so it's always safe to read from the returned channel).
func (self *testJobThread) ResultChan() <-chan *jobfile.RunRec {
	if self.runRecChan == nil {
		tmpChan := make(chan *jobfile.RunRec)
		close(tmpChan)
		return tmpChan
	}
	return self.runRecChan
}
