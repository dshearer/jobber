package main

import (
	"github.com/dshearer/jobber/jobfile"
	"time"
)

type RunLogEntry struct {
	Job       *jobfile.Job
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
	   Get the entries for runs that started between fromTo[0]
	   (inclusive) and fromTo[1] (exclusive).  If len(fromTo) == 1,
	   fromTo[1] defaults to just past the latest entry's start time.
	   If len(fromTo) == 0, fromTo[0] defaults to the start time of the
	   earliest entry, and fromTo[1] defaults to just past the latest
	   entry's start time.

	   The entries are returned in order of start time, descending.
	*/
	GetFromTime(fromTo ...time.Time) []*RunLogEntry

	/*
		Get the entries between index fromTo[0] (inclusive) and index
		fromTo[1] (exclusive).  Entries are indexed starting with the
		latest one.

		If len(fromTo) == 1, fromTo[1] defaults to Len().  If
		len(fromTo) == 0, fromTo[0] defaults to 0 and fromTo[1]
		defaults to Len()
	*/
	GetFromIndex(fromTo ...int) []*RunLogEntry

	Len() int

	Put(entry RunLogEntry)
}
