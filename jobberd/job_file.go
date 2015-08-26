package main

import (
    "io"
    "io/ioutil"
    "bufio"
    "gopkg.in/yaml.v2"
    "path/filepath"
    "os"
    "syscall"
    "os/user"
    "strconv"
    "fmt"
    "strings"
)

const (
    JobberFileName = ".jobber"
    TimeWildcard = "*"
)

type JobConfigEntry struct {
    Name             string
    Cmd              string
    Time             string
    OnError          *string "onError,omitempty"
    NotifyOnError    *bool   "notifyOnError,omitempty"
    NotifyOnFailure  *bool   "notifyOnFailure,omitempty"
}

func openUsersJobberFile(username string) (*os.File, error) {
    /*
     * Not all users listed in /etc/passwd have their own
     * jobber file.  E.g., some of them may share a home dir.
     * When this happens, we say that the jobber file belongs
     * to the user who owns that file.
     */
     
     // make path to jobber file
    user, err := user.Lookup(username)
    if err != nil {
        return nil, err
    }
    jobberFilePath := filepath.Join(user.HomeDir, JobberFileName)
    
    // open it
    f, err := os.Open(jobberFilePath)
    if err != nil {
        return nil, err
    }
    
    // check owner
    info, err := f.Stat()
    if err != nil {
        f.Close()
        return nil, err
    }
    uid, err := strconv.Atoi(user.Uid)
    if err != nil {
        f.Close()
        return nil, err
    }
    if uint32(uid) != info.Sys().(*syscall.Stat_t).Uid {
        f.Close()
        return nil, os.ErrNotExist
    }
    
    return f, nil
}

func (m *JobManager) LoadAllJobs() (int, error) {
    // get all users by reading passwd
    f, err := os.Open("/etc/passwd")
    if err != nil {
        ErrLogger.Printf("Failed to open /etc/passwd: %v\n", err)
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
                ErrLogger.Printf("Failed to load jobs for %s: %v\n", err, user)
            }
        }
    }
    
    ErrLogger.Printf("totalJobs: %v; len(m.jobs): %v", totalJobs, len(m.jobs));
    
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

func (m *JobManager) loadJobsForUser(username string) (int, error) {
    // read .jobber file
    var newJobs []*Job
    f, err := openUsersJobberFile(username)
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
    Logger.Printf("Loaded %v new jobs for %s.\n", len(newJobs), username)
    
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
        if config.OnError != nil {
            job.ErrorHandler, err = getErrorHandler(*config.OnError)
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
        
        job.Sec = WildcardTimeSpec
        job.Min = WildcardTimeSpec
        job.Hour = WildcardTimeSpec
        job.Mday = WildcardTimeSpec
        job.Mon = WildcardTimeSpec
        job.Wday = WildcardTimeSpec
        
        var timeParts []string = strings.Fields(config.Time)
        
        // sec
        if len(timeParts) > 0 {
            spec, err := parseTimeStr(timeParts[0], "sec", 0, 59)
            if err != nil {
                return nil, err
            }
            job.Sec = spec
        }
        
        // min
        if len(timeParts) > 1 {
            spec, err := parseTimeStr(timeParts[1], "minute", 0, 59)
            if err != nil {
                return nil, err
            }
            job.Min = spec
        }
        
        // hour
        if len(timeParts) > 2 {
            spec, err := parseTimeStr(timeParts[2], "hour", 0, 23)
            if err != nil {
                return nil, err
            }
            job.Hour = spec
        }
        
        // mday
        if len(timeParts) > 3 {
            spec, err := parseTimeStr(timeParts[3], "month day", 1, 31)
            if err != nil {
                return nil, err
            }
            job.Mday = spec
        }
        
        // month
        if len(timeParts) > 4 {
            spec, err := parseTimeStr(timeParts[4], "month", 1, 12)
            if err != nil {
                return nil, err
            }
            job.Mon = spec
        }
        
        // wday
        if len(timeParts) > 5 {
            spec, err := parseTimeStr(timeParts[5], "weekday", 0, 6)
            if err != nil {
                return nil, err
            }
            job.Wday = spec
        }
        
        if len(timeParts) > 6 {
            return nil, &JobberError{"Excess elements in 'time' field.", nil}
        }
        
        jobs = append(jobs, job)
    }
    return jobs, nil
}

type OneValTimeSpec struct {
    desc string
    val *int
}

func (s OneValTimeSpec) String() string {
    return s.desc
}

func (s OneValTimeSpec) Satisfied(v int) bool {
    if s.val == nil {
        return true
    } else {
        return *s.val == v
    }
}

type SetTimeSpec struct {
    desc string
    vals []int
}

func (s SetTimeSpec) String() string {
    return s.desc
}

func (s SetTimeSpec) Satisfied(v int) bool {
    for _, v2 := range s.vals {
        if v == v2 {
            return true
        }
    }
    return false
}

var WildcardTimeSpec OneValTimeSpec = OneValTimeSpec{val: nil, desc: "*"}

func parseTimeStr(s string, fieldName string, min int, max int) (TimeSpec, error) {
    errMsg := fmt.Sprintf("Invalid '%v' value", fieldName)
    
    if s == TimeWildcard {
        return WildcardTimeSpec, nil
    } else if strings.HasPrefix(s, "*/") {
        // parse step
        stepStr := s[2:]
        step, err := strconv.Atoi(stepStr)
        if err != nil {
            return nil, &JobberError{errMsg, err}
        }
        
        // make set of valid values
        vals := make([]int, 0)
        for v := min; v <= max; v = v + step {
            vals = append(vals, v)
        }
        
        // make spec
        spec := SetTimeSpec{vals: vals, desc: s}
        return spec, nil
        
    } else {
        // convert to int
        val, err := strconv.Atoi(s)
        if err != nil {
            return nil, &JobberError{errMsg, err}
        }
        
        // make TimeSpec
        spec := OneValTimeSpec{val: &val, desc: s}
        
        // check range
        if val < min {
            errMsg := fmt.Sprintf("%s: cannot be less than %v.", errMsg, min)
            return nil, &JobberError{errMsg, nil}
        } else if val > max {
            errMsg := fmt.Sprintf("%s: cannot be greater than %v.", errMsg, max)
            return nil, &JobberError{errMsg, nil}
        }
        
        return spec, nil
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
