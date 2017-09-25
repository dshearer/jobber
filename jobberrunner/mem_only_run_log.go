package main

import (
	"sort"
	"time"
)

/*
This is an impl of RunLog that is not backed by a file.  In order to
avoid running out of memory, it has a max length, after which it starts
throwing out the oldest entries.
*/
type memOnlyRunLog struct {
	/*
	   We need to support method Get, which returns entries in
	   descending start-time order.  We also need to support method
	   Put, which will usually be called in ASCENDING start-time order.
	   Moreover, Put will be called more frequently than Get.

	   Will keep a list of entries in ascending start-time order,
	   which means that Put will usually run in constant-time.
	*/

	entries []*RunLogEntry
}

func NewMemOnlyRunLog(maxLen int) *memOnlyRunLog {
	if maxLen <= 0 {
		panic("maxLen must be > 0")
	}

	log := memOnlyRunLog{
		entries: make([]*RunLogEntry, 0, maxLen),
	}
	return &log
}

func (self *memOnlyRunLog) Put(newEntry RunLogEntry) {
	/*
	   If the entries array would be too long after inserting the new
	   entry, we need to remove an entry first.  We remove the oldest
	   entry.

	   Remember: self.entries is sorted in ascending order.
	*/

	// assertions
	if cap(self.entries) == 0 {
		panic("Capacity is 0")
	}

	if len(self.entries)+1 > cap(self.entries) {
		// if the new entry is older than any other, do nothing
		if newEntry.Time.Before(self.entries[0].Time) {
			return
		} else {
			// remove oldest entry
			copy(self.entries, self.entries[1:])
			self.entries = self.entries[:len(self.entries)-1]
		}
	}

	// add the entry
	self.entries = append(self.entries, &newEntry)

	// make sure the array is sorted
	for i := len(self.entries) - 1; i >= 1; i-- {
		if newEntry.Time.Before(self.entries[i-1].Time) {
			// swap
			self.entries[i-1], self.entries[i] =
				self.entries[i], self.entries[i-1]
		} else {
			break
		}
	}

}

func reverseEntryArray(array []*RunLogEntry) []*RunLogEntry {
	result := make([]*RunLogEntry, len(array))
	i := 0
	for j := len(array) - 1; j >= 0; j-- {
		result[i] = array[j]
		i++
	}
	return result
}

func (self *memOnlyRunLog) GetFromTime(fromTo ...time.Time,
) []*RunLogEntry {

	/*
	   Let [e_0, ..., e_n] be the (ascending-ordered) list of entries
	   (self.entries).

	   We must return a descending-ordered sublist of entries
	   [e_j, ..., e_i] (j <= i) s.t.
	    	   - e_j.Time < to
	    	   - e_(j+1).Time >= to
	    	   - e_i.Time >= from
	    	   - e_(i-1).Time < from

	*/

	/*
		0 2 4 6 8 10
		0 1 2 3 4 5

		Find (2, 7) => (1, 4)

		1 == smallest idx s.t. arr[i] >= 2.
		4 == smallest idx s.t. arr[i] >= 7.
	*/

	if len(fromTo) > 2 {
		panic("Too many args.")
	}

	if len(self.entries) == 0 {
		return []*RunLogEntry{}
	}

	var to time.Time
	if len(fromTo) >= 2 {
		to = fromTo[1]
	} else {
		// set *to* to just after the latest entry's start time
		to = self.entries[len(self.entries)-1].Time.Add(time.Second)
	}

	var from time.Time
	if len(fromTo) >= 1 {
		from = fromTo[0]
	} else {
		// set *from* to start time of earliest entry
		from = self.entries[0].Time
	}

	if from.After(to) {
		panic("from is after to")
	}

	// do binary search to find beginning of range
	startIdx := sort.Search(len(self.entries), func(i int) bool {
		return !self.entries[i].Time.Before(from)
	})
	if startIdx == len(self.entries) {
		return []*RunLogEntry{}
	}

	// do binary search to find end of range
	endIdx := sort.Search(len(self.entries), func(i int) bool {
		return !self.entries[i].Time.Before(to)
	})

	// return in reverse order
	return reverseEntryArray(self.entries[startIdx:endIdx])
}

func (self *memOnlyRunLog) GetFromIndex(fromTo ...int) []*RunLogEntry {

	/*
	   Remember: self.entries is sorted in ascending order.  But we
	   must return in descending order.


	*/

	if len(fromTo) > 2 {
		panic("Too many args.")
	}

	if len(self.entries) == 0 {
		return []*RunLogEntry{}
	}

	var to int
	if len(fromTo) >= 2 {
		to = fromTo[1]
	} else {
		to = len(self.entries)
	}

	var from int
	if len(fromTo) >= 1 {
		from = fromTo[0]
	} else {
		from = 0
	}

	if from > to {
		panic("from > to")
	}

	/*
			self.entries is sorted in ascending order.  We must return in
			descending order.

		    self.entries: 0 1 2 3 4 5 6 7
		                  7 6 5 4 3 2 1 0

		    If from == 1 and to == 3 => (5, 7)
		    If from == 0 and to == 3 => (5, 8)
	*/

	// find entries
	actualTo := len(self.entries) - from
	actualFrom := len(self.entries) - to

	// reverse them
	return reverseEntryArray(self.entries[actualFrom:actualTo])
}

func (self *memOnlyRunLog) Len() int {
	return len(self.entries)
}
