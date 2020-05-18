package jobfile

import (
	"time"

	"github.com/dshearer/jobber/common"
)

type RunLogEntry struct {
	JobName string
	Time    time.Time
	Fate    common.SubprocFate
	Result  JobStatus
}

/*
This is a database of job runs.  It may be backed by a file, or it may
not: this depends on the user's preferences specified in the jobfile.
*/
type RunLog interface {
	/*
	    	There are two ways to use this method.

	    	"GetFromTime(t)": Get all entries for runs that started no
	    	later than t.

	    	"GetFromTime(t1, t2)": Get all entries that started no later
	    	than t1 but later than t2.

	   The entries are returned in order of start time, descending.
	*/
	GetFromTime(maxTime time.Time, minTime ...time.Time) ([]*RunLogEntry, error)

	/*
		There are two ways to use this method.

		"GetFromIndex(i)": Get all entries with index >= i.

		"GetFromIndex(i, j)": Get all entries with index >= i but
		< j.
	*/
	GetFromIndex(minIdx int, maxIdx ...int) ([]*RunLogEntry, error)

	/*
	   Get all entries.
	*/
	GetAll() ([]*RunLogEntry, error)

	Len() int

	Put(entry RunLogEntry) error
}
