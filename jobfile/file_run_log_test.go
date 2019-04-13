package jobfile

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type EntryEncodeDecodeTestCase struct {
	entry   RunLogEntry
	encoded string
}

var EntryEncodeDecodeTestCases = []EntryEncodeDecodeTestCase{
	{
		RunLogEntry{
			"My\n\nDumb\tJob",
			time.Unix(1506313655, 0),
			true,
			JobGood,
			time.Second,
		},
		"My\\n\\nDumb\\tJob\t1506313655000000000\ttrue\tGood\t1s                ",
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
