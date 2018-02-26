package jobfile

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dshearer/jobber/common"
	"gopkg.in/yaml.v2"
)

const (
	PrefsSectName           = "prefs"
	JobsSectName            = "jobs"
	gYamlStarter            = "---"
	gDefaultMemRunLogMaxLen = 100
)

type JobFile struct {
	Prefs UserPrefs
	Jobs  []*Job
}

type UserPrefs struct {
	Notifier RunRecNotifier
	RunLog   RunLog
	LogPath  string // for error msgs etc.  May be "".

	// output handling
	StdoutHandler JobOutputHandler
	StderrHandler JobOutputHandler
}

func (self *UserPrefs) String() string {
	s := ""
	s += fmt.Sprintf("Notifier: %v\n", self.Notifier)
	s += fmt.Sprintf("RunLog: %v\n", self.RunLog)
	if len(self.LogPath) > 0 {
		s += fmt.Sprintf("Log path: %v\n", self.LogPath)
	}
	s += fmt.Sprintf("StdoutHandler: %v\n", self.StdoutHandler)
	s += fmt.Sprintf("StderrHandler: %v", self.StderrHandler)
	return s
}

type JobOutputPrefsRaw struct {
	Where      *string `yaml:"where"`
	MaxAgeDays *int    `yaml:"maxAgeDays"`
}

type BothJobOutputPrefsRaw struct {
	Stdout *JobOutputPrefsRaw `yaml:"stdout"`
	Stderr *JobOutputPrefsRaw `yaml:"stderr"`
}

type RunLogRaw struct {
	Type string `yaml:"type"` // "file" or "memory"

	// fields for type == "memory":
	MaxLen *int `yaml:"maxLen"`

	// fields for type == "file":
	Path         *string `yaml:"path"`
	MaxFileLen   *string `yaml:"maxFileLen"`
	MaxHistories *int    `yaml:"maxHistories"`
}

type UserPrefsRaw struct {
	LogPath       *string                `yaml:"logPath"`
	NotifyProgram *string                `yaml:"notifyProgram"`
	RunLog        *RunLogRaw             `yaml:"runLog"`
	JobOutput     *BothJobOutputPrefsRaw `yaml:"jobOutput"`
}

type JobRaw struct {
	Name            string
	Cmd             string
	Time            string
	OnError         *string                `yaml:"onError"`
	NotifyOnSuccess *bool                  `yaml:"notifyOnSuccess"`
	NotifyOnError   *bool                  `yaml:"notifyOnError"`
	NotifyOnFailure *bool                  `yaml:"notifyOnFailure"`
	JobOutput       *BothJobOutputPrefsRaw `yaml:"jobOutput"`
}

func NewEmptyJobFile() *JobFile {
	prefs := UserPrefs{
		RunLog: NewMemOnlyRunLog(gDefaultMemRunLogMaxLen),
	}
	return &JobFile{
		Prefs: prefs,
		Jobs:  nil,
	}
}

const gBadJobfilePerms os.FileMode = 0022

func ShouldLoadJobfile(f *os.File, usr *user.User) (bool, error) {
	// check jobfile's owner
	ownsFile, err := common.UserOwnsFileF(usr, f)
	if err != nil {
		return false, err
	}
	if !ownsFile {
		msg := fmt.Sprintf("User %v doesn't own jobfile", usr.Username)
		return false, &common.Error{What: msg}
	}

	// check jobfile's perms
	stat, err := f.Stat()
	if err != nil {
		return false, err
	}
	if stat.Mode().Perm()&gBadJobfilePerms > 0 {
		msg := fmt.Sprintf(
			"Jobfile has bad permissions: %v. Problematic perms: %v",
			stat.Mode().Perm(),
			stat.Mode().Perm()&gBadJobfilePerms,
		)
		return false, &common.Error{What: msg}
	}

	return true, nil
}

func LoadJobfile(f *os.File, usr *user.User) (*JobFile, error) {
	/*
	   Jobber files have two sections: one begins with "[prefs]" on a
	   line, and the other begins with "[jobs]".  Both contain a YAML
	   document.  The "prefs" section can be parsed with struct
	   UserPrefs, and the "jobs" section is a YAML array of records
	   that can be parsed with struct JobRaw.

	   Legacy format: no section beginnings; whole file is YAML doc
	   for "jobs" section.
	*/

	// parse file into sections
	sections, err := findSections(f)
	if err != nil {
		return nil, err
	}

	var jfile JobFile = JobFile{
		Prefs: UserPrefs{
			Notifier:      MakeMailNotifier(),
			RunLog:        NewMemOnlyRunLog(100),
			StdoutHandler: NopJobOutputHandler{},
			StderrHandler: NopJobOutputHandler{},
		},
	}

	// check for invalid sections
	for sectName, _ := range sections {
		if sectName != PrefsSectName && sectName != JobsSectName {
			return nil, &common.Error{What: fmt.Sprintf("Invalid section: %v", sectName)}
		}
	}

	// parse "prefs" section
	prefsSection, prefsOk := sections[PrefsSectName]
	if prefsOk && len(prefsSection) > 0 {
		ptr, err := parsePrefsSect(prefsSection, usr)
		if err != nil {
			return nil, err
		}
		jfile.Prefs = *ptr
	}

	// parse "jobs" section
	jobsSection, jobsOk := sections[JobsSectName]
	if jobsOk {
		jobs, err := parseJobsSect(jobsSection, usr, &jfile.Prefs)
		if err != nil {
			return nil, err
		}
		jfile.Jobs = jobs
	}

	return &jfile, nil
}

/*
Find the sections of a jobfile.

Returns a map from a section name to the contents of that section.
*/
func findSections(f *os.File) (map[string]string, error) {
	// iterate over lines
	scanner := bufio.NewScanner(f)
	sectionsToLines := make(map[string][]string)
	lineNbr := 0
	sectNameRegexp := regexp.MustCompile("^\\[(\\w*)\\]\\s*$")
	scanner.Split(bufio.ScanLines)

	// to determine legacy vs new format, get first non-empty,
	// non-comment line
	legacyFormat := false
	var currSection *string
	for scanner.Scan() {
		lineNbr++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if len(trimmedLine) == 0 || trimmedLine[0] == '#' {
			// skip empty line or comment
			continue
		} else {
			var matches []string = sectNameRegexp.FindStringSubmatch(line)
			if matches != nil {
				/*
				   New format
				*/
				sectName := matches[1]
				sectionsToLines[sectName] = make([]string, 0)
				currSection = &sectName
			} else {
				/*
				   With legacy format, we treat the whole file as
				   belonging to the "jobs" section.
				*/
				legacyFormat = true
				tmp := JobsSectName
				currSection = &tmp
				sectionsToLines[JobsSectName] = make([]string, 1)
				sectionsToLines[JobsSectName][0] = line
			}
			break
		}
	}

	// handle rest of lines
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
					return nil, &common.Error{What: errMsg}
				}
				sectionsToLines[sectName] = make([]string, 0)
				currSection = &sectName
			} else {
				// save line
				sectionsToLines[*currSection] =
					append(sectionsToLines[*currSection], line)
			}
		}
	}

	// make return value
	retval := make(map[string]string)
	for sectName, lines := range sectionsToLines {
		retval[sectName] = strings.Join(lines, "\n")
	}
	return retval, nil
}

func parsePrefsSect(s string, usr *user.User) (*UserPrefs, error) {
	var rawPrefs UserPrefsRaw
	var userPrefs UserPrefs

	// parse as yaml
	if !strings.HasPrefix(s, gYamlStarter+"\n") {
		s = gYamlStarter + "\n" + s
	}
	err := yaml.UnmarshalStrict([]byte(s), &rawPrefs)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse \"%v\" section",
			PrefsSectName)
		return nil, &common.Error{What: errMsg, Cause: err}
	}

	// parse "logPath"
	if rawPrefs.LogPath != nil {
		/*
		   Relative paths are interpreted as relative to the user's
		   home dir.
		*/
		logPath := *rawPrefs.LogPath
		if filepath.IsAbs(logPath) {
			userPrefs.LogPath = logPath
		} else {
			if len(usr.HomeDir) == 0 {
				errMsg := fmt.Sprintf("User has no home directory, so "+
					"cannot interpret relative log file path %v",
					logPath)
				return nil, &common.Error{What: errMsg}
			}
			userPrefs.LogPath = filepath.Join(usr.HomeDir, logPath)
		}
	} // logPath

	// parse "notifyProgram"
	if rawPrefs.NotifyProgram != nil {
		userPrefs.Notifier = MakeProgramNotifier(*rawPrefs.NotifyProgram)
	} else {
		userPrefs.Notifier = MakeMailNotifier()
	}

	// parse "runLog"
	if rawPrefs.RunLog != nil {
		rawRunLog := rawPrefs.RunLog

		if rawRunLog.Type == "memory" {
			// make memory run log
			maxLen := gDefaultMemRunLogMaxLen
			if rawPrefs.RunLog.MaxLen != nil {
				maxLen = *rawRunLog.MaxLen
			}
			userPrefs.RunLog = NewMemOnlyRunLog(maxLen)

		} else if rawRunLog.Type == "file" {
			const defaultMaxFileLen int64 = 50 * (1 << 20)
			const defaultMaxHistories int = 5

			// check for file path
			if rawRunLog.Path == nil {
				return nil, &common.Error{What: "Missing path for run log"}
			}

			// get max file len
			maxFileLen := defaultMaxFileLen
			if rawRunLog.MaxFileLen != nil {
				maxFileLenStr := *rawRunLog.MaxFileLen

				if len(maxFileLenStr) == 0 {
					msg := fmt.Sprintf("Invalid max file len: '%v'",
						maxFileLenStr)
					return nil, &common.Error{What: msg}
				}

				lastChar := maxFileLenStr[len(maxFileLenStr)-1]
				if lastChar != 'm' && lastChar != 'M' {
					msg := fmt.Sprintf("Invalid max file len: '%v'",
						maxFileLenStr)
					return nil, &common.Error{What: msg}
				}

				numPart := maxFileLenStr[:len(maxFileLenStr)-1]
				tmp, err := strconv.Atoi(numPart)
				if err != nil {
					msg := fmt.Sprintf("Invalid max file len: '%v'",
						maxFileLenStr)
					return nil, &common.Error{What: msg, Cause: err}
				}
				maxFileLen = int64(tmp) * (1 << 20)
			}

			// get max histories
			maxHistories := defaultMaxHistories
			if rawRunLog.MaxHistories != nil {
				maxHistories = *rawRunLog.MaxHistories
			}

			// make file run log
			userPrefs.RunLog, err = NewFileRunLog(
				*rawRunLog.Path,
				maxFileLen,
				maxHistories,
			)
			if err != nil {
				return nil, err
			}

		} else {
			msg := fmt.Sprintf("Invalid run log type: %v", rawRunLog.Type)
			return nil, &common.Error{What: msg}
		}

	} else {
		userPrefs.RunLog = NewMemOnlyRunLog(gDefaultMemRunLogMaxLen)
	} // runLog

	// parse "jobOutput"
	userPrefs.StdoutHandler = NopJobOutputHandler{}
	userPrefs.StderrHandler = NopJobOutputHandler{}
	if rawPrefs.JobOutput != nil {
		bothPrefs := rawPrefs.JobOutput
		if bothPrefs.Stdout != nil {
			outputPrefs := bothPrefs.Stdout
			if err := checkJobOutputPrefs(outputPrefs); err != nil {
				return nil, err
			}
			userPrefs.StdoutHandler = FileJobOutputHandler{
				Where:      *outputPrefs.Where,
				MaxAgeDays: *outputPrefs.MaxAgeDays,
				Suffix:     "stdout",
			}
		}

		if bothPrefs.Stderr != nil {
			outputPrefs := bothPrefs.Stderr
			if err := checkJobOutputPrefs(outputPrefs); err != nil {
				return nil, err
			}
			userPrefs.StderrHandler = FileJobOutputHandler{
				Where:      *outputPrefs.Where,
				MaxAgeDays: *outputPrefs.MaxAgeDays,
				Suffix:     "stderr",
			}
		}
	} // jobOutput

	return &userPrefs, nil
}

func checkJobOutputPrefs(prefs *JobOutputPrefsRaw) error {
	if prefs.Where == nil {
		errMsg := "Job output prefs needs \"where\" field"
		return &common.Error{What: errMsg}
	} else if prefs.MaxAgeDays == nil {
		errMsg := "Job output prefs needs \"maxAgeDays\" field"
		return &common.Error{What: errMsg}
	} else {
		return nil
	}
}

func parseJobsSect(s string, usr *user.User, prefs *UserPrefs) ([]*Job, error) {
	// parse "jobs" section
	var jobConfigs []JobRaw
	if !strings.HasPrefix(s, gYamlStarter+"\n") {
		s = gYamlStarter + "\n" + s
	}
	err := yaml.UnmarshalStrict([]byte(s), &jobConfigs)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse \"%v\" section",
			JobsSectName)
		return nil, &common.Error{What: errMsg, Cause: err}
	}

	// make jobs
	var jobs []*Job
	for _, config := range jobConfigs {
		job := NewJob(config.Name, config.Cmd, usr.Username)
		var err error = nil

		// check name
		if len(config.Name) == 0 {
			return nil, &common.Error{What: "Job name cannot be empty."}
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
		if config.NotifyOnSuccess != nil {
			job.NotifyOnSuccess = *config.NotifyOnSuccess
		}

		// parse time spec
		var tmp *FullTimeSpec
		tmp, err = ParseFullTimeSpec(config.Time)
		if err != nil {
			return nil, err
		}
		job.FullTimeSpec = *tmp
		job.FullTimeSpec.Derandomize()

		// parse job output prefs
		var stdoutOutputPrefs *JobOutputPrefsRaw
		var stderrOutputPrefs *JobOutputPrefsRaw
		if config.JobOutput != nil {
			stdoutOutputPrefs = config.JobOutput.Stdout
			stderrOutputPrefs = config.JobOutput.Stderr
		}
		handler, err := makeOutputHandlerForJob(stdoutOutputPrefs,
			prefs.StdoutHandler, "stdout")
		if err != nil {
			return nil, err
		}
		job.StdoutHandler = handler
		handler, err = makeOutputHandlerForJob(stderrOutputPrefs,
			prefs.StderrHandler, "stderr")
		if err != nil {
			return nil, err
		}
		job.StderrHandler = handler

		jobs = append(jobs, job)
	}

	return jobs, nil
}

func makeOutputHandlerForJob(localPrefs *JobOutputPrefsRaw,
	globalHandler JobOutputHandler, suffix string) (JobOutputHandler, error) {

	if localPrefs == nil {
		return globalHandler, nil
	}

	globalFileHandler, hasGlobalFileHandler := globalHandler.(FileJobOutputHandler)
	if hasGlobalFileHandler {
		jobHandler := FileJobOutputHandler{Suffix: suffix}
		if localPrefs.Where == nil {
			jobHandler.Where = globalFileHandler.Where
		} else {
			jobHandler.Where = *localPrefs.Where
		}
		if localPrefs.MaxAgeDays == nil {
			jobHandler.MaxAgeDays = globalFileHandler.MaxAgeDays
		} else {
			jobHandler.MaxAgeDays = *localPrefs.MaxAgeDays
		}
		return jobHandler, nil

	} else {
		if err := checkJobOutputPrefs(localPrefs); err != nil {
			return nil, err
		}
		return FileJobOutputHandler{
			Where:      *localPrefs.Where,
			MaxAgeDays: *localPrefs.MaxAgeDays,
			Suffix:     suffix,
		}, nil
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
		return nil, &common.Error{What: "Invalid error handler: " + name}
	}
}
