package jobfile

import (
	"bufio"
	"fmt"
	"io/ioutil"
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
	Jobs  map[string]*Job
}

type UserPrefs struct {
	RunLog  RunLog
	LogPath string // for error msgs etc.  May be "".
}

func (self *UserPrefs) String() string {
	s := ""
	s += fmt.Sprintf("RunLog: %v\n", self.RunLog)
	if len(self.LogPath) > 0 {
		s += fmt.Sprintf("Log path: %v\n", self.LogPath)
	}
	return s
}

type JobFileV3Raw struct {
	Version     string              `yaml:"version"`
	Prefs       UserPrefsV3Raw      `yaml:"prefs"`
	ResultSinks []ResultSinkRaw     `yaml:"resultSinks"`
	Jobs        map[string]JobV3Raw `yaml:"jobs"`
}

type JobFileV1V2Raw struct {
	Prefs UserPrefsV1V2Raw
	Jobs  []JobV1V2Raw
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

type UserPrefsV3Raw struct {
	LogPath *string    `yaml:"logPath"`
	RunLog  *RunLogRaw `yaml:"runLog"`
}

type UserPrefsV1V2Raw struct {
	LogPath       *string    `yaml:"logPath"`
	RunLog        *RunLogRaw `yaml:"runLog"`
	NotifyProgram *string    `yaml:"notifyProgram"`
}

type JobV3Raw struct {
	Cmd             string          `json:"cmd" yaml:"cmd"`
	Time            string          `json:"time" yaml:"time"`
	OnError         *string         `json:"onError" yaml:"onError"`
	NotifyOnSuccess []ResultSinkRaw `json:"notifyOnSuccess" yaml:"notifyOnSuccess"`
	NotifyOnError   []ResultSinkRaw `json:"notifyOnError" yaml:"notifyOnError"`
	NotifyOnFailure []ResultSinkRaw `json:"notifyOnFailure" yaml:"notifyOnFailure"`
}

type JobV1V2Raw struct {
	Name            string  `json:"name" yaml:"name"`
	Cmd             string  `json:"cmd" yaml:"cmd"`
	Time            string  `json:"time" yaml:"time"`
	OnError         *string `json:"onError" yaml:"onError"`
	NotifyOnSuccess *bool   `json:"notifyOnSuccess" yaml:"notifyOnSuccess"`
	NotifyOnError   *bool   `json:"notifyOnError" yaml:"notifyOnError"`
	NotifyOnFailure *bool   `json:"notifyOnFailure" yaml:"notifyOnFailure"`
}

func NewEmptyJobFile() JobFile {
	return JobFile{
		Prefs: UserPrefs{
			RunLog: NewMemOnlyRunLog(gDefaultMemRunLogMaxLen),
		},
		Jobs: make(map[string]*Job),
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
	/* V3 jobfiles are pure YAML documents. */

	// parse it
	version, err := jobfileVersion(f)
	if err != nil {
		return nil, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}
	var parseFunc func(*os.File) (*JobFileV3Raw, error)
	if c := version.Compare(SemVer{Major: 1, Minor: 4}); c >= 0 {
		parseFunc = parseV3Jobfile
	} else {
		parseFunc = parseV1V2Jobfile
	}
	jobfileRaw, err := parseFunc(f)
	if err != nil {
		return nil, err
	}

	jfile := NewEmptyJobFile()

	// parse prefs
	if err := jobfileRaw.Prefs.ToPrefs(usr, &jfile.Prefs); err != nil {
		return nil, err
	}

	// parse jobs
	for jobName, jobRaw := range jobfileRaw.Jobs {
		job := NewJob()
		job.Name = jobName
		if err := jobRaw.ToJob(usr, &job); err != nil {
			return nil, err
		}
		jfile.Jobs[jobName] = &job
	}

	return &jfile, nil
}

func jobfileVersion(f *os.File) (*SemVer, error) {
	// read file
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// check version
	v1Content := make([]interface{}, 0)
	if err := yaml.Unmarshal(data, &v1Content); err == nil {
		return &SemVer{Major: 1}, nil

	} else if strings.HasPrefix(string(data), "[jobs]") ||
		strings.Contains(string(data), "\n[jobs]\n") ||
		strings.HasPrefix(string(data), "[prefs]") ||
		strings.Contains(string(data), "\n[prefs]\n") {
		return &SemVer{Major: 1, Minor: 2}, nil

	} else {
		var tmp struct {
			Version string `yaml:"version"`
		}
		if err := yaml.Unmarshal(data, &tmp); err != nil {
			return nil, err
		}
		if len(tmp.Version) == 0 {
			return nil, &common.Error{What: "Missing jobfile version"}
		}
		return ParseSemVer(tmp.Version)
	}
}

func parseV3Jobfile(f *os.File) (*JobFileV3Raw, error) {
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	dataStr := string(data)
	if !strings.HasPrefix(dataStr, gYamlStarter+"\n") {
		dataStr = gYamlStarter + "\n" + dataStr
	}
	var jobfileRaw JobFileV3Raw
	if err := yaml.UnmarshalStrict([]byte(dataStr), &jobfileRaw); err != nil {
		return nil, err
	}
	return &jobfileRaw, nil
}

func v1v2ToV3(v1v2Jobfile JobFileV1V2Raw) (*JobFileV3Raw, error) {
	var v3Jobfile JobFileV3Raw
	v3Jobfile.Jobs = make(map[string]JobV3Raw)

	// make prefs
	v3Jobfile.Prefs.LogPath = v1v2Jobfile.Prefs.LogPath
	v3Jobfile.Prefs.RunLog = v1v2Jobfile.Prefs.RunLog

	// make result sink
	resultSink := make(map[string]interface{})
	if v1v2Jobfile.Prefs.NotifyProgram != nil {
		resultSink["type"] = "program"
		resultSink["path"] = *v1v2Jobfile.Prefs.NotifyProgram
	} else {
		resultSink["type"] = "system-email"
	}
	resultSinkArray := []ResultSinkRaw{resultSink}

	// make jobs
	for _, v1v2JobRaw := range v1v2Jobfile.Jobs {
		var v3JobRaw JobV3Raw
		v3JobRaw.Cmd = v1v2JobRaw.Cmd
		v3JobRaw.Time = v1v2JobRaw.Time
		v3JobRaw.OnError = v1v2JobRaw.OnError

		notifyOnError := false
		if v1v2JobRaw.NotifyOnError != nil {
			notifyOnError = *v1v2JobRaw.NotifyOnError
		}
		notifyOnFailure := true
		if v1v2JobRaw.NotifyOnFailure != nil {
			notifyOnFailure = *v1v2JobRaw.NotifyOnFailure
		}
		notifyOnSuccess := false
		if v1v2JobRaw.NotifyOnSuccess != nil {
			notifyOnSuccess = *v1v2JobRaw.NotifyOnSuccess
		}

		if notifyOnError {
			v3JobRaw.NotifyOnError = resultSinkArray
		}
		if notifyOnFailure {
			v3JobRaw.NotifyOnFailure = resultSinkArray
		}
		if notifyOnSuccess {
			v3JobRaw.NotifyOnSuccess = resultSinkArray
		}

		_, ok := v3Jobfile.Jobs[v1v2JobRaw.Name]
		if ok {
			msg := fmt.Sprintf("Multiple jobs named \"%v\"", v1v2JobRaw.Name)
			return nil, &common.Error{What: msg}
		}
		v3Jobfile.Jobs[v1v2JobRaw.Name] = v3JobRaw
	}

	return &v3Jobfile, nil
}

func parseV1V2Jobfile(f *os.File) (*JobFileV3Raw, error) {
	/*
	   V2 jobfiles have two sections: one begins with "[prefs]" on a
	   line, and the other begins with "[jobs]".  Both contain a YAML
	   document.  The "prefs" section can be parsed with struct
	   UserPrefs, and the "jobs" section is a YAML array of records
	   that can be parsed with struct JobV1V2Raw.

	   V1 format: no section beginnings; whole file is YAML doc
	   for "jobs" section.
	*/

	// parse file into sections
	sections, err := findSections(f)
	if err != nil {
		return nil, err
	}

	var jfile JobFileV1V2Raw

	// check for invalid sections
	for sectName, _ := range sections {
		if sectName != PrefsSectName && sectName != JobsSectName {
			return nil, &common.Error{What: fmt.Sprintf("Invalid section: %v", sectName)}
		}
	}

	// parse "prefs" section
	prefsSection, prefsOk := sections[PrefsSectName]
	if prefsOk && len(prefsSection) > 0 {
		// parse as yaml
		if !strings.HasPrefix(prefsSection, gYamlStarter+"\n") {
			prefsSection = gYamlStarter + "\n" + prefsSection
		}
		err := yaml.UnmarshalStrict([]byte(prefsSection), &jfile.Prefs)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to parse \"%v\" section",
				PrefsSectName)
			return nil, &common.Error{What: errMsg, Cause: err}
		}
	}

	// parse "jobs" section
	jobsSection, jobsOk := sections[JobsSectName]
	if jobsOk {
		if !strings.HasPrefix(jobsSection, gYamlStarter+"\n") {
			jobsSection = gYamlStarter + "\n" + jobsSection
		}
		err := yaml.UnmarshalStrict([]byte(jobsSection), &jfile.Jobs)
		if err != nil {
			errMsg := fmt.Sprintf("Failed to parse \"%v\" section",
				JobsSectName)
			return nil, &common.Error{What: errMsg, Cause: err}
		}
	}

	// convert to 1.3 format
	return v1v2ToV3(jfile)
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

func (self RunLogRaw) ToRunLog() (RunLog, error) {
	if self.Type == "memory" {
		// make memory run log
		maxLen := gDefaultMemRunLogMaxLen
		if self.MaxLen != nil {
			maxLen = *self.MaxLen
		}
		return NewMemOnlyRunLog(maxLen), nil

	} else if self.Type == "file" {
		const defaultMaxFileLen int64 = 50 * (1 << 20)
		const defaultMaxHistories int = 5

		// check for file path
		if self.Path == nil {
			return nil, &common.Error{What: "Missing path for run log"}
		}

		// get max file len
		maxFileLen := defaultMaxFileLen
		if self.MaxFileLen != nil {
			maxFileLenStr := *self.MaxFileLen

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
		if self.MaxHistories != nil {
			maxHistories = *self.MaxHistories
		}

		// make file run log
		return NewFileRunLog(*self.Path, maxFileLen, maxHistories)

	} else {
		msg := fmt.Sprintf("Invalid run log type: %v", self.Type)
		return nil, &common.Error{What: msg}
	}
}

func (self UserPrefsV3Raw) ToPrefs(usr *user.User, dest *UserPrefs) error {
	// parse "logPath"
	if self.LogPath != nil {
		/*
		   Relative paths are interpreted as relative to the user's
		   home dir.
		*/
		logPath := *self.LogPath
		if filepath.IsAbs(logPath) {
			dest.LogPath = logPath
		} else {
			if len(usr.HomeDir) == 0 {
				errMsg := fmt.Sprintf("User has no home directory, so "+
					"cannot interpret relative log file path %v",
					logPath)
				return &common.Error{What: errMsg}
			}
			dest.LogPath = filepath.Join(usr.HomeDir, logPath)
		}
	} // logPath

	// parse "runLog"
	if self.RunLog != nil {
		runLog, err := self.RunLog.ToRunLog()
		if err != nil {
			return err
		}
		dest.RunLog = runLog
	} else {
		dest.RunLog = NewMemOnlyRunLog(gDefaultMemRunLogMaxLen)
	}

	return nil
}

func normalizeResultSinkArray(sinks []ResultSink) []ResultSink {
	// remove duplicates
	var newSinks []ResultSink
	alreadyHave := func(sink ResultSink) bool {
		for _, currSink := range newSinks {
			if currSink.Equals(sink) {
				return true
			}
		}
		return false
	}
	for _, currSink := range sinks {
		if !alreadyHave(currSink) {
			newSinks = append(newSinks, currSink)
		}
	}
	return newSinks
}

func (self JobV3Raw) ToJob(usr *user.User, dest *Job) error {
	// set cmd, user
	dest.Cmd = self.Cmd
	dest.User = usr.Username

	// set failure-handler
	if self.OnError != nil {
		var err error
		dest.ErrorHandler, err = GetErrorHandler(*self.OnError)
		if err != nil {
			return err
		}
	}

	// handle NotifyOnError
	for _, sinkRaw := range self.NotifyOnError {
		sink, err := MakeResultSinkFromConfig(sinkRaw)
		if err != nil {
			return err
		}
		dest.NotifyOnError = append(dest.NotifyOnError, sink)
	}
	dest.NotifyOnError = normalizeResultSinkArray(dest.NotifyOnError)

	// handle NotifyOnFailure
	for _, sinkRaw := range self.NotifyOnFailure {
		sink, err := MakeResultSinkFromConfig(sinkRaw)
		if err != nil {
			return err
		}
		dest.NotifyOnFailure = append(dest.NotifyOnFailure, sink)
	}
	dest.NotifyOnFailure = normalizeResultSinkArray(dest.NotifyOnFailure)

	// handle NotifyOnSuccess
	for _, sinkRaw := range self.NotifyOnSuccess {
		sink, err := MakeResultSinkFromConfig(sinkRaw)
		if err != nil {
			return err
		}
		dest.NotifyOnSuccess = append(dest.NotifyOnSuccess, sink)
	}
	dest.NotifyOnSuccess = normalizeResultSinkArray(dest.NotifyOnSuccess)

	// parse time spec
	tmp, err := ParseFullTimeSpec(self.Time)
	if err != nil {
		return err
	}
	dest.FullTimeSpec = *tmp
	dest.FullTimeSpec.Derandomize()

	return nil
}
