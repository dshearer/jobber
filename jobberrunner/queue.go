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

func (jq *JobQueue) SetJobs(now time.Time, jobs map[string]*jobfile.Job) {
	jq.q = make(jobQueueImpl, 0)
	heap.Init(&jq.q)

	for _, job := range jobs {
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
		<-ctx.Done()
		return nil

	} else {
		// get next-scheduled job
		job := heap.Pop(&jq.q).(*jobfile.Job)

		var timeFmt = "Jan _2 15:04:05"

		common.Logger.Printf("Next job to run is %v, at %v.", job.Name,
			job.NextRunTime.Format(timeFmt))

		// sleep till it's time to run it
		for now.Before(*job.NextRunTime) {
			sleepDur := job.NextRunTime.Sub(now)

			common.Logger.Printf("It is now %v.", now.Format(timeFmt))
			common.Logger.Printf("Sleeping for %v.", sleepDur)

			afterChan := time.After(sleepDur)
			select {
			case now = <-afterChan:
			case <-ctx.Done():
				// abort!
				heap.Push(&jq.q, job)
				return nil
			}
		}

		common.Logger.Printf("It is now %v, which is NOT before %v",
			now.Format(timeFmt), job.NextRunTime.Format(timeFmt))

		// schedule this job's next run
		job.NextRunTime = nextRunTime(job, now.Add(time.Second))
		if job.NextRunTime != nil {
			heap.Push(&jq.q, job)
		}

		// decide whether we really should run this job
		if job.ShouldRun() {
			common.Logger.Printf("Running %v", job.Name)
			return job
		} else {
			// skip this job
			common.Logger.Printf("Skipping %v", job.Name)
			return jq.Pop(ctx, now)
		}
	}
}
