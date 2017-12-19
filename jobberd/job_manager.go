package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
)

type RunLogEntry struct {
	Job       *jobfile.Job
	Time      time.Time
	Succeeded bool
	Result    jobfile.JobStatus
}

/* For sorting RunLogEntries: */
type runLogEntrySorter struct {
	entries []RunLogEntry
}

/* For sorting RunLogEntries: */
func (s *runLogEntrySorter) Len() int {
	return len(s.entries)
}

/* For sorting RunLogEntries: */
func (s *runLogEntrySorter) Swap(i, j int) {
	s.entries[i], s.entries[j] = s.entries[j], s.entries[i]
}

/* For sorting RunLogEntries: */
func (s *runLogEntrySorter) Less(i, j int) bool {
	return s.entries[i].Time.After(s.entries[j].Time)
}

type JobManager struct {
	userPrefs     map[string]jobfile.UserPrefs
	jobs          []*jobfile.Job
	loadedJobs    bool
	runLog        []RunLogEntry
	cmdChan       chan ICmd
	mainThreadCtx *JobberContext
	mainThreadCtl JobberCtl
	jobRunner     *JobRunnerThread
	Shell         string
}

func NewJobManager() (*JobManager, error) {
	jm := JobManager{Shell: "/bin/sh"}
	jm.userPrefs = make(map[string]jobfile.UserPrefs)
	jm.loadedJobs = false
	jm.jobRunner = NewJobRunnerThread()
	return &jm, nil
}

func (m *JobManager) Launch() (chan<- ICmd, error) {
	if m.mainThreadCtx != nil {
		return nil, &common.Error{"Already launched.", nil}
	}

	common.Logger.Println("Launching.")
	if !m.loadedJobs {
		_, err := m.loadAllJobs()
		if err != nil {
			common.ErrLogger.Printf("Failed to load jobs: %v.\n", err)
			return nil, err
		}
	}

	// make main thread
	m.cmdChan = make(chan ICmd)
	m.runMainThread()
	return m.cmdChan, nil
}

func (m *JobManager) Cancel() {
	if m.mainThreadCtl.Cancel != nil {
		m.mainThreadCtl.Cancel()
	}
}

func (m *JobManager) Wait() {
	if m.mainThreadCtl.Wait != nil {
		m.mainThreadCtl.Wait()
	}
}

func (m *JobManager) jobsForUser(username string) []*jobfile.Job {
	jobs := make([]*jobfile.Job, 0)
	for _, job := range m.jobs {
		if username == job.User {
			jobs = append(jobs, job)
		}
	}
	return jobs
}

func (m *JobManager) runLogEntriesForUser(username string) []RunLogEntry {
	entries := make([]RunLogEntry, 0)
	for _, entry := range m.runLog {
		if username == entry.Job.User {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (m *JobManager) loadAllJobs() (int, error) {
	// get all users by reading passwd
	f, err := os.Open("/etc/passwd")
	if err != nil {
		common.ErrLogger.Printf("Failed to open /etc/passwd: %v\n", err)
		return 0, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	totalJobs := 0
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) > 0 {
			user := parts[0]
			nbr, err := m.loadJobsForUser(user)
			totalJobs += nbr
			if err != nil {
				common.ErrLogger.Printf("Failed to load jobs for %s: %v\n", err, user)
			}
		}
	}

	common.ErrLogger.Printf("totalJobs: %v; len(m.jobs): %v", totalJobs, len(m.jobs))

	return len(m.jobs), nil
}

func (m *JobManager) reloadAllJobs() (int, error) {
	// stop job-runner thread and wait for current runs to end
	m.jobRunner.Cancel()
	for rec := range m.jobRunner.RunRecChan() {
		m.handleRunRec(rec)
	}
	m.jobRunner.Wait()

	// remove jobs
	amt := len(m.jobs)
	m.jobs = make([]*jobfile.Job, 0)

	// reload jobs
	amt, err := m.loadAllJobs()

	// restart job-runner thread
	m.jobRunner.Start(m.jobs, m.Shell, m.mainThreadCtx)

	return amt, err
}

func (m *JobManager) reloadJobsForUser(username string) (int, error) {
	// stop job-runner thread and wait for current runs to end
	m.jobRunner.Cancel()
	for rec := range m.jobRunner.RunRecChan() {
		m.handleRunRec(rec)
	}
	m.jobRunner.Wait()

	// remove user's jobs
	newJobList := make([]*jobfile.Job, 0)
	for _, job := range m.jobs {
		if job.User != username {
			newJobList = append(newJobList, job)
		}
	}
	m.jobs = newJobList

	// reload user's jobs
	amt, err := m.loadJobsForUser(username)

	// restart job-runner thread
	m.jobRunner.Start(m.jobs, m.Shell, m.mainThreadCtx)

	return amt, err
}

func (m *JobManager) loadJobsForUser(username string) (int, error) {
	// read .jobber file
	jobberFile, err := jobfile.LoadJobberFileForUser(username)
	if err != nil {
		return -1, err
	}
	m.userPrefs[username] = jobberFile.Prefs
	m.jobs = append(m.jobs, jobberFile.Jobs...)

	return len(jobberFile.Jobs), nil
}

func (m *JobManager) handleRunRec(rec *jobfile.RunRec) {
	if rec.Err != nil {
		common.ErrLogger.Panicln(rec.Err)
	}

	m.runLog = append(m.runLog,
		RunLogEntry{rec.Job, rec.RunTime, rec.Succeeded, rec.NewStatus})

	/* NOTE: error-handler was already applied by the job, if necessary. */

	if (!rec.Succeeded && rec.Job.NotifyOnError) ||
		(rec.Job.NotifyOnFailure && rec.NewStatus == jobfile.JobFailed) || (rec.Succeeded && rec.Job.NotifyOnSuccess) {
		// notify user
		m.userPrefs[rec.Job.User].Notifier(rec)
	}
}

func (m *JobManager) runMainThread() {
	m.mainThreadCtx, m.mainThreadCtl = NewJobberContext(BackgroundJobberContext())

	go func() {
		/*
		   All modifications to the job manager's state occur here.
		*/

		// start job-runner thread
		m.jobRunner.Start(m.jobs, m.Shell, m.mainThreadCtx)

	Loop:
		for {
			select {
			case <-m.mainThreadCtx.Done():
				break Loop

			case rec, ok := <-m.jobRunner.RunRecChan():
				if ok {
					m.handleRunRec(rec)
				} else {
					common.ErrLogger.Printf("jobfile.Job-runner thread ended prematurely.\n")
					break Loop
				}

			case cmd, ok := <-m.cmdChan:
				if ok {
					//fmt.Printf("JobManager: processing cmd.\n")
					shouldStop := m.doCmd(cmd)
					if shouldStop {
						break Loop
					}
				} else {
					common.ErrLogger.Printf("Command channel was closed.\n")
					break Loop
				}
			}
		}

		// cancel main thread
		m.mainThreadCtl.Cancel()

		// consume all run-records
		for rec := range m.jobRunner.RunRecChan() {
			m.handleRunRec(rec)
		}

		// finish up (and wait for job-runner thread to finish)
		m.mainThreadCtx.Finish()
	}()
}

func (m *JobManager) doCmd(cmd ICmd) bool { // runs in main thread

	/*
	   Security:

	   It is jobberd's responsibility to enforce the security policy.

	   It does so by assuming that cmd.RequestingUser() is truly the name
	   of the requesting user.
	*/

	switch cmd.(type) {
	case *ReloadCmd:
		/* Policy: Only root can reload other users' jobfiles. */

		// load jobs
		var err error
		var amt int
		if cmd.(*ReloadCmd).ForAllUsers {
			if cmd.RequestingUser() != "root" {
				cmd.RespChan() <- &ErrorCmdResp{&common.Error{What: "You must be root."}}
				break
			}

			amt, err = m.reloadAllJobs()
		} else {
			amt, err = m.reloadJobsForUser(cmd.RequestingUser())
		}

		// send response
		if err != nil {
			common.ErrLogger.Printf("Failed to load jobs: %v.\n", err)
			cmd.RespChan() <- &ErrorCmdResp{err}
		} else {
			cmd.RespChan() <- &SuccessCmdResp{fmt.Sprintf("Loaded %v jobs.", amt)}
		}

		return false

	case *CatCmd:
		/* Policy: Only root can cat other users' jobs. */

		var catCmd *CatCmd = cmd.(*CatCmd)

		// enforce policy
		if catCmd.jobUser != catCmd.RequestingUser() && catCmd.RequestingUser() != "root" {
			cmd.RespChan() <- &ErrorCmdResp{&common.Error{What: "You must be root."}}
			break
		}

		// find job to cat
		var job_p *jobfile.Job
		for _, job := range m.jobsForUser(catCmd.jobUser) {
			if job.Name == catCmd.job {
				job_p = job
				break
			}
		}
		if job_p == nil {
			msg := fmt.Sprintf("No job named \"%v\".", catCmd.job)
			cmd.RespChan() <- &ErrorCmdResp{&common.Error{What: msg}}
			break
		}

		// make and send response
		cmd.RespChan() <- &SuccessCmdResp{job_p.Cmd}

		return false

	case *ListJobsCmd:
		/* Policy: Only root can list other users' jobs. */

		// get jobs
		var jobs []*jobfile.Job
		if cmd.(*ListJobsCmd).ForAllUsers {
			if cmd.RequestingUser() != "root" {
				cmd.RespChan() <- &ErrorCmdResp{&common.Error{What: "You must be root."}}
				break
			}

			jobs = m.jobs
		} else {
			jobs = m.jobsForUser(cmd.RequestingUser())
		}

		// make response
		var buffer bytes.Buffer
		var writer *tabwriter.Writer = tabwriter.NewWriter(&buffer, 5, 0, 2, ' ', 0)
		fmt.Fprintf(writer, "NAME\tSTATUS\tSEC/MIN/HR/MDAY/MTH/WDAY\tNEXT RUN TIME\tNOTIFY ON ERR\tNOTIFY ON FAIL\tNOTIFY ON SUCCESS\tERR HANDLER\n")
		strs := make([]string, 0, len(m.jobs))
		for _, j := range jobs {
			schedStr := fmt.Sprintf("%v %v %v %v %v %v",
				j.FullTimeSpec.Sec,
				j.FullTimeSpec.Min,
				j.FullTimeSpec.Hour,
				j.FullTimeSpec.Mday,
				j.FullTimeSpec.Mon,
				j.FullTimeSpec.Wday)
			var runTimeStr string = "none"
			if j.NextRunTime != nil && !j.Paused {
				runTimeStr = j.NextRunTime.Format("Jan _2 15:04:05 2006")
			}
			statusStr := j.Status.String()
			if j.Paused {
				statusStr += " (Paused)"
			}
			s := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v",
				j.Name,
				statusStr,
				schedStr,
				runTimeStr,
				j.NotifyOnError,
				j.NotifyOnFailure,
				j.NotifyOnSuccess,
				j.ErrorHandler)
			strs = append(strs, s)
		}
		fmt.Fprintf(writer, "%v", strings.Join(strs, "\n"))
		writer.Flush()

		// send response
		cmd.RespChan() <- &SuccessCmdResp{buffer.String()}

		return false

	case *ListHistoryCmd:
		/* Policy: Only root can see the histories of other users' jobs. */

		// get log entries
		var entries []RunLogEntry
		if cmd.(*ListHistoryCmd).ForAllUsers {
			if cmd.RequestingUser() != "root" {
				cmd.RespChan() <- &ErrorCmdResp{&common.Error{What: "You must be root."}}
				break
			}

			entries = m.runLog
		} else {
			entries = m.runLogEntriesForUser(cmd.RequestingUser())
		}
		sort.Sort(&runLogEntrySorter{entries})

		// make response
		var buffer bytes.Buffer
		var writer *tabwriter.Writer = tabwriter.NewWriter(&buffer, 5, 0, 2, ' ', 0)
		fmt.Fprintf(writer, "TIME\tJOB\tUSER\tSUCCEEDED\tRESULT\t\n")
		strs := make([]string, 0, len(m.jobs))
		for _, e := range entries {
			s := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t", e.Time, e.Job.Name, e.Job.User, e.Succeeded, e.Result)
			strs = append(strs, s)
		}
		fmt.Fprintf(writer, "%v", strings.Join(strs, "\n"))
		writer.Flush()

		// send response
		cmd.RespChan() <- &SuccessCmdResp{buffer.String()}

		return false

	case *StopCmd:
		/* Policy: Only root can stop jobberd. */

		if cmd.RequestingUser() != "root" {
			cmd.RespChan() <- &ErrorCmdResp{&common.Error{What: "You must be root."}}
			break
		}

		common.Logger.Println("Stopping.")
		return true

	case *TestCmd:
		/* Policy: Only root can test other users' jobs. */

		var testCmd *TestCmd = cmd.(*TestCmd)

		// enforce policy
		if testCmd.jobUser != testCmd.RequestingUser() && testCmd.RequestingUser() != "root" {
			cmd.RespChan() <- &ErrorCmdResp{&common.Error{What: "You must be root."}}
			break
		}

		// find job to test
		var job_p *jobfile.Job
		for _, job := range m.jobsForUser(testCmd.jobUser) {
			if job.Name == testCmd.job {
				job_p = job
				break
			}
		}
		if job_p == nil {
			msg := fmt.Sprintf("No job named \"%v\".", testCmd.job)
			cmd.RespChan() <- &ErrorCmdResp{&common.Error{What: msg}}
			break
		}

		// run the job in this thread
		runRec := RunJob(job_p, nil, m.Shell, true)

		// send response
		if runRec.Err != nil {
			cmd.RespChan() <- &ErrorCmdResp{runRec.Err}
			break
		}
		cmd.RespChan() <- &SuccessCmdResp{Details: runRec.Describe()}

		return false

	case *PauseCmd:
		/* Policy: Users can pause only their own jobs. */

		var pauseCmd *PauseCmd = cmd.(*PauseCmd)

		// look up jobs to pause
		var usersJobs []*jobfile.Job = m.jobsForUser(pauseCmd.RequestingUser())
		jobsToPause := make([]*jobfile.Job, 0)
		if len(pauseCmd.jobs) > 0 {
			for _, jobName := range pauseCmd.jobs {
				foundJob := false
				for _, job := range usersJobs {
					if job.Name == jobName {
						jobsToPause = append(jobsToPause, job)
						foundJob = true
						break
					}
				}
				if !foundJob {
					msg := fmt.Sprintf("No job named \"%v\".", jobName)
					cmd.RespChan() <- &ErrorCmdResp{&common.Error{What: msg}}
					return false
				}
			}
		} else {
			jobsToPause = usersJobs
		}

		// pause them
		amtPaused := 0
		for _, job := range jobsToPause {
			if !job.Paused {
				job.Paused = true
				amtPaused += 1
			}
		}

		// make and send response
		cmd.RespChan() <- &SuccessCmdResp{fmt.Sprintf("Paused %v jobs.", amtPaused)}
		return false

	case *ResumeCmd:
		/* Policy: Users can pause only their own jobs. */

		var resumeCmd *ResumeCmd = cmd.(*ResumeCmd)

		// look up jobs to pause
		var usersJobs []*jobfile.Job = m.jobsForUser(resumeCmd.RequestingUser())
		jobsToResume := make([]*jobfile.Job, 0)
		if len(resumeCmd.jobs) > 0 {
			for _, jobName := range resumeCmd.jobs {
				foundJob := false
				for _, job := range usersJobs {
					if job.Name == jobName {
						jobsToResume = append(jobsToResume, job)
						foundJob = true
						break
					}
				}
				if !foundJob {
					msg := fmt.Sprintf("No job named \"%v\".", jobName)
					cmd.RespChan() <- &ErrorCmdResp{&common.Error{What: msg}}
					return false
				}
			}
		} else {
			jobsToResume = usersJobs
		}

		// pause them
		amtResumed := 0
		for _, job := range jobsToResume {
			if job.Paused {
				job.Paused = false
				amtResumed += 1
			}
		}

		// make and send response
		cmd.RespChan() <- &SuccessCmdResp{fmt.Sprintf("Resumed %v jobs.", amtResumed)}
		return false

	default:
		cmd.RespChan() <- &ErrorCmdResp{&common.Error{What: "Unknown command."}}
		return false
	}

	return false
}
