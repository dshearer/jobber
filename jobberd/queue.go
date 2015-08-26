package main

import (
    "time"
    "container/heap"
    "code.google.com/p/go.net/context"
)

func monthToInt(m time.Month) int {
    switch m {
        case time.January : return 1
        case time.February : return 2
        case time.March : return 3
        case time.April : return 4
        case time.May : return 5
        case time.June : return 6
        case time.July : return 7
        case time.August : return 8
        case time.September : return 9
        case time.October : return 10
        case time.November : return 11
        default : return 12
    }
}

func weekdayToInt(d time.Weekday) int {
    switch d {
        case time.Sunday: return 0
        case time.Monday: return 1
        case time.Tuesday: return 2
        case time.Wednesday: return 3
        case time.Thursday: return 4
        case time.Friday: return 5
        default: return 6
    }
}

func setTimeComp(t time.Time, oldValue int, newValue int, unit time.Duration) time.Time {
    delta := newValue - oldValue
    return t.Add(time.Duration(delta) * unit)
}

func monthHasDay(month time.Month, day int) bool {
    t := time.Date(2014, month, day, 0, 0, 0, 0, time.Now().Location())
    return t.Month() == month
}

func nextRunTime(job *Job, now time.Time) *time.Time {
    var max time.Time = now.Add(time.Hour * 24 * 366)
    for next := now; next.Before(max); next = next.Add(time.Second) {
        var a bool = true
        a = a && job.Sec.Satisfied(next.Second())
        a = a && job.Min.Satisfied(next.Minute())
        a = a && job.Hour.Satisfied(next.Hour())
        a = a && (job.Wday.Satisfied(weekdayToInt(next.Weekday())) || job.Mday.Satisfied(next.Day()))
        a = a && job.Mon.Satisfied(monthToInt(next.Month()))
        if a {
            Logger.Printf("Scheduled %v: %v\n", job.Name, next)
            return &next
        }
    }
    Logger.Printf("Failed to schedule %v\n", job.Name)
    return nil
}

type priQueue []*Job // implements heap.Interface

func (q priQueue) Len() int {
    return len(q)
}

func (q priQueue) Less(i, j int) bool {
    return q[i].NextRunTime.Before(*q[j].NextRunTime)
}

func (q priQueue) Swap(i, j int) {
    q[i], q[j] = q[j], q[i]
}

func (q *priQueue) Push(x interface{}) {
    *q = append(*q, x.(*Job))
}

func (q *priQueue) Pop() interface{} {
    n := len(*q)
    if n == 0 {
        return nil
    } else {
        item := (*q)[n - 1]
        *q = (*q)[0 : n - 1]
        return item
    }
}

type JobQueue struct {
    q   priQueue
}

func (jq *JobQueue) SetJobs(now time.Time, jobs []*Job) {
    jq.q = make(priQueue, 0)
    heap.Init(&jq.q)
    
    for i := 0; i < len(jobs); i++ {
        var job *Job = jobs[i]
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
func (jq *JobQueue) Pop(now time.Time, ctx context.Context) *Job {
    if jq.Empty() {
        // just wait till the context has been canceled
        <-ctx.Done()
        return nil
        
    } else {
        // get next-scheduled job
        job := heap.Pop(&jq.q).(*Job)
        
        // sleep till it's time to run it
        if now.Before(*job.NextRunTime) {
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
            return jq.Pop(now, ctx)
        }
    }
}




