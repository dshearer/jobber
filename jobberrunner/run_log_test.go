package main

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"sort"
	"testing"
	"time"
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

func entriesToTimes(entries []*RunLogEntry) []time.Time {
	times := make([]time.Time, len(entries))
	for i := 0; i < len(entries); i++ {
		times[i] = entries[i].Time
	}
	return times
}

type RunLogTestSuite struct {
	suite.Suite
	makeRunLog    func() RunLog
	runLog        RunLog
	putEntries    []*RunLogEntry
	expEntryArray []*RunLogEntry
}

func NewRunLogTestSuite(makeRunLog func() RunLog) *RunLogTestSuite {
	return &RunLogTestSuite{makeRunLog: makeRunLog}
}

func (self *RunLogTestSuite) SetupTest() {
	self.runLog = self.makeRunLog()

	// make entries
	now := time.Now()
	self.putEntries = []*RunLogEntry{
		&RunLogEntry{Time: now},
		&RunLogEntry{Time: now.Add(time.Hour)},
		&RunLogEntry{Time: now.Add(time.Hour)},
		&RunLogEntry{Time: now.Add(5 * time.Hour)},
		&RunLogEntry{Time: now.Add(6 * time.Hour)},
		&RunLogEntry{Time: now.Add(2 * time.Hour)},
		&RunLogEntry{Time: now.Add(7 * time.Hour)},
	}

	// put entries
	for _, entry := range self.putEntries {
		tmp := self.runLog.Len()
		self.runLog.Put(*entry)
		require.Equal(self.T(), tmp+1, self.runLog.Len())
	}

	// make expected entries
	self.expEntryArray = make([]*RunLogEntry, len(self.putEntries))
	copy(self.expEntryArray, self.putEntries)
	sort.Sort(RunLogEntrySorter(self.expEntryArray))
}

func (self *RunLogTestSuite) TestGetFromTime() {
	earliestTime := self.expEntryArray[len(self.expEntryArray)-1].Time

	require.Equal(
		self.T(),
		[]*RunLogEntry{},
		self.runLog.GetFromTime(self.expEntryArray[0].Time,
			self.expEntryArray[0].Time),
	)

	require.Equal(
		self.T(),
		entriesToTimes(self.expEntryArray),
		entriesToTimes(self.runLog.GetFromTime()),
	)

	require.Equal(
		self.T(),
		self.runLog.GetFromTime(),
		self.runLog.GetFromTime(earliestTime),
	)
}

func (self *RunLogTestSuite) TestGetFromIndex() {
	require.Equal(
		self.T(),
		[]*RunLogEntry{},
		self.runLog.GetFromIndex(0, 0),
	)

	require.Equal(
		self.T(),
		[]*RunLogEntry{},
		self.runLog.GetFromIndex(1, 1),
	)

	require.Equal(
		self.T(),
		entriesToTimes(self.expEntryArray),
		entriesToTimes(self.runLog.GetFromIndex()),
	)

	require.Equal(
		self.T(),
		self.runLog.GetFromIndex(),
		self.runLog.GetFromIndex(0),
	)

	midIdx := self.runLog.Len() / 2
	require.Equal(
		self.T(),
		entriesToTimes(self.expEntryArray[midIdx:]),
		entriesToTimes(self.runLog.GetFromIndex(midIdx)),
	)

	require.Equal(
		self.T(),
		entriesToTimes(self.expEntryArray[1:len(self.expEntryArray)-1]),
		entriesToTimes(self.runLog.GetFromIndex(1, self.runLog.Len()-1)),
	)
}

func TestRunLog(t *testing.T) {
	makeRunLog := func() RunLog {
		return NewMemOnlyRunLog(10)
	}
	suite.Run(t, NewRunLogTestSuite(makeRunLog))
}
