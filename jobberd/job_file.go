package main

import (
    "io"
    "io/ioutil"
    "encoding/json"
    "path/filepath"
    "os"
    "os/user"
)

const (
    HomeDirRoot    = "/home"
    RootHomeDir = "/root"
    JobberFileName = ".jobber"
)

type JobConfigEntry struct {
    Name             string
    Cmd              string
    Time             ConfigTimeSpec
    OnError          string
    NotifyOnError    *bool
    NotifyOnFailure  *bool
}

type ConfigTimeSpec struct {
    Sec  *TimeSpec
    Min  *TimeSpec
    Hour *TimeSpec
    Mday *TimeSpec
    Mon  *TimeSpec
    Wday *TimeSpec
}

func (m *JobManager) LoadAllJobs() (int, error) {
    // load jobs for normal users
    err := filepath.Walk(HomeDirRoot, m.procHomeFile)
    if err != nil {
        return -1, err
    }
    
    // load jobs for root
    _, err = m.loadJobsForUser("root")
    if err != nil {
        return -1, err
    } else {
        return len(m.jobs), nil
    }
}

func (m *JobManager) ReloadAllJobs() (int, error) {
    // stop job-runner thread and wait for current runs to end
    m.jobRunner.Cancel()
    for rec := range m.jobRunner.RunRecChan() {
        m.handleRunRec(rec)
    }
    m.jobRunner.Wait()
    
    // remove jobs
    amt := len(m.jobs)
    m.jobs = make([]*Job, 0)
    Logger.Printf("Removed %v jobs.\n", amt)
    
    // reload jobs
    amt, err := m.LoadAllJobs()
    
    // restart job-runner thread
    m.jobRunner.Start(m.jobs, m.Shell, m.mainThreadCtx)
    
    return amt, err
}

func (m *JobManager) ReloadJobsForUser(username string) (int, error) {
    // stop job-runner thread and wait for current runs to end
    m.jobRunner.Cancel()
    for rec := range m.jobRunner.RunRecChan() {
        m.handleRunRec(rec)
    }
    m.jobRunner.Wait()
    
    // remove user's jobs
    newJobList := make([]*Job, 0)
    for _, job := range m.jobs {
        if job.User != username {
            newJobList = append(newJobList, job)
        }
    }
    Logger.Printf("Removed %v jobs.\n", len(m.jobs) - len(newJobList))
    m.jobs = newJobList
    
    // reload user's jobs
    amt, err := m.loadJobsForUser(username)
    
    // restart job-runner thread
    m.jobRunner.Start(m.jobs, m.Shell, m.mainThreadCtx)
    
    return amt, err
}

func (m *JobManager) procHomeFile(path string, info os.FileInfo, err error) error {
    if err != nil {
        return err
    } else if path == HomeDirRoot {
        return nil
    } else if info.IsDir() {
        username := filepath.Base(path)
        
        // check whether this user exists
        _, err = user.Lookup(username)
        if err == nil {
            /* User exists. */
            _, err = m.loadJobsForUser(username)
            if err != nil {
                ErrLogger.Printf("Failed to load jobs for %v: %v.\n", username, err)
            }
        }
        
        return filepath.SkipDir
    } else {
        return nil
    }
}

func (m *JobManager) loadJobsForUser(username string) (int, error) {
    // compute .jobber file path
    var jobberFilePath string
    if username == "root" {
        jobberFilePath = filepath.Join(RootHomeDir, JobberFileName)
    } else {
        jobberFilePath = filepath.Join(HomeDirRoot, username, JobberFileName)
    }
    
    // read .jobber file
    var newJobs []*Job
    f, err := os.Open(jobberFilePath)
    if err != nil {
        if os.IsNotExist(err) {
            newJobs = make([]*Job, 0)
        } else {
            return -1, err
        }
    } else {
        defer f.Close()
        newJobs, err = readJobFile(f, username)
        if err != nil {
            return -1, err
        }
    }
    m.jobs = append(m.jobs, newJobs...)
    Logger.Printf("Loaded %v new jobs.\n", len(newJobs))
    
    return len(newJobs), nil
}

func readJobFile(r io.Reader, username string) ([]*Job, error) {
    // read config file
    data, err := ioutil.ReadAll(r)
    if err != nil {
        return nil, err
    }
    var configs []JobConfigEntry
    err = json.Unmarshal(data, &configs)
    if err != nil {
        return nil, err
    }
    
    // make jobs
    jobs := make([]*Job, 0, len(configs))
    for _, config := range configs {
        job := NewJob(config.Name, config.Cmd, username)
        
        // check name
        if len(config.Name) == 0 {
            return nil, &JobberError{"Job name cannot be empty.", nil}
        }
        
        // set failure-handler
        if len(config.OnError) > 0 {
            job.ErrorHandler, err = getErrorHandler(config.OnError)
            if err != nil {
                return nil, err
            }
        }
        
        // set notify prefs
        if config.NotifyOnError != nil {
            job.NotifyOnError = *config.NotifyOnError
        }
        if config.NotifyOnFailure != nil {
            job.NotifyOnFailure = *config.NotifyOnFailure
        }
        
        // sec
        if config.Time.Sec != nil {
            if *config.Time.Sec < 0 || *config.Time.Sec > 59 {
                return nil, &JobberError{"Invalid 'sec' value.", nil}
            }
            job.Sec = *config.Time.Sec
        }
        
        // min
        if config.Time.Min != nil {
            if *config.Time.Min < 0 || *config.Time.Min > 59 {
                return nil, &JobberError{"Invalid 'min' value.", nil}
            }
            job.Min = *config.Time.Min
        }
        
        // hour
        if config.Time.Hour != nil {
            if *config.Time.Hour < 0 || *config.Time.Hour > 23 {
                return nil, &JobberError{"Invalid 'hour' value.", nil}
            }
            job.Hour = *config.Time.Hour
        }
        
        // mday
        if config.Time.Mday != nil {
            if *config.Time.Mday < 1 || *config.Time.Mday > 31 {
                return nil, &JobberError{"Invalid 'mday' value.", nil}
            }
            job.Mday = *config.Time.Mday
        }
        
        // month
        if config.Time.Mon != nil {
            if *config.Time.Mon < 1 || *config.Time.Mon > 12 {
                return nil, &JobberError{"Invalid 'mon' value.", nil}
            }
            job.Mon = *config.Time.Mon
        }
        
        // wday
        if config.Time.Wday != nil {
            if *config.Time.Wday < 0 || *config.Time.Wday > 6 {
                return nil, &JobberError{"Invalid 'wday' value.", nil}
            }
            job.Wday = *config.Time.Wday
        }
        
        jobs = append(jobs, job)
    }
    return jobs, nil
}

func getErrorHandler(name string) (*ErrorHandler, error) {
    switch name {
        case ErrorHandlerStopName: return &ErrorHandlerStop, nil
        case ErrorHandlerBackoffName: return &ErrorHandlerBackoff, nil
        case ErrorHandlerContinueName: return &ErrorHandlerContinue, nil
        default: return nil, &JobberError{"Invalid error handler: " + name, nil}
    }
}
