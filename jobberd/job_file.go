package main

import (
    "io"
    "io/ioutil"
    "gopkg.in/yaml.v2"
    "path/filepath"
    "os"
    "os/user"
    "strconv"
    "fmt"
    "strings"
)

const (
    HomeDirRoot    = "/home"
    RootHomeDir = "/root"
    JobberFileName = ".jobber"
    TimeWildcard = "*"
)

type JobConfigEntry struct {
    Name             string
    Cmd              string
    Time             string
    OnError          string
    NotifyOnError    *bool   "notifyOnError,omitempty"
    NotifyOnFailure  *bool   "notifyOnFailure,omitempty"
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
        ErrLogger.Printf("Failed to load jobs for root: %v\n", err)
    }
    return len(m.jobs), nil
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
    err = yaml.Unmarshal(data, &configs)
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
        
        var timeParts []string = strings.Fields(config.Time)
        
        // sec
        if len(timeParts) > 0 {
            val, err := parseTimeStr(timeParts[0], "sec", 0, 59)
            if err != nil {
                return nil, err
            }
            job.Sec.Value = val
        }
        
        // min
        if len(timeParts) > 1 {
            val, err := parseTimeStr(timeParts[1], "minute", 0, 59)
            if err != nil {
                return nil, err
            }
            job.Min.Value = val
        }
        
        // hour
        if len(timeParts) > 2 {
            val, err := parseTimeStr(timeParts[2], "hour", 0, 23)
            if err != nil {
                return nil, err
            }
            job.Hour.Value = val
        }
        
        // mday
        if len(timeParts) > 3 {
            val, err := parseTimeStr(timeParts[3], "month day", 1, 31)
            if err != nil {
                return nil, err
            }
            job.Mday.Value = val
        }
        
        // month
        if len(timeParts) > 4 {
            val, err := parseTimeStr(timeParts[4], "month", 1, 12)
            if err != nil {
                return nil, err
            }
            job.Mon.Value = val
        }
        
        // wday
        if len(timeParts) > 5 {
            val, err := parseTimeStr(timeParts[5], "weekday", 0, 6)
            if err != nil {
                return nil, err
            }
            job.Wday.Value = val
        }
        
        if len(timeParts) > 6 {
            return nil, &JobberError{"Excess elements in 'time' field.", nil}
        }
        
        jobs = append(jobs, job)
    }
    return jobs, nil
}

func parseTimeStr(s string, fieldName string, min uint, max uint) (*uint, error) {
    errMsg := fmt.Sprintf("Invalid '%v' value", fieldName)
    
    if s == TimeWildcard {
        return nil, nil
    } else {
        // convert to int
        val, err := strconv.Atoi(s)
        if err != nil {
            return nil, &JobberError{errMsg, err}
        }
        
        // check sign
        if val < 0 {
            errMsg := fmt.Sprintf("%s: cannot be negative.", errMsg)
            return nil, &JobberError{errMsg, nil}
        }
        uval := uint(val)
        
        // check range
        if uval < min {
            errMsg := fmt.Sprintf("%s: cannot be less than %v.", errMsg, min)
            return nil, &JobberError{errMsg, nil}
        } else if uval > max {
            errMsg := fmt.Sprintf("%s: cannot be greater than %v.", errMsg, max)
            return nil, &JobberError{errMsg, nil}
        }
        
        return &uval, nil
    }
}

func getErrorHandler(name string) (*ErrorHandler, error) {
    switch name {
        case ErrorHandlerStopName: return &ErrorHandlerStop, nil
        case ErrorHandlerBackoffName: return &ErrorHandlerBackoff, nil
        case ErrorHandlerContinueName: return &ErrorHandlerContinue, nil
        default: return nil, &JobberError{"Invalid error handler: " + name, nil}
    }
}
