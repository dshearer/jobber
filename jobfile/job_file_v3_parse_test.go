package jobfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const gV3JobFileContents = `
version: 1.4

# Must be able
# to deal with comments.
prefs:
  # Which could be (almost) anywhere.
  runLog:
    type: memory
    maxLen: 10000


# Even here!

resultSinks:
  - &programSink
    type: program
    path: /my/program.sh

  - &sysEmailSink
    type: system-email

jobs:
  DailyBackup:
    cmd: backup daily
    # And here
    time: 0 0 14
    onError: Stop
    notifyOnError: [*programSink]
    notifyOnFailure: [*sysEmailSink]
    notifyOnSuccess: [*sysEmailSink]

# So many comments...
`

var gV3JobFile = JobFile{
	Prefs: UserPrefs{
		RunLog: NewMemOnlyRunLog(10000),
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
			NotifyOnError:   []ResultSink{ProgramResultSink{Path: "/my/program.sh"}},
			NotifyOnFailure: []ResultSink{SystemEmailResultSink{}},
			NotifyOnSuccess: []ResultSink{SystemEmailResultSink{}},
		},
	},
}

var gJobFileV3TestCases = []JobFileTestCase{
	{
		Input:  gV3JobFileContents,
		Output: gV3JobFile,
	},
	{
		Input: `
version: 1.4
prefs:
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
version: 1.4
prefs:
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
version: 1.4
prefs:
    runLog:
        type: file
        path: /dir/does/not/exist/claudius
        maxFileLen: 10m
        maxHistories: 20
`,
		Error: true,
	},
	{
		Input: `
version: 1.4
prefs:
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
		Input: `
version: 1.4
prefs:
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
		Input: `
version: 1.4
prefs:
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog:  NewMemOnlyRunLog(100),
				LogPath: "",
			},
		},
	},
	{
		Input: `
version: 1.4

resultSinks:
  - &programSink
    type: program
    path: /my/program.sh

  - &sysEmailSink
    type: system-email

  - &programSink2
    type: program
    path: /my/program2.sh

  - &fsSink
    type: filesystem
    path: /some/dir
    data: [stdout, stderr]
    maxAgeDays: 10

  - &stdoutSink
    type: stdout
    data: [stderr]

  - &socketSink
    type: socket
    proto: tcp6
    address: :1234
    data: [stdout]

jobs:
  Job1:
    time: '*'
    cmd: exit 0
  Job2:
    time: '*'
    cmd: exit 0
    notifyOnError: []
  Job3:
    time: '*'
    cmd: exit 0
    notifyOnError: [*programSink]
  Job4:
    time: '*'
    cmd: exit 0
    notifyOnError: [*programSink, *programSink]
  Job5:
    time: '*'
    cmd: exit 0
    notifyOnError: [*programSink, *sysEmailSink]
  Job6:
    time: '*'
    cmd: exit 0
    notifyOnError: [*programSink, *programSink2]
  Job7:
    time: '*'
    cmd: exit 0
    notifyOnError: [*programSink, *sysEmailSink, *programSink]
  Job8:
    time: '*'
    cmd: exit 0
    notifyOnError: [*fsSink, *stdoutSink, *socketSink]
    notifyOnSuccess: [*socketSink]
`,
		Output: JobFile{
			Prefs: UserPrefs{
				RunLog:  NewMemOnlyRunLog(100),
				LogPath: "",
			},
			Jobs: map[string]*Job{
				"Job1": &Job{
					Name:            "Job1",
					FullTimeSpec:    gEverySecTimeSpec,
					Cmd:             "exit 0",
					User:            gUserEx.Username,
					ErrorHandler:    ContinueErrorHandler{},
					NotifyOnError:   nil,
					NotifyOnFailure: nil,
					NotifyOnSuccess: nil,
				},
				"Job2": &Job{
					Name:            "Job2",
					FullTimeSpec:    gEverySecTimeSpec,
					Cmd:             "exit 0",
					User:            gUserEx.Username,
					ErrorHandler:    ContinueErrorHandler{},
					NotifyOnError:   nil,
					NotifyOnFailure: nil,
					NotifyOnSuccess: nil,
				},
				"Job3": &Job{
					Name:            "Job3",
					FullTimeSpec:    gEverySecTimeSpec,
					Cmd:             "exit 0",
					User:            gUserEx.Username,
					ErrorHandler:    ContinueErrorHandler{},
					NotifyOnError:   []ResultSink{ProgramResultSink{Path: "/my/program.sh"}},
					NotifyOnFailure: nil,
					NotifyOnSuccess: nil,
				},
				"Job4": &Job{
					Name:            "Job4",
					FullTimeSpec:    gEverySecTimeSpec,
					Cmd:             "exit 0",
					User:            gUserEx.Username,
					ErrorHandler:    ContinueErrorHandler{},
					NotifyOnError:   []ResultSink{ProgramResultSink{Path: "/my/program.sh"}},
					NotifyOnFailure: nil,
					NotifyOnSuccess: nil,
				},
				"Job5": &Job{
					Name:         "Job5",
					FullTimeSpec: gEverySecTimeSpec,
					Cmd:          "exit 0",
					User:         gUserEx.Username,
					ErrorHandler: ContinueErrorHandler{},
					NotifyOnError: []ResultSink{
						ProgramResultSink{Path: "/my/program.sh"},
						SystemEmailResultSink{},
					},
					NotifyOnFailure: nil,
					NotifyOnSuccess: nil,
				},
				"Job6": &Job{
					Name:         "Job6",
					FullTimeSpec: gEverySecTimeSpec,
					Cmd:          "exit 0",
					User:         gUserEx.Username,
					ErrorHandler: ContinueErrorHandler{},
					NotifyOnError: []ResultSink{
						ProgramResultSink{Path: "/my/program.sh"},
						ProgramResultSink{Path: "/my/program2.sh"},
					},
					NotifyOnFailure: nil,
					NotifyOnSuccess: nil,
				},
				"Job7": &Job{
					Name:         "Job7",
					FullTimeSpec: gEverySecTimeSpec,
					Cmd:          "exit 0",
					User:         gUserEx.Username,
					ErrorHandler: ContinueErrorHandler{},
					NotifyOnError: []ResultSink{
						ProgramResultSink{Path: "/my/program.sh"},
						SystemEmailResultSink{},
					},
					NotifyOnFailure: nil,
					NotifyOnSuccess: nil,
				},
				"Job8": &Job{
					Name:         "Job8",
					FullTimeSpec: gEverySecTimeSpec,
					Cmd:          "exit 0",
					User:         gUserEx.Username,
					ErrorHandler: ContinueErrorHandler{},
					NotifyOnError: []ResultSink{
						FilesystemResultSink{
							Path:       "/some/dir",
							Data:       RESULT_SINK_DATA_STDOUT | RESULT_SINK_DATA_STDERR,
							MaxAgeDays: 10,
						},
						StdoutResultSink{
							Data: RESULT_SINK_DATA_STDERR,
						},
						&SocketResultSink{
							Proto:   "tcp6",
							Address: ":1234",
							Data:    RESULT_SINK_DATA_STDOUT,
						},
					},
					NotifyOnFailure: nil,
					NotifyOnSuccess: []ResultSink{
						&SocketResultSink{
							Proto:   "tcp6",
							Address: ":1234",
							Data:    RESULT_SINK_DATA_STDOUT,
						},
					},
				},
			},
		},
	},
	{
		Input: `[unparseable
`,
		Error: true,
	},
}

func TestLoadJobFileV3(t *testing.T) {
	for _, testCase := range gJobFileV3TestCases {
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
		var raw *JobFileRaw
		raw, err = LoadJobfile(f)
		var file *JobFile
		if raw != nil {
			file, err = raw.Activate(&gUserEx)
		}

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
		for jobName := range testCase.Output.Jobs {
			require.Equal(t, testCase.Output.Jobs[jobName].FullTimeSpec, file.Jobs[jobName].FullTimeSpec)
			require.Equal(t, testCase.Output.Jobs[jobName].NotifyOnError, file.Jobs[jobName].NotifyOnError)
			require.Equal(t, testCase.Output.Jobs[jobName].NotifyOnFailure, file.Jobs[jobName].NotifyOnFailure)
			require.Equal(t, testCase.Output.Jobs[jobName].NotifyOnSuccess, file.Jobs[jobName].NotifyOnSuccess)
			require.Equal(t, *testCase.Output.Jobs[jobName], *file.Jobs[jobName])
		}
	}
}
