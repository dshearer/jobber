package main

import (
	"github.com/dshearer/jobber/jobfile"
	"time"
)

type RunLogEntry struct {
	JobName   string
	Time      time.Time
	Succeeded bool
	Result    jobfile.JobStatus
}

/*
This is a database of job runs.  It may be backed by a file, or it may
not: this depends on the user's preferences specified in the jobfile.
*/
type RunLog interface {
	/*
	    	Get the entries for runs that started no later than fromTo[0]
	    	and later than fromTo[1].

	    	If len(fromTo) == 1, then fromTo[1] defaults to a time earlier
	    	than the earliest entry.

	    	If len(fromTo) == 0, then fromTo[0] defaults to the time of
	    	the latest entry, and fromTo[1] defaults to a time earlier
	    	than the earliest entry.

	   The entries are returned in order of start time, descending.
	*/
	GetFromTime(fromTo ...time.Time) ([]*RunLogEntry, error)

	/*
		Get the entries between index fromTo[0] (inclusive) and index
		fromTo[1] (exclusive).  Entries are indexed starting with the
		latest one.

		If len(fromTo) == 1, fromTo[1] defaults to Len().

		If len(fromTo) == 0, fromTo[0] defaults to 0 and fromTo[1]
		defaults to Len()
	*/
	GetFromIndex(fromTo ...int) ([]*RunLogEntry, error)

	Len() int

	Put(entry RunLogEntry) error
}
