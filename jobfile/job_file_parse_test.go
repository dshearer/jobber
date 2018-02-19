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

var gUserEx = user.User{Username: "bob", HomeDir: "/home/bob"}

var EverySecTimeSpec = FullTimeSpec{
	Sec:  &WildcardTimeSpec{},
	Min:  &WildcardTimeSpec{},
	Hour: &WildcardTimeSpec{},
	Mday: &WildcardTimeSpec{},
	Mon:  &WildcardTimeSpec{},
	Wday: &WildcardTimeSpec{},
}

var gDailyBackup = Job{
	Name: "DailyBackup",
	Cmd:  "backup daily",
	User: gUserEx.Username,
	FullTimeSpec: FullTimeSpec{
		Sec:  &OneValTimeSpec{0},
		Min:  &OneValTimeSpec{0},
		Hour: &OneValTimeSpec{14},
		Mday: &WildcardTimeSpec{},
		Mon:  &WildcardTimeSpec{},
		Wday: &WildcardTimeSpec{},
	},
	ErrorHandler:    &ErrorHandlerStop,
	NotifyOnError:   false,
	NotifyOnFailure: true,
	NotifyOnSuccess: false,
}

var gWeeklyBackup = Job{
	Name: "WeeklyBackup",
	Cmd: `multi-
line
script
`,
	User: gUserEx.Username,
	FullTimeSpec: FullTimeSpec{
		Sec:  &OneValTimeSpec{0},
		Min:  &OneValTimeSpec{0},
		Hour: &OneValTimeSpec{14},
		Mday: &WildcardTimeSpec{},
		Mon:  &WildcardTimeSpec{},
		Wday: &OneValTimeSpec{1},
	},
	ErrorHandler:    &ErrorHandlerBackoff,
	NotifyOnError:   true,
	NotifyOnFailure: false,
	NotifyOnSuccess: false,
}

var gSuccessReport = Job{
	Name: "SuccessReport",
	Cmd: `multi-
line
script
`,
	User: gUserEx.Username,
	FullTimeSpec: FullTimeSpec{
		Sec:  &OneValTimeSpec{0},
		Min:  &OneValTimeSpec{0},
		Hour: &OneValTimeSpec{14},
		Mday: &WildcardTimeSpec{},
		Mon:  &WildcardTimeSpec{},
		Wday: &OneValTimeSpec{1},
	},
	ErrorHandler:    &ErrorHandlerBackoff,
	NotifyOnError:   false,
	NotifyOnFailure: false,
	NotifyOnSuccess: true,
}

var gJobA = Job{
	Name:            "JobA",
	Cmd:             "whatever",
	User:            gUserEx.Username,
	FullTimeSpec:    EverySecTimeSpec,
	ErrorHandler:    &ErrorHandlerBackoff,
	NotifyOnError:   true,
	NotifyOnFailure: false,
	NotifyOnSuccess: false,
}

var gJobB = Job{
	Name:            "JobB",
	Cmd:             "whatever",
	User:            gUserEx.Username,
	FullTimeSpec:    EverySecTimeSpec,
	ErrorHandler:    &ErrorHandlerBackoff,
	NotifyOnError:   true,
	NotifyOnFailure: false,
	NotifyOnSuccess: false,
}

const gNewJobFileContents = `
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
  cmd: |
    multi-
    line
    script
  time: 0 0 14 * * 1
  onError: Backoff
  notifyOnError: false
  notifyOnFailure: false
  notifyOnSuccess: true

# So many comments...
`

var gNewJobFile = JobFile{
	Prefs: UserPrefs{
		RunLog: NewMemOnlyRunLog(100),
	},
	Jobs: []*Job{
		&gDailyBackup,
		&gWeeklyBackup,
		&gSuccessReport,
	},
}

const gLegacyJobFileContents = `---
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

var gLegacyJobFile = JobFile{
	Prefs: UserPrefs{
		RunLog: NewMemOnlyRunLog(100),
	},
	Jobs: []*Job{
		&gDailyBackup,
		&gWeeklyBackup,
		&gJobA,
		&gJobB,
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

var gTestCases = []JobFileTestCase{
	{
		Input:  gNewJobFileContents,
		Output: gNewJobFile,
	},
	{
		Input:  gLegacyJobFileContents,
		Output: gLegacyJobFile,
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
			Jobs: []*Job{&gDailyBackup},
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
			&OneValTimeSpec{0},
			&OneValTimeSpec{0},
			&OneValTimeSpec{14},
			&WildcardTimeSpec{},
			&WildcardTimeSpec{},
			&WildcardTimeSpec{}}},
		{"0 0 14 * * 1", FullTimeSpec{
			&OneValTimeSpec{0},
			&OneValTimeSpec{0},
			&OneValTimeSpec{14},
			&WildcardTimeSpec{},
			&WildcardTimeSpec{},
			&OneValTimeSpec{1}}},
		{"0 0 */2 * * 1", FullTimeSpec{
			&OneValTimeSpec{0},
			&OneValTimeSpec{0},
			&SetTimeSpec{"*/2", evens},
			&WildcardTimeSpec{},
			&WildcardTimeSpec{},
			&OneValTimeSpec{1}}},
		{"0 0 1,4,7,10,13,16,19,22 * * 1", FullTimeSpec{
			&OneValTimeSpec{0},
			&OneValTimeSpec{0},
			&SetTimeSpec{"1,4,7,10,13,16,19,22", threes},
			&WildcardTimeSpec{},
			&WildcardTimeSpec{},
			&OneValTimeSpec{1}}},
		{"10,20,30,40 0 14 1 8 0-5", FullTimeSpec{
			&SetTimeSpec{"10,20,30,40", []int{10, 20, 30, 40}},
			&OneValTimeSpec{0},
			&OneValTimeSpec{14},
			&OneValTimeSpec{1},
			&OneValTimeSpec{8},
			&SetTimeSpec{"0-5", makeRange(0, 6)}}},
		{"0 0 R * * 1", FullTimeSpec{
			&OneValTimeSpec{0},
			&OneValTimeSpec{0},
			&RandomTimeSpec{desc: "R", vals: makeRange(0, 24)},
			&WildcardTimeSpec{},
			&WildcardTimeSpec{},
			&OneValTimeSpec{1}}},
		{"0 0 R2-4 * * 1", FullTimeSpec{
			&OneValTimeSpec{0},
			&OneValTimeSpec{0},
			&RandomTimeSpec{desc: "R2-4", vals: makeRange(2, 5)},
			&WildcardTimeSpec{},
			&WildcardTimeSpec{},
			&OneValTimeSpec{1}}},
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
		require.NotNil(t, file.Prefs.Notifier)

		// can't compare functions
		testCase.Output.Prefs.Notifier = nil
		file.Prefs.Notifier = nil
		for _, job := range testCase.Output.Jobs {
			job.ErrorHandler.Apply = nil
		}
		for _, job := range file.Jobs {
			require.NotNil(t, job.ErrorHandler)
			job.ErrorHandler.Apply = nil
		}

		require.Equal(t, testCase.Output.Prefs, file.Prefs)
		require.Equal(t, len(testCase.Output.Jobs), len(file.Jobs))
		for i := 0; i < len(file.Jobs); i++ {
			require.Equal(t, testCase.Output.Jobs[i], file.Jobs[i])
		}
	}
}
