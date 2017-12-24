package jobfile

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

const NewJobFileEx string = `
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

const LegacyJobFileEx string = `---
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

const UsernameEx string = "bob"

var EverySecTimeSpec FullTimeSpec = FullTimeSpec{WildcardTimeSpec{},
	WildcardTimeSpec{},
	WildcardTimeSpec{},
	WildcardTimeSpec{},
	WildcardTimeSpec{},
	WildcardTimeSpec{}}

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
		{"10,20 0 14 1 8 0-5", FullTimeSpec{
			SetTimeSpec{"10,20", []int{10, 20}},
			OneValTimeSpec{0},
			OneValTimeSpec{14},
			OneValTimeSpec{1},
			OneValTimeSpec{8},
			SetTimeSpec{"0-5", []int{0, 1, 2, 3, 4, 5}}}},
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

func TestReadNewJobberFile(t *testing.T) {
	/*
	 * Set up
	 */
	f, err := ioutil.TempFile("", "Testing")
	if err != nil {
		panic(fmt.Sprintf("Failed to make tempfile: %v", err))
	}
	defer os.Remove(f.Name())
	f.Write([]byte(NewJobFileEx))
	f.Close()

	/*
	 * Call
	 */
	var file *JobFile
	file, err = LoadJobFile(f.Name(), UsernameEx)

	/*
	 * Test
	 */
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	if file == nil {
		t.Fatalf("jobs == nil")
	}

	// test prefs
	require.NotNil(t, file.Prefs.Notifier)
	require.NotNil(t, file.Prefs.RunLog)

	// test jobs
	require.Equal(t, 3, len(file.Jobs))

	// test DailyBackup
	daily := file.Jobs[0]
	require.Equal(t, "DailyBackup", daily.Name)
	require.Equal(t, "backup daily", daily.Cmd)
	require.Equal(t, &ErrorHandlerStop, daily.ErrorHandler)
	require.Equal(t, false, daily.NotifyOnError)
	require.Equal(t, true, daily.NotifyOnFailure)
	require.Equal(t, false, daily.NotifyOnSuccess)

	// test WeeklyBackup
	weekly := file.Jobs[1]
	require.Equal(t, "WeeklyBackup", weekly.Name)
	require.Equal(t, `multi-
line
script
`, weekly.Cmd)
	require.Equal(t, &ErrorHandlerBackoff, weekly.ErrorHandler)
	require.Equal(t, true, weekly.NotifyOnError)
	require.Equal(t, false, weekly.NotifyOnFailure)
	require.Equal(t, false, weekly.NotifyOnSuccess)

	notifysuccess := file.Jobs[2]
	require.Equal(t, "SuccessReport", notifysuccess.Name)
	require.Equal(t, `multi-
line
script
`, notifysuccess.Cmd)
	require.Equal(t, &ErrorHandlerBackoff, notifysuccess.ErrorHandler)
	require.Equal(t, false, notifysuccess.NotifyOnError)
	require.Equal(t, false, notifysuccess.NotifyOnFailure)
	require.Equal(t, true, notifysuccess.NotifyOnSuccess)
}

func TestReadLegacyJobberFile(t *testing.T) {
	/*
	 * Set up
	 */
	f, err := ioutil.TempFile("", "Testing")
	if err != nil {
		panic(fmt.Sprintf("Failed to make tempfile: %v", err))
	}
	defer os.Remove(f.Name())
	f.Write([]byte(LegacyJobFileEx))
	f.Close()

	/*
	 * Call
	 */
	var file *JobFile
	file, err = LoadJobFile(f.Name(), UsernameEx)

	/*
	 * Test
	 */
	if err != nil {
		t.Fatalf("Got error: %v", err)
	}
	if file == nil {
		t.Fatalf("jobs == nil")
	}
	require.Equal(t, 4, len(file.Jobs))

	// test DailyBackup
	daily := file.Jobs[0]
	require.Equal(t, "DailyBackup", daily.Name)
	require.Equal(t, "backup daily", daily.Cmd)
	require.Equal(t, &ErrorHandlerStop, daily.ErrorHandler)
	require.Equal(t, false, daily.NotifyOnError)
	require.Equal(t, true, daily.NotifyOnFailure)

	// test WeeklyBackup
	weekly := file.Jobs[1]
	require.Equal(t, "WeeklyBackup", weekly.Name)
	require.Equal(t, `multi-
line
script
`, weekly.Cmd)
	require.Equal(t, &ErrorHandlerBackoff, weekly.ErrorHandler)
	require.Equal(t, true, weekly.NotifyOnError)
	require.Equal(t, false, weekly.NotifyOnFailure)

	// test JobA
	jobA := file.Jobs[2]
	require.Equal(t, "JobA", jobA.Name)
	require.Equal(t, EverySecTimeSpec, jobA.FullTimeSpec)

	// test JobB
	jobB := file.Jobs[3]
	require.Equal(t, "JobB", jobB.Name)
	require.Equal(t, EverySecTimeSpec, jobB.FullTimeSpec)
}

const JobberFileWithNoPrefsEx string = `
[jobs]
- name: DailyBackup
  cmd: backup daily
  time: 0 0 14
  onError: Stop
  notifyOnError: false
  notifyOnFailure: true
`

func TestJobberFileWithNoPrefs(t *testing.T) {
	/*
	 * Set up
	 */
	f, err := ioutil.TempFile("", "Testing")
	if err != nil {
		panic(fmt.Sprintf("Failed to make tempfile: %v", err))
	}
	defer os.Remove(f.Name())
	f.Write([]byte(JobberFileWithNoPrefsEx))
	f.Close()

	/*
	 * Call
	 */
	var file *JobFile
	file, err = LoadJobFile(f.Name(), UsernameEx)

	/*
	 * Test
	 */
	require.Nil(t, err, "%v", err)
	require.NotNil(t, file)
	require.NotNil(t, file.Prefs.RunLog)
	require.NotNil(t, file.Prefs.Notifier)
}

const JobberFileWithMemOnlyRunLogEx string = `
[prefs]
runLog:
    type: memory
    maxLen: 10
`

func TestJobberFileWithMemOnlyRunLog(t *testing.T) {
	/*
	 * Set up
	 */
	f, err := ioutil.TempFile("", "Testing")
	if err != nil {
		panic(fmt.Sprintf("Failed to make tempfile: %v", err))
	}
	defer os.Remove(f.Name())
	f.Write([]byte(JobberFileWithMemOnlyRunLogEx))
	f.Close()

	/*
	 * Call
	 */
	var file *JobFile
	file, err = LoadJobFile(f.Name(), UsernameEx)

	/*
	 * Test
	 */
	require.Nil(t, err, "%v", err)
	require.NotNil(t, file)
	require.NotNil(t, file.Prefs.Notifier)
	require.NotNil(t, file.Prefs.RunLog)

	// check run log type
	require.IsType(t, &memOnlyRunLog{}, file.Prefs.RunLog)

	// check max len
	memRunLog := file.Prefs.RunLog.(*memOnlyRunLog)
	require.Equal(t, 10, memRunLog.MaxLen())
}

const JobberFileWithFileRunLogEx string = `
[prefs]
runLog:
    type: file
    path: /tmp/claudius
    maxFileLen: 10m
    maxHistories: 20
`

func TestJobberFileWithFileRunLog(t *testing.T) {
	/*
	 * Set up
	 */
	f, err := ioutil.TempFile("", "Testing")
	if err != nil {
		panic(fmt.Sprintf("Failed to make tempfile: %v", err))
	}
	defer os.Remove(f.Name())
	f.Write([]byte(JobberFileWithFileRunLogEx))
	f.Close()

	/*
	 * Call
	 */
	var file *JobFile
	file, err = LoadJobFile(f.Name(), UsernameEx)

	/*
	 * Test
	 */
	require.Nil(t, err, "%v", err)
	require.NotNil(t, file)
	require.NotNil(t, file.Prefs.Notifier)
	require.NotNil(t, file.Prefs.RunLog)

	// check run log type
	require.IsType(t, &fileRunLog{}, file.Prefs.RunLog)

	// check path
	fileRunLog := file.Prefs.RunLog.(*fileRunLog)
	require.Equal(t, "/tmp/claudius", fileRunLog.FilePath())

	// check max file len
	require.Equal(t, int64(10*(1<<20)), fileRunLog.MaxFileLen())

	// check max histories
	require.Equal(t, 20, fileRunLog.MaxHistories())
}

func TestLoadJobFileWithMissingJobberFile(t *testing.T) {
	/*
	 * Call
	 */
	file, err := LoadJobFile("/invalid/path", UsernameEx)

	/*
	 * Test
	 */
	require.Nil(t, file)
	require.NotNil(t, err)
	require.True(t, os.IsNotExist(err))
}
