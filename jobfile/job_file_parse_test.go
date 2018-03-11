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
		RunLog:        NewMemOnlyRunLog(100),
		StdoutHandler: NopJobOutputHandler{},
		StderrHandler: NopJobOutputHandler{},
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
			NotifyOnError:   NopRunRecNotifier{},
			NotifyOnFailure: MailRunRecNotifier{},
			NotifyOnSuccess: NopRunRecNotifier{},
			StdoutHandler:   NopJobOutputHandler{},
			StderrHandler:   NopJobOutputHandler{},
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
			NotifyOnError:   MailRunRecNotifier{},
			NotifyOnFailure: NopRunRecNotifier{},
			NotifyOnSuccess: NopRunRecNotifier{},
			StdoutHandler:   NopJobOutputHandler{},
			StderrHandler:   NopJobOutputHandler{},
		},
		"JobA": &Job{
			Name:            "JobA",
			Cmd:             "whatever",
			User:            gUserEx.Username,
			FullTimeSpec:    gEverySecTimeSpec,
			ErrorHandler:    BackoffErrorHandler{},
			NotifyOnError:   MailRunRecNotifier{},
			NotifyOnFailure: NopRunRecNotifier{},
			NotifyOnSuccess: NopRunRecNotifier{},
			StdoutHandler:   NopJobOutputHandler{},
			StderrHandler:   NopJobOutputHandler{},
		},
		"JobB": &Job{
			Name:            "JobB",
			Cmd:             "whatever",
			User:            gUserEx.Username,
			FullTimeSpec:    gEverySecTimeSpec,
			ErrorHandler:    BackoffErrorHandler{},
			NotifyOnError:   MailRunRecNotifier{},
			NotifyOnFailure: NopRunRecNotifier{},
			NotifyOnSuccess: NopRunRecNotifier{},
			StdoutHandler:   NopJobOutputHandler{},
			StderrHandler:   NopJobOutputHandler{},
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
		NotifyProgram: NewString("~/handleError"),
		RunLog:        NewMemOnlyRunLog(100),
		StdoutHandler: NopJobOutputHandler{},
		StderrHandler: NopJobOutputHandler{},
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
			NotifyOnError:   NopRunRecNotifier{},
			NotifyOnFailure: ProgramRunRecNotifier{Program: "~/handleError"},
			NotifyOnSuccess: NopRunRecNotifier{},
			StdoutHandler:   NopJobOutputHandler{},
			StderrHandler:   NopJobOutputHandler{},
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
			NotifyOnError:   ProgramRunRecNotifier{Program: "~/handleError"},
			NotifyOnFailure: NopRunRecNotifier{},
			NotifyOnSuccess: NopRunRecNotifier{},
			StdoutHandler:   NopJobOutputHandler{},
			StderrHandler:   NopJobOutputHandler{},
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
			NotifyOnError:   NopRunRecNotifier{},
			NotifyOnFailure: NopRunRecNotifier{},
			NotifyOnSuccess: ProgramRunRecNotifier{Program: "~/handleError"},
			StdoutHandler:   NopJobOutputHandler{},
			StderrHandler:   NopJobOutputHandler{},
		},
	},
}

const gV3JobFileContents = `
# Must be able
# to deal with comments.
prefs:
  # Which could be (almost) anywhere.
  notifyProgram: ~/handleError

# Even here!

jobs:
  DailyBackup:
    cmd: backup daily
  # And here
    time: 0 0 14
    onError: Stop
    notifyOnError: false
    notifyOnFailure: true

  WeeklyBackup:
    cmd: | # And even here
      multi-
      line
      script
    time: 0 0 14 * * 1
    onError: Backoff  # Here
    notifyOnError: true
    notifyOnFailure: false

  SuccessReport:
    cmd: exit 0
    time: 0 0 14 * * 1
    onError: Backoff
    notifyOnError: false
    notifyOnFailure: false
    notifyOnSuccess: true

# So many comments...
`

var gV3JobFile = gV2JobFile

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

var gTestCases = []JobFileTestCase{
	{
		Input:  gV1JobFileContents,
		Output: gV1JobFile,
	},
	{
		Input:  gV2JobFileContents,
		Output: gV2JobFile,
	},
	{
		Input:  gV3JobFileContents,
		Output: gV3JobFile,
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
				RunLog:        NewMemOnlyRunLog(100),
				StdoutHandler: NopJobOutputHandler{},
				StderrHandler: NopJobOutputHandler{},
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
					NotifyOnError:   NopRunRecNotifier{},
					NotifyOnFailure: MailRunRecNotifier{},
					NotifyOnSuccess: NopRunRecNotifier{},
					StdoutHandler:   NopJobOutputHandler{},
					StderrHandler:   NopJobOutputHandler{},
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
				RunLog:        NewMemOnlyRunLog(10),
				StdoutHandler: NopJobOutputHandler{},
				StderrHandler: NopJobOutputHandler{},
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
				RunLog:        gFileRunLog,
				StdoutHandler: NopJobOutputHandler{},
				StderrHandler: NopJobOutputHandler{},
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
				RunLog:        NewMemOnlyRunLog(100),
				LogPath:       "/my/log/path",
				StdoutHandler: NopJobOutputHandler{},
				StderrHandler: NopJobOutputHandler{},
			},
		},
	},
	{
		Input: `[prefs]
logPath: my/log/path
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog:        NewMemOnlyRunLog(100),
				LogPath:       filepath.Join(gUserEx.HomeDir, "my/log/path"),
				StdoutHandler: NopJobOutputHandler{},
				StderrHandler: NopJobOutputHandler{},
			},
		},
	},
	{
		Input: `[prefs]
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog:        NewMemOnlyRunLog(100),
				LogPath:       "",
				StdoutHandler: NopJobOutputHandler{},
				StderrHandler: NopJobOutputHandler{},
			},
		},
	},
	{
		Input: `[prefs]
jobOutput:
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog:        NewMemOnlyRunLog(100),
				LogPath:       "",
				StdoutHandler: NopJobOutputHandler{},
				StderrHandler: NopJobOutputHandler{},
			},
		},
	},
	{
		Input: `[prefs]
jobOutput:
    stdout:
        where: /tmp
`,
		Error: true,
	},
	{
		Input: `[prefs]
jobOutput:
    stdout:
        maxAgeDays: 10
`,
		Error: true,
	},
	{
		Input: `[prefs]
jobOutput:
    stdout:
        where: /tmp
        maxAgeDays: 10
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog:  NewMemOnlyRunLog(100),
				LogPath: "",
				StdoutHandler: FileJobOutputHandler{
					Where:      "/tmp",
					MaxAgeDays: 10,
					Suffix:     "stdout",
				},
				StderrHandler: NopJobOutputHandler{},
			},
		},
	},
	{
		Input: `[prefs]
jobOutput:
    stdout:
        where: /tmp
        maxAgeDays: 10
    stderr:
        where: /tmp2
        maxAgeDays: 100

[jobs]
- name: JobA
  cmd: exit 0
  time: '*'
  notifyOnError: false
  notifyOnFailure: false
  notifyOnSuccess: false
  jobOutput:
      stdout:
          where: /tmp3
      stderr:
          maxAgeDays: 200
- name: JobB
  cmd: exit 0
  time: '*'
  notifyOnError: false
  notifyOnFailure: false
  notifyOnSuccess: false
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog:  NewMemOnlyRunLog(100),
				LogPath: "",
				StdoutHandler: FileJobOutputHandler{
					Where:      "/tmp",
					MaxAgeDays: 10,
					Suffix:     "stdout",
				},
				StderrHandler: FileJobOutputHandler{
					Where:      "/tmp2",
					MaxAgeDays: 100,
					Suffix:     "stderr",
				},
			}, // prefs
			Jobs: map[string]*Job{
				"JobA": &Job{
					Name:            "JobA",
					Cmd:             "exit 0",
					User:            gUserEx.Username,
					FullTimeSpec:    gEverySecTimeSpec,
					ErrorHandler:    ContinueErrorHandler{},
					NotifyOnError:   NopRunRecNotifier{},
					NotifyOnFailure: NopRunRecNotifier{},
					NotifyOnSuccess: NopRunRecNotifier{},
					StdoutHandler: FileJobOutputHandler{
						Where:      "/tmp3",
						MaxAgeDays: 10,
						Suffix:     "stdout",
					},
					StderrHandler: FileJobOutputHandler{
						Where:      "/tmp2",
						MaxAgeDays: 200,
						Suffix:     "stderr",
					},
				}, // job
				"JobB": &Job{
					Name:            "JobB",
					Cmd:             "exit 0",
					User:            gUserEx.Username,
					FullTimeSpec:    gEverySecTimeSpec,
					ErrorHandler:    ContinueErrorHandler{},
					NotifyOnError:   NopRunRecNotifier{},
					NotifyOnFailure: NopRunRecNotifier{},
					NotifyOnSuccess: NopRunRecNotifier{},
					StdoutHandler: FileJobOutputHandler{
						Where:      "/tmp",
						MaxAgeDays: 10,
						Suffix:     "stdout",
					},
					StderrHandler: FileJobOutputHandler{
						Where:      "/tmp2",
						MaxAgeDays: 100,
						Suffix:     "stderr",
					},
				}, // job
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

func makeRange(start, end int) []int {
	arr := make([]int, 0, end-start)
	for i := start; i < end; i++ {
		arr = append(arr, i)
	}
	return arr
}

func TestParseFullTimeSpec(t *testing.T) {
	evens := []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22}
	threes := []int{1, 4, 7, 10, 13, 16, 19, 22}
	cases := []struct {
		str  string
		spec FullTimeSpec
	}{
		{"0 0 14", FullTimeSpec{
			OneValTimeSpec{0},
			OneValTimeSpec{0},
			OneValTimeSpec{14},
			WildcardTimeSpec{},
			WildcardTimeSpec{},
			WildcardTimeSpec{}}},
		{"0 0 14 * * 1", FullTimeSpec{
			OneValTimeSpec{0},
			OneValTimeSpec{0},
			OneValTimeSpec{14},
			WildcardTimeSpec{},
			WildcardTimeSpec{},
			OneValTimeSpec{1}}},
		{"0 0 */2 * * 1", FullTimeSpec{
			OneValTimeSpec{0},
			OneValTimeSpec{0},
			SetTimeSpec{"*/2", evens},
			WildcardTimeSpec{},
			WildcardTimeSpec{},
			OneValTimeSpec{1}}},
		{"0 0 1,4,7,10,13,16,19,22 * * 1", FullTimeSpec{
			OneValTimeSpec{0},
			OneValTimeSpec{0},
			SetTimeSpec{"1,4,7,10,13,16,19,22", threes},
			WildcardTimeSpec{},
			WildcardTimeSpec{},
			OneValTimeSpec{1}}},
		{"10,20,30,40 0 14 1 8 0-5", FullTimeSpec{
			SetTimeSpec{"10,20,30,40", []int{10, 20, 30, 40}},
			OneValTimeSpec{0},
			OneValTimeSpec{14},
			OneValTimeSpec{1},
			OneValTimeSpec{8},
			SetTimeSpec{"0-5", makeRange(0, 6)}}},
		{"0 0 R * * 1", FullTimeSpec{
			OneValTimeSpec{0},
			OneValTimeSpec{0},
			&RandomTimeSpec{desc: "R", vals: makeRange(0, 24)},
			WildcardTimeSpec{},
			WildcardTimeSpec{},
			OneValTimeSpec{1}}},
		{"0 0 R2-4 * * 1", FullTimeSpec{
			OneValTimeSpec{0},
			OneValTimeSpec{0},
			&RandomTimeSpec{desc: "R2-4", vals: makeRange(2, 5)},
			WildcardTimeSpec{},
			WildcardTimeSpec{},
			OneValTimeSpec{1}}},
	}

	for _, c := range cases {
		/*
		 * Call
		 */
		var result *FullTimeSpec
		var err error
		result, err = ParseFullTimeSpec(c.str)

		/*
		 * Test
		 */
		if err != nil {
			t.Fatalf("Got error: %v", err)
		}
		require.Equal(t, c.spec, *result)
	}
}

func TestLoadJobFile(t *testing.T) {
	for _, testCase := range gTestCases {
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
			require.Equal(t, testCase.Output.Jobs[jobName].StderrHandler, file.Jobs[jobName].StderrHandler)
			require.Equal(t, testCase.Output.Jobs[jobName].StdoutHandler, file.Jobs[jobName].StdoutHandler)
			require.Equal(t, testCase.Output.Jobs[jobName].FullTimeSpec, file.Jobs[jobName].FullTimeSpec)
			require.Equal(t, testCase.Output.Jobs[jobName].NotifyOnError, file.Jobs[jobName].NotifyOnError)
			require.Equal(t, testCase.Output.Jobs[jobName].NotifyOnFailure, file.Jobs[jobName].NotifyOnFailure)
			require.Equal(t, testCase.Output.Jobs[jobName].NotifyOnSuccess, file.Jobs[jobName].NotifyOnSuccess)
			require.Equal(t, *testCase.Output.Jobs[jobName], *file.Jobs[jobName])
		}
	}
}
