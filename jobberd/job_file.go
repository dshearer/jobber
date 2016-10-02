package main

import (
	"bufio"
	"fmt"
	"github.com/dshearer/jobber/Godeps/_workspace/src/gopkg.in/yaml.v2"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

const (
	JobberFileName = ".jobber"
    PrefsSectName = "prefs"
    JobsSectName = "jobs"
	TimeWildcard   = "*"
)

type JobberFile struct {
    Prefs    UserPrefs
    Jobs     []*Job
}

type UserPrefs struct {
	Notifier	    RunRecNotifier
}

type JobConfigEntry struct {
	Name            string
	Cmd             string
	Time            string
	OnError         *string "onError,omitempty"
	NotifyOnError   *bool   "notifyOnError,omitempty"
	NotifyOnFailure *bool   "notifyOnFailure,omitempty"
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

func readJobberFile(r io.Reader, username string) (*JobberFile, error) {
    /*
    Jobber files have two sections: one begins with "[prefs]" on a line, and 
    the other begins with "[jobs]".  Both contain a YAML document.  The "prefs"
    section can be parsed with struct UserPrefs, and the "jobs" section is a 
    YAML array of records that can be parsed with struct JobConfigEntry.
    
    Legacy format: no section beginnings; whole file is YAML doc for "jobs"
    section.
    */
    
    // iterate over lines
    scanner := bufio.NewScanner(r)
    sectionsToLines := make(map[string][]string)
    var currSection *string
    lineNbr := 0
    sectNameRegexp := regexp.MustCompile("^\\[(\\w*)\\]\\s*$")
    legacyFormat := false
    scanner.Split(bufio.ScanLines)
    for scanner.Scan() {
        lineNbr++
        line := scanner.Text()
        
        if legacyFormat {
            // save line
            sectionsToLines[*currSection] = 
                append(sectionsToLines[*currSection], line)
                
        } else {
            // check whether line begins a section
            var matches []string = sectNameRegexp.FindStringSubmatch(line)
            if matches != nil {
                // we are entering a (new) section
                sectName := matches[1]
                _, ok := sectionsToLines[sectName]
                if ok {
                    errMsg := 
                        fmt.Sprintf("Line %v: another section called \"%v\".", 
                                    lineNbr,
                                    sectName)
                    return nil, &JobberError{errMsg, nil}
                }
                sectionsToLines[sectName] = make([]string, 0)
                currSection = &sectName
                
            } else if currSection == nil {
                if len(strings.TrimSpace(line)) > 0 {
                    /*
                    To support legacy format, treat whole file as YAML doc
                    for "jobs" section.
                    */
                    Logger.Println("Using legacy jobber file format.")
                    legacyFormat = true
                    tmp := JobsSectName
                    currSection = &tmp
                    sectionsToLines[JobsSectName] = make([]string, 1)
                    sectionsToLines[JobsSectName][0] = line
                }
                
            } else {
                // save line
                sectionsToLines[*currSection] = 
                    append(sectionsToLines[*currSection], line)
            }
        }
    }
    
    // parse "prefs" section
    const yamlStarter string = "---"
    rawPrefs := map[string]interface{} {}
    prefsLines, prefsOk := sectionsToLines[PrefsSectName]
    if prefsOk && len(prefsLines) > 0 {
        Logger.Println("Got prefs section")
        prefsSection := strings.Join(prefsLines, "\n")
        if strings.TrimRight(prefsLines[0], " \t") != yamlStarter {
            prefsSection = yamlStarter + "\n" + prefsSection
        }
	    err := yaml.Unmarshal([]byte(prefsSection), &rawPrefs)
	    if err != nil {
	        errMsg := fmt.Sprintf("Failed to parse \"%v\" section", 
	                              PrefsSectName)
	        return nil, &JobberError{errMsg, err}
	    }
    }
    
    // parse "jobs" section
    var jobConfigs []JobConfigEntry
    jobsLines, jobsOk := sectionsToLines[JobsSectName]
    if jobsOk && len(jobsLines) > 0 {
        Logger.Println("Got jobs section")
        jobsSection := strings.Join(jobsLines, "\n")
        if strings.TrimRight(jobsLines[0], " \t") != yamlStarter {
            jobsSection = yamlStarter + "\n" + jobsSection
        }
	    err := yaml.Unmarshal([]byte(jobsSection), &jobConfigs)
	    if err != nil {
	        errMsg := fmt.Sprintf("Failed to parse \"%v\" section", 
	                              JobsSectName)
	        return nil, &JobberError{errMsg, err}
	    }
    }
    
    // make prefs
    var userPrefs UserPrefs
    noteProgVal, hasNoteProg := rawPrefs["notifyProgram"]
    if hasNoteProg {
        noteProgValStr, ok := noteProgVal.(string)
        if !ok {
	        errMsg := fmt.Sprintf("Invalid value for preference \"notifyProgram\": %v", 
	                              noteProgVal)
	        return nil, &JobberError{errMsg, nil}
        }
        userPrefs.Notifier = MakeProgramNotifier(noteProgValStr)
    } else {
        userPrefs.Notifier = MakeMailNotifier()
    }
    
    // make jobs
    jobs := make([]*Job, 0, len(jobConfigs))
    for _, config := range jobConfigs {
        job := NewJob(config.Name, config.Cmd, username)
        var err error = nil
        
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
        
        // parse time spec
        var tmp *FullTimeSpec
        tmp, err = parseFullTimeSpec(config.Time)
        if err != nil {
            return nil, err
        }
        job.FullTimeSpec = *tmp
        
        jobs = append(jobs, job)
    }
    
    return &JobberFile{userPrefs, jobs}, nil
}

type WildcardTimeSpec struct {
}

func (s WildcardTimeSpec) String() string {
	return "*"
}

func (s WildcardTimeSpec) Satisfied(v int) bool {
	return true
}

type OneValTimeSpec struct {
	val int
}

func (s OneValTimeSpec) String() string {
	return fmt.Sprintf("%v", s.val)
}

func (s OneValTimeSpec) Satisfied(v int) bool {
	return s.val == v
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

func parseFullTimeSpec(s string) (*FullTimeSpec, error) {
	var fullSpec FullTimeSpec
	fullSpec.Sec = WildcardTimeSpec{}
	fullSpec.Min = WildcardTimeSpec{}
	fullSpec.Hour = WildcardTimeSpec{}
	fullSpec.Mday = WildcardTimeSpec{}
	fullSpec.Mon = WildcardTimeSpec{}
	fullSpec.Wday = WildcardTimeSpec{}

	var timeParts []string = strings.Fields(s)

	// sec
	if len(timeParts) > 0 {
		spec, err := parseTimeSpec(timeParts[0], "sec", 0, 59)
		if err != nil {
			return nil, err
		}
		fullSpec.Sec = spec
	}

	// min
	if len(timeParts) > 1 {
		spec, err := parseTimeSpec(timeParts[1], "minute", 0, 59)
		if err != nil {
			return nil, err
		}
		fullSpec.Min = spec
	}

	// hour
	if len(timeParts) > 2 {
		spec, err := parseTimeSpec(timeParts[2], "hour", 0, 23)
		if err != nil {
			return nil, err
		}
		fullSpec.Hour = spec
	}

	// mday
	if len(timeParts) > 3 {
		spec, err := parseTimeSpec(timeParts[3], "month day", 1, 31)
		if err != nil {
			return nil, err
		}
		fullSpec.Mday = spec
	}

	// month
	if len(timeParts) > 4 {
		spec, err := parseTimeSpec(timeParts[4], "month", 1, 12)
		if err != nil {
			return nil, err
		}
		fullSpec.Mon = spec
	}

	// wday
	if len(timeParts) > 5 {
		spec, err := parseTimeSpec(timeParts[5], "weekday", 0, 6)
		if err != nil {
			return nil, err
		}
		fullSpec.Wday = spec
	}

	if len(timeParts) > 6 {
		return nil, &JobberError{"Excess elements in 'time' field.", nil}
	}

	return &fullSpec, nil
}

func parseTimeSpec(s string, fieldName string, min int, max int) (TimeSpec, error) {
	errMsg := fmt.Sprintf("Invalid '%v' value", fieldName)

	if s == TimeWildcard {
		return WildcardTimeSpec{}, nil
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

	} else if strings.Contains(s, ",") {
		// split step
		stepStrs := strings.Split(s, ",")

		// make set of valid values
		vals := make([]int, 0)
		for _,stepStr := range stepStrs {
			step, err := strconv.Atoi(stepStr)
			if err != nil {
				return nil, &JobberError{errMsg, err}
			}
			vals = append(vals, step)
		}

		// make spec
		spec := SetTimeSpec{vals: vals, desc: s}
		return spec, nil
	} else if strings.Contains(s, "-") {
		// get range extremes
		extremes := strings.Split(s, "-")
		begin, err := strconv.Atoi(extremes[0])

		if err != nil {
			return nil, &JobberError{errMsg, err}
		}

		end, err := strconv.Atoi(extremes[1])

		if err != nil {
			return nil, &JobberError{errMsg, err}
		}

		// make set of valid values
		vals := make([]int, 0)

		for v := begin; v <= end; v++ {
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
		spec := OneValTimeSpec{val}

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
	case ErrorHandlerStopName:
		return &ErrorHandlerStop, nil
	case ErrorHandlerBackoffName:
		return &ErrorHandlerBackoff, nil
	case ErrorHandlerContinueName:
		return &ErrorHandlerContinue, nil
	default:
		return nil, &JobberError{"Invalid error handler: " + name, nil}
	}
}
