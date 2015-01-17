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

func nextRunTime(job *Job, now time.Time) time.Time {
    var next time.Time = now
    for {
        if job.Sec.Value != nil && next.Second() != int(*job.Sec.Value) {
            /*
             * The earliest possible time is the next time t
             * s.t. t.Second == job.Sec.
             */
            if next.Second() > int(*job.Sec.Value) {
                next = next.Add(time.Minute)
            }
            next = setTimeComp(next, next.Second(), int(*job.Sec.Value), time.Second)
            
        } else if job.Min.Value != nil && next.Minute() != int(*job.Min.Value) {
            /*
             * The earliest possible time is the next time t
             * s.t. t.Minute == job.Min.
             */
            if next.Minute() > int(*job.Min.Value) {
                next = next.Add(time.Hour)
            }
            next = setTimeComp(next, next.Minute(), int(*job.Min.Value), time.Minute)
            next = setTimeComp(next, next.Second(), 0, time.Second)
            
        } else if job.Hour.Value != nil && next.Hour() != int(*job.Hour.Value) {
            if next.Hour() > int(*job.Hour.Value) {
                next = next.AddDate(0, 0, 1) // add 1 day
            }
            next = setTimeComp(next, next.Hour(), int(*job.Hour.Value), time.Hour)
            next = setTimeComp(next, next.Minute(), 0, time.Minute)
            next = setTimeComp(next, next.Second(), 0, time.Second)
            
        } else if job.Wday.Value != nil && weekdayToInt(next.Weekday()) != int(*job.Wday.Value) {
            if weekdayToInt(next.Weekday()) > int(*job.Wday.Value) {
                next = next.AddDate(0, 0, 7) // add 7 days
            }
            deltaDays := int(*job.Wday.Value) - weekdayToInt(next.Weekday())
            next = next.AddDate(0, 0, deltaDays)
            next = setTimeComp(next, next.Hour(), 0, time.Hour)
            next = setTimeComp(next, next.Minute(), 0, time.Minute)
            next = setTimeComp(next, next.Second(), 0, time.Second)
            
        } else if job.Mday.Value != nil && next.Day() != int(*job.Mday.Value) {
            if next.Day() > int(*job.Mday.Value) || !monthHasDay(next.Month(), int(*job.Mday.Value)) {
                next = next.AddDate(0, 1, 0) // add 1 month
                deltaDays := 1 - next.Day()
                next = next.AddDate(0, 0, deltaDays) // set mday to 1
            } else {
                deltaDays := int(*job.Mday.Value) - next.Day()
                next = next.AddDate(0, 0, deltaDays) // set mday to job.Mday
            }
            next = setTimeComp(next, next.Hour(), 0, time.Hour)
            next = setTimeComp(next, next.Minute(), 0, time.Minute)
            next = setTimeComp(next, next.Second(), 0, time.Second)
            
        } else if job.Mon.Value != nil && monthToInt(next.Month()) != int(*job.Mon.Value) {
            if monthToInt(next.Month()) > int(*job.Mon.Value) {
                next = next.AddDate(1, 0, 0) // add 1 year
            }
            deltaMonths := int(*job.Mon.Value) - monthToInt(next.Month())
            next = next.AddDate(0, deltaMonths, 0)
            deltaDays := 1 - next.Day()
            next = next.AddDate(0, 0, deltaDays) // set mday to 1
            next = setTimeComp(next, next.Hour(), 0, time.Hour)
            next = setTimeComp(next, next.Minute(), 0, time.Minute)
            next = setTimeComp(next, next.Second(), 0, time.Second)
            
        } else {
            break
        }
    }
    return next;
}

type scheduledJob struct {
    job      *Job
    runTime  time.Time
}

type priQueue []scheduledJob // implements heap.Interface

func (q priQueue) Len() int {
    return len(q)
}

func (q priQueue) Less(i, j int) bool {
    return q[i].runTime.Before(q[j].runTime)
}

func (q priQueue) Swap(i, j int) {
    q[i], q[j] = q[j], q[i]
}

func (q *priQueue) Push(x interface{}) {
    *q = append(*q, x.(scheduledJob))
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
    jq.q = make(priQueue, len(jobs))
    for i := 0; i < len(jobs); i++ {
        jq.q[i] = scheduledJob{jobs[i], nextRunTime(jobs[i], now)}
    }
    heap.Init(&jq.q)
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
        schedJob := heap.Pop(&jq.q).(scheduledJob)
        job := schedJob.job
        
        // sleep till it's time to run it
        if now.Before(schedJob.runTime) {
            afterChan := time.After(schedJob.runTime.Sub(now))
            select {
                case now = <-afterChan:
                case <-ctx.Done():
                    // abort!
                    heap.Push(&jq.q, schedJob)
                    return nil
            }
        }
        
        // schedule this job's next run
        schedJob2 := scheduledJob{job, nextRunTime(job, now.Add(time.Second))}
        heap.Push(&jq.q, schedJob2)
        
        // decide whether we really should run this job
        if job.ShouldRun() {
            return job
        } else {
            // skip this job
            return jq.Pop(now, ctx)
        }
    }
}




