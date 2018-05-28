package main

import (
	"container/heap"
	"context"
	"time"

	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
)

func nextRunTime(job *jobfile.Job, now time.Time) *time.Time {
	/*
	 * We test every second from now till 2 years from now,
	 * looking for a time that satisfies the job's schedule
	 * criteria.
	 */

	var year time.Duration = time.Hour * 24 * 365
	var max time.Time = now.Add(2 * year)
	for next := now; next.Before(max); next = next.Add(time.Second) {
		if job.FullTimeSpec.Satisfied(next) {
			return &next
		}
	}

	return nil
}

/*
 * jobQueueImpl is a priority queue containing Jobs that sorts
 * them by next run time.
 */
type jobQueueImpl []*jobfile.Job // implements heap.Interface

func (q jobQueueImpl) Len() int {
	return len(q)
}

func (q jobQueueImpl) Less(i, j int) bool {
	return q[i].NextRunTime.Before(*q[j].NextRunTime)
}

func (q jobQueueImpl) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *jobQueueImpl) Push(x interface{}) {
	*q = append(*q, x.(*jobfile.Job))
}

func (q *jobQueueImpl) Pop() interface{} {
	n := len(*q)
	if n == 0 {
		return nil
	} else {
		item := (*q)[n-1]
		*q = (*q)[0 : n-1]
		return item
	}
}

/*
 * A priority queue containing jobs.  It's a public
 * wrapper for an instance of jobQueueImpl.
 */
type JobQueue struct {
	q jobQueueImpl
}

func (jq *JobQueue) SetJobs(now time.Time, jobs []*jobfile.Job) {
	jq.q = make(jobQueueImpl, 0)
	heap.Init(&jq.q)

	for i := 0; i < len(jobs); i++ {
		var job *jobfile.Job = jobs[i]
		job.NextRunTime = nextRunTime(job, now)
		if job.NextRunTime != nil {
			heap.Push(&jq.q, job)
		}
	}
}

func (jq *JobQueue) Empty() bool {
	return jq.q.Len() == 0
}

/*!
 * Get the next job to run, after sleeping until the time it's supposed
 * to run.
 *
 * @return The next job to run, or nil if the context has been canceled.
 */
func (jq *JobQueue) Pop(ctx context.Context, now time.Time) *jobfile.Job {
	if jq.Empty() {
		// just wait till the context has been canceled
		common.Logger.Println("Queue: waiting...")
		<-ctx.Done()
		common.Logger.Println("Queue: done")
		return nil

	} else {
		// get next-scheduled job
		job := heap.Pop(&jq.q).(*jobfile.Job)

		// sleep till it's time to run it
		for now.Before(*job.NextRunTime) {
			afterChan := time.After(job.NextRunTime.Sub(now))
			select {
			case now = <-afterChan:
			case <-ctx.Done():
				// abort!
				heap.Push(&jq.q, job)
				return nil
			}
		}

		// schedule this job's next run
		job.NextRunTime = nextRunTime(job, now.Add(time.Second))
		if job.NextRunTime != nil {
			heap.Push(&jq.q, job)
		}

		// decide whether we really should run this job
		if job.ShouldRun() {
			return job
		} else {
			// skip this job
			return jq.Pop(ctx, now)
		}
	}
}
