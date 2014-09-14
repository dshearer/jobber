package jobber

import (
    "io"
    "io/ioutil"
    "strconv"
    "encoding/json"
)

type JobConfigEntry struct {
    Name string
    Cmd string
    Time TimeSpec
}

type TimeSpec struct {
    Min string
    Hour string
    Mday string
    Mon string
    Wday string
}

func ReadJobFile(r io.Reader, username string) ([]*Job, error) {
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
        
        // min
        if len(config.Time.Min) > 0 && config.Time.Min != "*" {
            job.Min, err = strToTimePred(config.Time.Min)
            if err != nil {
                return nil, err
            }
        }
        
        // hour
        if len(config.Time.Hour) > 0 && config.Time.Hour != "*" {
            job.Hour, err = strToTimePred(config.Time.Hour)
            if err != nil {
                return nil, err
            }
        }
        
        // mday
        if len(config.Time.Mday) > 0 && config.Time.Mday != "*" {
            job.Mday, err = strToTimePred(config.Time.Mday)
            if err != nil {
                return nil, err
            }
        }
        
        // month
        if len(config.Time.Mon) > 0 && config.Time.Mon != "*" {
            job.Mon, err = strToTimePred(config.Time.Mon)
            if err != nil {
                return nil, err
            }
        }
        
        // wday
        if len(config.Time.Wday) > 0 && config.Time.Wday != "*" {
            job.Wday, err = strToTimePred(config.Time.Wday)
            if err != nil {
                return nil, err
            }
        }
        
        jobs = append(jobs, job)
    }
    return jobs, nil
}

func strToTimePred(s string) (TimePred, error) {
    v, err := strconv.Atoi(s)
    if err != nil {
        return TimePred{nil, ""}, err
    }
    return TimePred{func(i int) bool { return i == v }, s}, nil
}
