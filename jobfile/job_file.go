package jobfile

import (
	"bufio"
	"fmt"
	"github.com/dshearer/jobber/common"
	"gopkg.in/yaml.v2"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
}

type JobConfigEntry struct {
	Name            string
	Cmd             string
	Time            string
	OnError         *string `yaml:"onError,omitempty"`
	NotifyOnSuccess *bool   `yaml:"notifyOnSuccess,omitempty"`
	NotifyOnError   *bool   `yaml:"notifyOnError,omitempty"`
	NotifyOnFailure *bool   `yaml:"notifyOnFailure,omitempty"`
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
	if stat.Mode().Perm()&0022 > 0 {
		msg := fmt.Sprintf("Jobfile has bad permissions: %v",
			stat.Mode().Perm())
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
	   that can be parsed with struct JobConfigEntry.

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
			Notifier: MakeMailNotifier(),
			RunLog:   NewMemOnlyRunLog(100),
		},
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
		jobs, err := parseJobsSect(jobsSection, usr)
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
	// parse as yaml
	var rawPrefs map[string]interface{}

	if !strings.HasPrefix(s, gYamlStarter+"\n") {
		s = gYamlStarter + "\n" + s
	}
	err := yaml.Unmarshal([]byte(s), &rawPrefs)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse \"%v\" section",
			PrefsSectName)
		return nil, &common.Error{What: errMsg, Cause: err}
	}

	// parse "notifyProgram"
	var userPrefs UserPrefs
	noteProgVal, hasNoteProg := rawPrefs["notifyProgram"]
	if hasNoteProg {
		noteProgValStr, ok := noteProgVal.(string)
		if !ok {
			errMsg := fmt.Sprintf("Invalid value for preference \"notifyProgram\": %v",
				noteProgVal)
			return nil, &common.Error{What: errMsg}
		}
		userPrefs.Notifier = MakeProgramNotifier(noteProgValStr)
	} else {
		userPrefs.Notifier = MakeMailNotifier()
	}

	// parse "runLog"
	runLogVal, hasRunLog := rawPrefs["runLog"]
	if hasRunLog {
		runLogValMap, ok := runLogVal.(map[interface{}]interface{})
		if !ok {
			errMsg := fmt.Sprintf("Invalid value for preference \"runLog\": %v",
				runLogVal)
			return nil, &common.Error{What: errMsg}
		}

		// get type
		typeVal, ok := runLogValMap["type"].(string)
		if !ok {
			errMsg := fmt.Sprintf("Preference \"runLog\" needs \"type\"")
			return nil, &common.Error{What: errMsg}
		}

		if typeVal == "memory" {
			// make memory run log
			maxLen := gDefaultMemRunLogMaxLen
			tmp, ok := runLogValMap["maxLen"].(int)
			if ok {
				maxLen = tmp
			}
			userPrefs.RunLog = NewMemOnlyRunLog(maxLen)
		} else if typeVal == "file" {
			const defaultMaxFileLen int64 = 50 * (1 << 20)
			const defaultMaxHistories int = 5

			// get file path
			filePath, ok := runLogValMap["path"].(string)
			if !ok {
				msg := fmt.Sprintf("Missing run log path")
				return nil, &common.Error{What: msg}
			}

			// get max file len
			maxFileLen := defaultMaxFileLen
			maxFileLenStr, ok := runLogValMap["maxFileLen"].(string)
			if ok {
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
			tmp, ok := runLogValMap["maxHistories"].(int)
			if ok {
				maxHistories = tmp
			}

			// make file run log
			userPrefs.RunLog, err = NewFileRunLog(
				filePath,
				maxFileLen,
				maxHistories,
			)
			if err != nil {
				return nil, err
			}
		} else {
			msg := fmt.Sprintf("Invalid run log type: %v",
				typeVal)
			return nil, &common.Error{What: msg}
		}
	} else {
		userPrefs.RunLog = NewMemOnlyRunLog(gDefaultMemRunLogMaxLen)
	}

	// parse LogPath
	logPathVal, hasLogPath := rawPrefs["logPath"]
	if hasLogPath && logPathVal != nil {
		// ensure it's a string
		logPath, ok := logPathVal.(string)
		if !ok {
			errMsg := fmt.Sprintf("Invalid value for preference "+
				"\"logPath\": %v", logPathVal)
			return nil, &common.Error{What: errMsg}
		}

		/*
		   Relative paths are interpreted as relative to the user's
		   home dir.
		*/
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
	}

	return &userPrefs, nil
}

func parseJobsSect(s string, usr *user.User) ([]*Job, error) {
	// parse "jobs" section
	var jobConfigs []JobConfigEntry
	if !strings.HasPrefix(s, gYamlStarter+"\n") {
		s = gYamlStarter + "\n" + s
	}
	err := yaml.Unmarshal([]byte(s), &jobConfigs)
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

		jobs = append(jobs, job)
	}

	return jobs, nil
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
