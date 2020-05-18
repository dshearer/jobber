package jobfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/dshearer/jobber/common"
	"github.com/stretchr/testify/require"
)

type EntryEncodeDecodeTestCase struct {
	entry   RunLogEntry
	encoded string
}

var EntryEncodeDecodeTestCases = []EntryEncodeDecodeTestCase{
	{
		RunLogEntry{
			JobName:  "My\n\nDumb\tJob",
			Time:     time.Unix(1506313655, 0),
			Fate:     common.SubprocFateSucceeded,
			Result:   JobGood,
			ExecTime: time.Second,
		},
		"My\\n\\nDumb\\tJob\t1506313655000000000\tsucceeded\tGood\t1s           ",
	},
	{
		RunLogEntry{
			JobName:  "My\n\nDumb\tJob",
			Time:     time.Unix(1506313655, 0),
			Fate:     common.SubprocFateFailed,
			Result:   JobGood,
			ExecTime: time.Second,
		},
		"My\\n\\nDumb\\tJob\t1506313655000000000\tfailed\tGood\t1s              ",
	},
	{
		RunLogEntry{
			JobName:  "My\n\nDumb\tJob",
			Time:     time.Unix(1506313655, 0),
			Fate:     common.SubprocFateCancelled,
			Result:   JobGood,
			ExecTime: time.Second,
		},
		"My\\n\\nDumb\\tJob\t1506313655000000000\tcancelled\tGood\t1s           ",
	},
}

var EntryDecodeTestCases = []EntryEncodeDecodeTestCase{
	// deprecated values for "Fate"
	{
		RunLogEntry{
			JobName:  "My\n\nDumb\tJob",
			Time:     time.Unix(1506313655, 0),
			Fate:     common.SubprocFateSucceeded,
			Result:   JobGood,
			ExecTime: time.Second,
		},
		"My\\n\\nDumb\\tJob\t1506313655000000000\ttrue\tGood\t1s                ",
	},
	{
		RunLogEntry{
			JobName:  "My\n\nDumb\tJob",
			Time:     time.Unix(1506313655, 0),
			Fate:     common.SubprocFateFailed,
			Result:   JobGood,
			ExecTime: time.Second,
		},
		"My\\n\\nDumb\\tJob\t1506313655000000000\tfalse\tGood\t1s                ",
	},
}

func TestEntryEncodeDecode(t *testing.T) {
	for _, testCase := range EntryEncodeDecodeTestCases {
		// test encodeRunLogEntry
		require.Equal(
			t,
			testCase.encoded,
			encodeRunLogEntry(&testCase.entry),
		)

		// test decodeRunLogEntry
		actualEntry, _ := decodeRunLogEntry(testCase.encoded)
		require.Equal(t, testCase.entry, *actualEntry)
	}
}

func TestEntryDecode(t *testing.T) {
	for _, testCase := range EntryDecodeTestCases {
		// test decodeRunLogEntry
		actualEntry, _ := decodeRunLogEntry(testCase.encoded)
		require.Equal(t, testCase.entry, *actualEntry)
	}
}

func TestWithOneEntry(t *testing.T) {
	/*
		Set up
	*/

	// make entry
	var entry RunLogEntry
	entry.JobName = "TestJob"
	entry.ExecTime = time.Minute * 4
	entry.Fate = common.SubprocFateCancelled
	entry.Result = JobGood
	entry.Time = time.Unix(1506313655, 0)

	// make log file
	f, err := ioutil.TempFile("", "Testing")
	if err != nil {
		panic(fmt.Sprintf("Failed to make tempfile: %v", err))
	}
	defer f.Close()
	defer os.Remove(f.Name())
	logFilePath := f.Name()
	log, err := NewFileRunLog(logFilePath, 10*(1<<20), 100)
	if err != nil {
		t.Fatal(err)
	}

	// add entry
	if err := log.Put(entry); err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 1, log.Len())

	/*
		Call
	*/
	entries, err := log.GetAll()
	if err != nil {
		t.Fatal(err)
	}

	/*
		Test
	*/
	require.Equal(t, 1, len(entries))
	require.Equal(t, entry, *entries[0])
}
