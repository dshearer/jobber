package jobfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func NewString(val string) *string {
	return &val
}

var gUserEx = user.User{Username: "bob", HomeDir: "/home/bob"}

var gEverySecTimeSpec = FullTimeSpec{
	Sec:  WildcardTimeSpec{},
	Min:  WildcardTimeSpec{},
	Hour: WildcardTimeSpec{},
	Mday: WildcardTimeSpec{},
	Mon:  WildcardTimeSpec{},
	Wday: WildcardTimeSpec{},
}

const gV1JobFileContents = `---
- name: DailyBackup
  cmd: backup daily
  time: 0 0 14
  onError: Stop
  notifyOnError: false
  notifyOnFailure: true

- name: WeeklyBackup
  cmd: |
    multi-
    line
    script
  time: 0 0 14 * * 1
  onError: Backoff
  notifyOnError: true
  notifyOnFailure: false

- name: JobA
  cmd: whatever
  time: "* * * * * *"
  onError: Backoff
  notifyOnError: true
  notifyOnFailure: false

- name: JobB
  cmd: whatever
  time: '*'
  onError: Backoff
  notifyOnError: true
  notifyOnFailure: false`

var gV1JobFile = JobFile{
	Prefs: UserPrefs{
		RunLog: NewMemOnlyRunLog(100),
	},
	Jobs: map[string]*Job{
		"DailyBackup": &Job{
			Name: "DailyBackup",
			Cmd:  "backup daily",
			User: gUserEx.Username,
			FullTimeSpec: FullTimeSpec{
				Sec:  OneValTimeSpec{0},
				Min:  OneValTimeSpec{0},
				Hour: OneValTimeSpec{14},
				Mday: WildcardTimeSpec{},
				Mon:  WildcardTimeSpec{},
				Wday: WildcardTimeSpec{},
			},
			ErrorHandler:    StopErrorHandler{},
			NotifyOnError:   nil,
			NotifyOnFailure: []ResultSink{SystemEmailResultSink{}},
			NotifyOnSuccess: nil,
		},
		"WeeklyBackup": &Job{
			Name: "WeeklyBackup",
			Cmd: `multi-
line
script
`,
			User: gUserEx.Username,
			FullTimeSpec: FullTimeSpec{
				Sec:  OneValTimeSpec{0},
				Min:  OneValTimeSpec{0},
				Hour: OneValTimeSpec{14},
				Mday: WildcardTimeSpec{},
				Mon:  WildcardTimeSpec{},
				Wday: OneValTimeSpec{1},
			},
			ErrorHandler:    BackoffErrorHandler{},
			NotifyOnError:   []ResultSink{SystemEmailResultSink{}},
			NotifyOnFailure: nil,
			NotifyOnSuccess: nil,
		},
		"JobA": &Job{
			Name:            "JobA",
			Cmd:             "whatever",
			User:            gUserEx.Username,
			FullTimeSpec:    gEverySecTimeSpec,
			ErrorHandler:    BackoffErrorHandler{},
			NotifyOnError:   []ResultSink{SystemEmailResultSink{}},
			NotifyOnFailure: nil,
			NotifyOnSuccess: nil,
		},
		"JobB": &Job{
			Name:            "JobB",
			Cmd:             "whatever",
			User:            gUserEx.Username,
			FullTimeSpec:    gEverySecTimeSpec,
			ErrorHandler:    BackoffErrorHandler{},
			NotifyOnError:   []ResultSink{SystemEmailResultSink{}},
			NotifyOnFailure: nil,
			NotifyOnSuccess: nil,
		},
	},
}

const gV2JobFileContents = `
# Must be able
# to deal with comments.
[prefs]
# Which could be (almost) anywhere.
notifyProgram: ~/handleError

# Even here!

[jobs]
- name: DailyBackup
  cmd: backup daily
# And here
  time: 0 0 14
  onError: Stop
  notifyOnError: false
  notifyOnFailure: true

- name: WeeklyBackup
  cmd: | # And even here
    multi-
    line
    script
  time: 0 0 14 * * 1
  onError: Backoff  # Here
  notifyOnError: true
  notifyOnFailure: false

- name: SuccessReport
  cmd: exit 0
  time: 0 0 14 * * 1
  onError: Backoff
  notifyOnError: false
  notifyOnFailure: false
  notifyOnSuccess: true

# So many comments...
`

var gV2JobFile = JobFile{
	Prefs: UserPrefs{
		RunLog: NewMemOnlyRunLog(100),
	},
	Jobs: map[string]*Job{
		"DailyBackup": &Job{
			Name: "DailyBackup",
			Cmd:  "backup daily",
			User: gUserEx.Username,
			FullTimeSpec: FullTimeSpec{
				Sec:  OneValTimeSpec{0},
				Min:  OneValTimeSpec{0},
				Hour: OneValTimeSpec{14},
				Mday: WildcardTimeSpec{},
				Mon:  WildcardTimeSpec{},
				Wday: WildcardTimeSpec{},
			},
			ErrorHandler:    StopErrorHandler{},
			NotifyOnError:   nil,
			NotifyOnFailure: []ResultSink{ProgramResultSink{Path: "~/handleError"}},
			NotifyOnSuccess: nil,
		},
		"WeeklyBackup": &Job{
			Name: "WeeklyBackup",
			Cmd: `multi-
line
script
`,
			User: gUserEx.Username,
			FullTimeSpec: FullTimeSpec{
				Sec:  OneValTimeSpec{0},
				Min:  OneValTimeSpec{0},
				Hour: OneValTimeSpec{14},
				Mday: WildcardTimeSpec{},
				Mon:  WildcardTimeSpec{},
				Wday: OneValTimeSpec{1},
			},
			ErrorHandler:    BackoffErrorHandler{},
			NotifyOnError:   []ResultSink{ProgramResultSink{Path: "~/handleError"}},
			NotifyOnFailure: nil,
			NotifyOnSuccess: nil,
		},
		"SuccessReport": &Job{
			Name: "SuccessReport",
			Cmd:  "exit 0",
			User: gUserEx.Username,
			FullTimeSpec: FullTimeSpec{
				Sec:  OneValTimeSpec{0},
				Min:  OneValTimeSpec{0},
				Hour: OneValTimeSpec{14},
				Mday: WildcardTimeSpec{},
				Mon:  WildcardTimeSpec{},
				Wday: OneValTimeSpec{1},
			},
			ErrorHandler:    BackoffErrorHandler{},
			NotifyOnError:   nil,
			NotifyOnFailure: nil,
			NotifyOnSuccess: []ResultSink{ProgramResultSink{Path: "~/handleError"}},
		},
	},
}

type JobFileTestCase struct {
	Input  string
	Output JobFile
	Error  bool
}

var gFileRunLog, _ = NewFileRunLog(
	"/tmp/claudius",
	int64(10*(1<<20)),
	20,
)

var gJobFileV1V2TestCases = []JobFileTestCase{
	{
		Input:  gV1JobFileContents,
		Output: gV1JobFile,
	},
	{
		Input:  gV2JobFileContents,
		Output: gV2JobFile,
	},
	{
		Input: `
[jobs]
- name: DailyBackup
  cmd: backup daily
  time: 0 0 14
  onError: Stop
  notifyOnError: false
  notifyOnFailure: true
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog: NewMemOnlyRunLog(100),
			},
			Jobs: map[string]*Job{
				"DailyBackup": &Job{
					Name: "DailyBackup",
					Cmd:  "backup daily",
					User: gUserEx.Username,
					FullTimeSpec: FullTimeSpec{
						Sec:  OneValTimeSpec{0},
						Min:  OneValTimeSpec{0},
						Hour: OneValTimeSpec{14},
						Mday: WildcardTimeSpec{},
						Mon:  WildcardTimeSpec{},
						Wday: WildcardTimeSpec{},
					},
					ErrorHandler:    StopErrorHandler{},
					NotifyOnError:   nil,
					NotifyOnFailure: []ResultSink{SystemEmailResultSink{}},
					NotifyOnSuccess: nil,
				},
			},
		},
	},
	{
		Input: `
[prefs]
runLog:
    type: memory
    maxLen: 10
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog: NewMemOnlyRunLog(10),
			},
			Jobs: nil,
		},
	},
	{
		Input: `
[prefs]
runLog:
    type: memory
    maxLen: 10
[jobs]
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog: NewMemOnlyRunLog(10),
			},
			Jobs: nil,
		},
	},
	{
		Input: `
[prefs]
runLog:
    type: file
    path: /tmp/claudius
    maxFileLen: 10m
    maxHistories: 20
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog: gFileRunLog,
			},
			Jobs: nil,
		},
	},
	{
		Input: `
[prefs]
runLog:
    type: file
    path: /dir/does/not/exist/claudius
    maxFileLen: 10m
    maxHistories: 20
`,
		Error: true,
	},
	{
		Input: `[prefs]
logPath: /my/log/path
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog:  NewMemOnlyRunLog(100),
				LogPath: "/my/log/path",
			},
		},
	},
	{
		Input: `[prefs]
logPath: my/log/path
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog:  NewMemOnlyRunLog(100),
				LogPath: filepath.Join(gUserEx.HomeDir, "my/log/path"),
			},
		},
	},
	{
		Input: `[prefs]
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog:  NewMemOnlyRunLog(100),
				LogPath: "",
			},
		},
	},
	{
		Input: `[badSection]
`,
		Error: true,
	},
	{
		Input: `[unparseable
`,
		Error: true,
	},
}

func TestLoadJobFileV1V2(t *testing.T) {
	for _, testCase := range gJobFileV1V2TestCases {
		/*
		 * Set up
		 */
		fmt.Printf("Input:\n%v\n", testCase.Input)

		// make jobfile
		f, err := ioutil.TempFile("", "Testing")
		if err != nil {
			panic(fmt.Sprintf("Failed to make tempfile: %v", err))
		}
		defer f.Close()
		defer os.Remove(f.Name())
		f.WriteString(testCase.Input)
		f.Seek(0, 0)

		/*
		 * Call
		 */
		var file *JobFile
		file, err = LoadJobfile(f, &gUserEx)

		/*
		 * Test
		 */

		if testCase.Error {
			/* We expect error */
			require.NotNil(t, err, "Expected error, but didn't get one")
			continue
		}

		require.Nil(t, err, "%v", err)
		require.NotNil(t, file)
		require.Equal(t, testCase.Output.Prefs, file.Prefs)
		require.Equal(t, len(testCase.Output.Jobs), len(file.Jobs))
		for jobName, _ := range testCase.Output.Jobs {
			require.Equal(t, testCase.Output.Jobs[jobName].FullTimeSpec, file.Jobs[jobName].FullTimeSpec)
			require.Equal(t, testCase.Output.Jobs[jobName].NotifyOnError, file.Jobs[jobName].NotifyOnError)
			require.Equal(t, testCase.Output.Jobs[jobName].NotifyOnFailure, file.Jobs[jobName].NotifyOnFailure)
			require.Equal(t, testCase.Output.Jobs[jobName].NotifyOnSuccess, file.Jobs[jobName].NotifyOnSuccess)
			require.Equal(t, *testCase.Output.Jobs[jobName], *file.Jobs[jobName])
		}
	}
}
