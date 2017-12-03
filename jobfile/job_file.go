package jobfile

import (
	"bufio"
	"fmt"
	"github.com/dshearer/jobber/common"
	"gopkg.in/yaml.v2"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	JobFileName             = ".jobber"
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
}

type JobConfigEntry struct {
	Name            string
	Cmd             string
	Time            string
	OnError         *string "onError,omitempty"
	NotifyOnError   *bool   "notifyOnError,omitempty"
	NotifyOnFailure *bool   "notifyOnFailure,omitempty"
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

func LoadJobFile(path string, username string) (*JobFile, error) {
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
	sections, err := findSections(path)
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
		ptr, err := parsePrefsSect(prefsSection)
		if err != nil {
			return nil, err
		}
		jfile.Prefs = *ptr
	}

	// parse "jobs" section
	jobsSection, jobsOk := sections[JobsSectName]
	if jobsOk {
		jobs, err := parseJobsSect(jobsSection, username)
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
func findSections(path string) (map[string]string, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

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
					return nil, &common.Error{errMsg, nil}
				}
				sectionsToLines[sectName] = make([]string, 0)
				currSection = &sectName

			} else if currSection == nil {
				if len(strings.TrimSpace(line)) > 0 {
					/*
					   To support legacy format, treat whole file as YAML doc
					   for "jobs" section.
					*/
					common.Logger.Println("Using legacy jobber file format.")
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

	// make return value
	retval := make(map[string]string)
	for sectName, lines := range sectionsToLines {
		retval[sectName] = strings.Join(lines, "\n")
	}
	return retval, nil
}

func parsePrefsSect(s string) (*UserPrefs, error) {
	// parse as yaml
	var rawPrefs map[string]interface{}

	if !strings.HasPrefix(s, gYamlStarter+"\n") {
		s = gYamlStarter + "\n" + s
	}
	err := yaml.Unmarshal([]byte(s), &rawPrefs)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse \"%v\" section",
			PrefsSectName)
		return nil, &common.Error{errMsg, err}
	}

	// parse "notifyProgram"
	var userPrefs UserPrefs
	noteProgVal, hasNoteProg := rawPrefs["notifyProgram"]
	if hasNoteProg {
		noteProgValStr, ok := noteProgVal.(string)
		if !ok {
			errMsg := fmt.Sprintf("Invalid value for preference \"notifyProgram\": %v",
				noteProgVal)
			return nil, &common.Error{errMsg, nil}
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
			return nil, &common.Error{errMsg, nil}
		}

		// get type
		typeVal, ok := runLogValMap["type"].(string)
		if !ok {
			errMsg := fmt.Sprintf("Preference \"runLog\" needs \"type\"",
				runLogVal)
			return nil, &common.Error{errMsg, nil}
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
				return nil, &common.Error{msg, nil}
			}

			// get max file len
			maxFileLen := defaultMaxFileLen
			maxFileLenStr, ok := runLogValMap["maxFileLen"].(string)
			if ok {
				if len(maxFileLenStr) == 0 {
					msg := fmt.Sprintf("Invalid max file len: '%v'",
						maxFileLenStr)
					return nil, &common.Error{msg, nil}
				}

				lastChar := maxFileLenStr[len(maxFileLenStr)-1]
				if lastChar != 'm' && lastChar != 'M' {
					msg := fmt.Sprintf("Invalid max file len: '%v'",
						maxFileLenStr)
					return nil, &common.Error{msg, nil}
				}

				numPart := maxFileLenStr[:len(maxFileLenStr)-1]
				tmp, err := strconv.Atoi(numPart)
				if err != nil {
					msg := fmt.Sprintf("Invalid max file len: '%v'",
						maxFileLenStr)
					return nil, &common.Error{msg, err}
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
			errMsg := fmt.Sprintf("Invalid run log type: %v",
				typeVal)
			return nil, &common.Error{errMsg, nil}
		}
	} else {
		userPrefs.RunLog = NewMemOnlyRunLog(gDefaultMemRunLogMaxLen)
	}

	return &userPrefs, nil
}

func parseJobsSect(s string, username string) ([]*Job, error) {
	// parse "jobs" section
	var jobConfigs []JobConfigEntry
	if !strings.HasPrefix(s, gYamlStarter+"\n") {
		s = gYamlStarter + "\n" + s
	}
	err := yaml.Unmarshal([]byte(s), &jobConfigs)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse \"%v\" section",
			JobsSectName)
		return nil, &common.Error{errMsg, err}
	}

	// make jobs
	var jobs []*Job
	for _, config := range jobConfigs {
		job := NewJob(config.Name, config.Cmd, username)
		var err error = nil

		// check name
		if len(config.Name) == 0 {
			return nil, &common.Error{"Job name cannot be empty.", nil}
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
		return nil, &common.Error{"Invalid error handler: " + name, nil}
	}
}
