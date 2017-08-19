package jobfile

import (
	"bufio"
	"fmt"
	"github.com/dshearer/jobber/Godeps/_workspace/src/gopkg.in/yaml.v2"
	"github.com/dshearer/jobber/common"
	"os"
	"regexp"
	"strings"
)

const (
	JobFileName   = ".jobber"
	PrefsSectName = "prefs"
	JobsSectName  = "jobs"
)

type JobFile struct {
	Prefs UserPrefs
	Jobs  []*Job
}

type UserPrefs struct {
	Notifier RunRecNotifier
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
	return &JobFile{Jobs: make([]*Job, 0)}
}

func LoadJobFile(path string, username string) (*JobFile, error) {
	/*
	   Jobber files have two sections: one begins with "[prefs]" on a line, and
	   the other begins with "[jobs]".  Both contain a YAML document.  The "prefs"
	   section can be parsed with struct UserPrefs, and the "jobs" section is a
	   YAML array of records that can be parsed with struct JobConfigEntry.

	   Legacy format: no section beginnings; whole file is YAML doc for "jobs"
	   section.
	*/

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

	// parse "prefs" section
	const yamlStarter string = "---"
	rawPrefs := map[string]interface{}{}
	prefsLines, prefsOk := sectionsToLines[PrefsSectName]
	if prefsOk && len(prefsLines) > 0 {
		common.Logger.Println("Got prefs section")
		prefsSection := strings.Join(prefsLines, "\n")
		if strings.TrimRight(prefsLines[0], " \t") != yamlStarter {
			prefsSection = yamlStarter + "\n" + prefsSection
		}
		err := yaml.Unmarshal([]byte(prefsSection), &rawPrefs)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to parse \"%v\" section",
				PrefsSectName)
			return nil, &common.Error{errMsg, err}
		}
	}

	// parse "jobs" section
	var jobConfigs []JobConfigEntry
	jobsLines, jobsOk := sectionsToLines[JobsSectName]
	if jobsOk && len(jobsLines) > 0 {
		common.Logger.Println("Got jobs section")
		jobsSection := strings.Join(jobsLines, "\n")
		if strings.TrimRight(jobsLines[0], " \t") != yamlStarter {
			jobsSection = yamlStarter + "\n" + jobsSection
		}
		err := yaml.Unmarshal([]byte(jobsSection), &jobConfigs)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to parse \"%v\" section",
				JobsSectName)
			return nil, &common.Error{errMsg, err}
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
			return nil, &common.Error{errMsg, nil}
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

	return &JobFile{userPrefs, jobs}, nil
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
