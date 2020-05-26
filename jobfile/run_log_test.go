package jobfile

import (
	"io/ioutil"
	"os"
	"path"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type RunLogEntrySorter []*RunLogEntry

func (self RunLogEntrySorter) Len() int {
	return len(self)
}

func (self RunLogEntrySorter) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self RunLogEntrySorter) Less(i, j int) bool {
	return self[i].Time.After(self[j].Time)
}

func checkEntriesForNil(entries []*RunLogEntry, t *testing.T) {
	for i, entry := range entries {
		require.NotNil(t, entry, "Entry %v is nil", i)
	}
}

func entriesToTimes(entries []*RunLogEntry) []time.Time {
	times := make([]time.Time, len(entries))
	for i := 0; i < len(entries); i++ {
		times[i] = entries[i].Time
	}
	return times
}

type RunLogTestSuite struct {
	suite.Suite
	makeRunLog    func() (RunLog, error)
	debugInfo     func(RunLog) string
	runLog        RunLog
	putEntries    []*RunLogEntry
	expEntryArray []*RunLogEntry
}

func NewRunLogTestSuite(makeRunLog func() (RunLog, error),
	debugInfo func(RunLog) string) *RunLogTestSuite {

	return &RunLogTestSuite{
		makeRunLog: makeRunLog,
		debugInfo:  debugInfo,
	}
}

func (self *RunLogTestSuite) SetupTest() {
	var err error
	self.runLog, err = self.makeRunLog()
	require.Nil(self.T(), err)
	require.Equal(self.T(), 0, self.runLog.Len())

	// make entries
	startTime := time.Unix(1509148800, 0) // midnight 28 Oct 2017 UTC
	self.putEntries = []*RunLogEntry{
		&RunLogEntry{
			JobName: "Entry 0",
			Time:    startTime,
		},
		&RunLogEntry{
			JobName: "Entry 1",
			Time:    startTime.Add(time.Hour),
		},
		&RunLogEntry{
			JobName: "Entry \t2",
			Time:    startTime.Add(time.Hour),
		},
		&RunLogEntry{
			JobName: "Entry 3\t",
			Time:    startTime.Add(5 * time.Hour),
		},
		&RunLogEntry{
			JobName: "Entry \n4",
			Time:    startTime.Add(6 * time.Hour),
		},
		&RunLogEntry{
			JobName: "Entry 5\n",
			Time:    startTime.Add(5 * time.Hour),
		},
		&RunLogEntry{
			JobName: "Entry 6\\",
			Time:    startTime.Add(7 * time.Hour),
		},
	}

	// put entries
	for i, entry := range self.putEntries {
		self.T().Logf("Putting entry %v", i)
		tmp := self.runLog.Len()
		err := self.runLog.Put(*entry)
		if self.debugInfo != nil {
			self.T().Logf(self.debugInfo(self.runLog))
		}
		require.Nil(self.T(), err, "runLog.Put returned error: %v", err)
		require.Equal(self.T(), tmp+1, self.runLog.Len())
	}

	// make expected entries
	self.expEntryArray = make([]*RunLogEntry, len(self.putEntries))
	copy(self.expEntryArray, self.putEntries)
	sort.Sort(RunLogEntrySorter(self.expEntryArray))
}

func (self *RunLogTestSuite) TestGetFromTime() {
	latestTime := self.expEntryArray[0].Time

	entries, err := self.runLog.GetFromTime(self.expEntryArray[0].Time,
		self.expEntryArray[0].Time)
	require.Nil(self.T(), err)
	require.Equal(
		self.T(),
		0,
		len(entries),
	)

	entries, err = self.runLog.GetAll()
	require.Nil(self.T(), err)
	require.Equal(
		self.T(),
		entriesToTimes(self.expEntryArray),
		entriesToTimes(entries),
	)

	entries_1, err := self.runLog.GetAll()
	require.Nil(self.T(), err)
	entries_2, err := self.runLog.GetFromTime(latestTime)
	require.Nil(self.T(), err)
	require.Equal(
		self.T(),
		entries_1,
		entries_2,
	)
}

func (self *RunLogTestSuite) TestGetFromIndex() {
	entries, err := self.runLog.GetFromIndex(0, 0)
	require.Nil(self.T(), err)
	require.Equal(
		self.T(),
		[]*RunLogEntry{},
		entries,
	)

	entries, err = self.runLog.GetFromIndex(1, 1)
	require.Nil(self.T(), err)
	require.Equal(
		self.T(),
		[]*RunLogEntry{},
		entries,
	)

	entries, err = self.runLog.GetAll()
	checkEntriesForNil(entries, self.T())
	require.Nil(self.T(), err)
	require.Equal(
		self.T(),
		entriesToTimes(self.expEntryArray),
		entriesToTimes(entries),
	)

	entry_1, err := self.runLog.GetAll()
	require.Nil(self.T(), err)
	entry_2, err := self.runLog.GetFromIndex(0)
	require.Nil(self.T(), err)
	require.Equal(
		self.T(),
		entry_1,
		entry_2,
	)

	midIdx := self.runLog.Len() / 2
	entries, err = self.runLog.GetFromIndex(midIdx)
	require.Nil(self.T(), err)
	require.Equal(
		self.T(),
		entriesToTimes(self.expEntryArray[midIdx:]),
		entriesToTimes(entries),
	)

	entries, err = self.runLog.GetFromIndex(1, self.runLog.Len()-1)
	require.Nil(self.T(), err)
	require.Equal(
		self.T(),
		entriesToTimes(self.expEntryArray[1:len(self.expEntryArray)-1]),
		entriesToTimes(entries),
	)
}

func TestMemOnlyRunLog(t *testing.T) {
	makeRunLog := func() (RunLog, error) {
		return NewMemOnlyRunLog(10), nil
	}
	suite.Run(t, NewRunLogTestSuite(makeRunLog, nil))
}

func TestFileRunLog(t *testing.T) {
	tmpDir, err := ioutil.TempDir("/tmp", "unittest-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	path := path.Join(tmpDir, "runlog")
	makeRunLog := func() (RunLog, error) {
		maxFileLen := (gLogEntryLen+1)*2 + gLogEntryLen
		log, err := NewFileRunLog(path, maxFileLen, 3)
		if err != nil {
			return nil, err
		}
		log.(*fileRunLog).deleteAll()
		return log, nil
	}
	debugInfo := func(log RunLog) string {
		return log.(*fileRunLog).debugInfo()
	}
	suite.Run(t, NewRunLogTestSuite(makeRunLog, debugInfo))
}
