package main

import (
    "io"
    "io/ioutil"
    "encoding/json"
    "fmt"
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
    Time             TimeSpec
    OnError          string
    NotifyOnError    *bool
    NotifyOnFailure  *bool
}

type TimeSpec struct {
    Sec  *int
    Min  *int
    Hour *int
    Mday *int
    Mon  *int
    Wday *int
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
            _, err = m.LoadJobsForUser(username)
            if err != nil {
                m.errorLogger.Printf("Failed to load jobs for %v: %v.\n", username, err)
            }
        }
        
        return filepath.SkipDir
    } else {
        return nil
    }
}

func (m *JobManager) LoadAllJobs() (int, error) {
    // load jobs for normal users
    err := filepath.Walk(HomeDirRoot, m.procHomeFile)
    if err != nil {
        return -1, err
    }
    
    // load jobs for root
    _, err = m.LoadJobsForUser("root")
    if err != nil {
        return -1, err
    } else {
        return len(m.jobs), nil
    }
}

func (m *JobManager) ReloadAllJobs() (int, error) {
    // remove jobs
    amt := len(m.jobs)
    m.jobs = make([]*Job, 0)
    m.logger.Printf("Removed %v jobs.\n", amt)
    
    // reload jobs
    return m.LoadAllJobs()
}

func (m *JobManager) ReloadJobsForUser(username string) (int, error) {
    // remove user's jobs
    newJobList := make([]*Job, 0)
    for _, job := range m.jobs {
        if job.User != username {
            newJobList = append(newJobList, job)
        }
    }
    m.logger.Printf("Removed %v jobs.\n", len(m.jobs) - len(newJobList))
    m.jobs = newJobList
    
    // reload user's jobs
    return m.LoadJobsForUser(username)
}

func (m *JobManager) LoadJobsForUser(username string) (int, error) {
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
    m.logger.Printf("Loaded %v new jobs.\n", len(newJobs))
    
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
            job.Sec, err = makeTimePred(*config.Time.Sec)
            if err != nil {
                return nil, err
            }
        }
        
        // min
        if config.Time.Min != nil {
            job.Min, err = makeTimePred(*config.Time.Min)
            if err != nil {
                return nil, err
            }
        }
        
        // hour
        if config.Time.Hour != nil {
            job.Hour, err = makeTimePred(*config.Time.Hour)
            if err != nil {
                return nil, err
            }
        }
        
        // mday
        if config.Time.Mday != nil {
            job.Mday, err = makeTimePred(*config.Time.Mday)
            if err != nil {
                return nil, err
            }
        }
        
        // month
        if config.Time.Mon != nil {
            job.Mon, err = makeTimePred(*config.Time.Mon)
            if err != nil {
                return nil, err
            }
        }
        
        // wday
        if config.Time.Wday != nil {
            job.Wday, err = makeTimePred(*config.Time.Wday)
            if err != nil {
                return nil, err
            }
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

func makeTimePred(v int) (TimePred, error) {
    return TimePred{func(i int) bool { return i == v }, fmt.Sprintf("%v", v)}, nil
}
