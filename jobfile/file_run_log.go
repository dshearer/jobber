package jobfile

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dshearer/jobber/common"
)

/*
This is an implementation of RunLog that stores entries on disk.
*/
type fileRunLog struct {
	/*
		    Common operations:
		 	- append entry (ascending order)
			- get recent entries

			It is important to guarantee that the log reflects all jobs
			that have run, so we must write entries to disk as soon as we
			get them.

			It is also important to avoid using too much memory.

			Speed is not as important, for both Get* and Put*
			operations.

			As mentioned, all entries are written to a file.  When the file
			reaches a max size, it is rotated out and a new, empty file
			starts being used.
	*/

	filePath     string
	maxFileLen   int64
	maxHistories int // max number of historical files

	index []backingFileDtor
}

func (self *fileRunLog) String() string {
	return fmt.Sprintf(
		"FileRunLog{filePath: %v, maxFileLen: %v, maxHistories: %v}",
		self.filePath,
		self.maxFileLen,
		self.maxHistories,
	)
}

/*
A backing file consists of one or more gLogEntryLen-byte entries separated by
'\n'.

m = number of entries in current backing file
s = smallest entry index in a backing file
n = max entries per backing file
  = floor(maxFileLen / (gLogEntryLen+1)) +
      floor((maxFileLen % (gLogEntryLen+1))/gLogEntryLen)

Backing file 0     Backing file 1     Backing file 2      Backing file k
+-------------+    +-------------+    +-------------+     +-------------+
|entry m-1    |    |entry s+n-1  |    |entry s+n-1  |     |entry s+n-1  |
|entry m-2    |    |entry s+n-2  |    |entry s+n-2  |     |entry s+n-2  |
|.            |    |.            |    |.            | ... |.            |
|.            |    |.            |    |.            |     |.            |
|.            |    |.            |    |.            |     |.            |
|entry 1      |    |entry s+1    |    |entry s+1    |     |entry s+1    |
|entry 0      |    |entry s      |    |entry s      |     |entry s      |
+-------------+    +-------------+    +-------------+     +-------------+
   s = 0              s = m              s = m+n             s = m+(k-1)n

                                                                                                                                /gLogEntryLen)
 Rel Index   File offset            Contents
+-----------------------------------------------+
 n-1         0                       entry + '\n'
 n-2         gLogEntryLen+1          entry + '\n'
 n-3         2(gLogEntryLen+1)       entry + '\n'
 .           .                       .
 .           .                       .
 i           (n-i-1)(gLogEntryLen+1) entry + '\n'
 .           .                       .
 1           (n-2)(gLogEntryLen+1)   entry + '\n'
 0           (n-1)(gLogEntryLen+1)   entry

*/

const (
	gMaxJobNameLen int64 = 16
	gLogEntryLen   int64 = 64
)

type backingFileDtor struct {
	path         string
	nbrEntries   int
	earliestTime time.Time // of 1st entry
	latestTime   time.Time // of last entry
	startIdx     int       // of last entry
}

/*
Get offset of entry at given relative index.  Remember that the last entry
has index 0.  Panics if idx is not a valid relative index.
*/
func (self *backingFileDtor) offsetOfEntry(idx int) int64 {
	if idx >= self.nbrEntries {
		panic("Invalid entry index")
	}
	return int64(self.nbrEntries-idx-1) * (gLogEntryLen + 1)
}

/*
Read the bytes of the entry at the given relative index.  Panics if there is
no such entry.
*/
func (self *backingFileDtor) readEntryBytes(idx int,
	f *os.File) ([]byte, error) {

	offset := self.offsetOfEntry(idx)
	buf := make([]byte, gLogEntryLen)
	if _, err := f.ReadAt(buf, offset); err != nil {
		msg := fmt.Sprintf("Failed to read backing file %v", self.path)
		return nil, &common.Error{What: msg, Cause: err}
	}
	return buf, nil
}

/*
Read entry at the given relative index.
*/
func (self *backingFileDtor) readEntry(idx int,
	f *os.File) (*RunLogEntry, error) {

	buf, err := self.readEntryBytes(idx, f)
	if err != nil {
		msg := fmt.Sprintf("Failed to read %v", self.path)
		return nil, &common.Error{What: msg, Cause: err}
	}
	entry, err := decodeRunLogEntry(string(buf))
	if err != nil {
		msg := fmt.Sprintf("Entry %v in %v is bad", idx, self.path)
		return nil, &common.Error{What: msg, Cause: err}
	}
	return entry, nil
}

/*
Get the relative index i of the entry with the lowest index whose
timestamp is not after the given timestamp.

If there is no such entry, returns self.nbrEntries.
*/
func (self *backingFileDtor) firstEntryNotAfter(t time.Time,
	f *os.File) (int, error) {

	var searchErr error = nil
	pred := func(i int) bool {
		entry, searchErr := self.readEntry(i, f)
		if searchErr != nil {
			return false
		}
		return !entry.Time.After(t)
	}
	idx := sort.Search(self.nbrEntries, pred)
	if searchErr != nil {
		return self.nbrEntries, searchErr
	}
	return idx, nil
}

/*
Append an entry to the backing file.  NOTE: Entry's timestamp must not
be earlier than any other entry's.
*/
func (self *backingFileDtor) pushEntry(entry *RunLogEntry) error {
	// open backing file
	f, err := os.OpenFile(
		self.path,
		os.O_APPEND|os.O_WRONLY,
		0600,
	)
	if err != nil {
		return err
	}
	defer f.Close()

	// write entry
	return self.pushEntryF(entry, f)
}

/*
Append an entry to the backing file.  NOTE: Entry's timestamp must not
be earlier than any other entry's.
*/
func (self *backingFileDtor) pushEntryF(entry *RunLogEntry,
	f *os.File) error {

	// update timestamps
	if self.nbrEntries == 0 {
		self.earliestTime = entry.Time
		self.latestTime = entry.Time
	} else {
		self.latestTime = entry.Time
	}

	// update nbrEntries
	self.nbrEntries++

	// serialize entry
	entryStr := encodeRunLogEntry(entry)
	if self.nbrEntries > 1 {
		entryStr = "\n" + entryStr
	}

	// write to backing file
	if _, err := f.WriteString(entryStr); err != nil {
		return err
	}

	return nil
}

func (self *backingFileDtor) popEntryF(f *os.File) (*RunLogEntry, error) {
	if self.nbrEntries == 0 {
		panic("This backing file is empty!")
	}

	// read latest entry
	entry, err := self.readEntry(0, f)
	if err != nil {
		return nil, err
	}

	// update nbrEntries
	self.nbrEntries--

	// calc new file len
	newFileLen := int64(self.nbrEntries-1)*(gLogEntryLen+1) + gLogEntryLen

	// truncate file
	if err = f.Truncate(newFileLen); err != nil {
		return nil, err
	}

	return entry, nil
}

/*
Return whether this file can take another entry.
*/
func (self *backingFileDtor) isFull(maxFileLen int64) bool {
	// compute new entry's len
	newEntryLen := gLogEntryLen
	if self.nbrEntries > 0 {
		newEntryLen += 1
	}

	// compute backing file's current len
	var fileLen int64 = 0
	if self.nbrEntries == 1 {
		fileLen = gLogEntryLen
	} else if self.nbrEntries > 1 {
		fileLen = int64(self.nbrEntries-1)*(gLogEntryLen+1) +
			gLogEntryLen
	}

	return fileLen+newEntryLen > maxFileLen
}

/*
Returns nil if there is no such file.
*/
func makeBackingFileDtor(path string, startIdx int) (*backingFileDtor, error) {
	dtor := backingFileDtor{path: path, nbrEntries: 0, startIdx: startIdx}

	// open file
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	defer f.Close()

	// stat it
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	// count entries
	if stat.Size() == 0 {
		dtor.nbrEntries = 0
	} else if stat.Size() < gLogEntryLen {
		msg := fmt.Sprintf(
			"Invalid log file: %v: size is less than entry size",
			path,
		)
		return nil, &common.Error{What: msg}
	} else {
		remainder := stat.Size() - gLogEntryLen
		if remainder%(gLogEntryLen+1) != 0 {
			msg := fmt.Sprintf(
				"Invalid log file: %v: size is not multiple of entry size",
				path,
			)
			return nil, &common.Error{What: msg}
		} else {
			dtor.nbrEntries = int(remainder/(gLogEntryLen+1) + 1)
		}
	}

	if dtor.nbrEntries > 0 {
		// get earliest time
		entry, err := dtor.readEntry(dtor.nbrEntries-1, f)
		if err != nil {
			return nil, err
		}
		dtor.earliestTime = entry.Time

		// get latest time
		entry, err = dtor.readEntry(0, f)
		if err != nil {
			return nil, err
		}
		dtor.latestTime = entry.Time
	}

	return &dtor, nil
}

/*
Makes an index of all the backing files.

POSTCONDITION: self.index has exactly one descriptor for the current
backing file and one descriptor for each historical backing file.
If there was no current backing file, then self.index contains exactly
one backing file descriptor (with nbrEntries == 0) and there is now
one (empty) current backing file.
*/
func (self *fileRunLog) makeIndex() error {
	self.index = nil

	// look at current backing file
	dtor, err := makeBackingFileDtor(self.filePath, 0)
	if err != nil {
		return err
	}
	if dtor == nil {
		/* There is no current backing file.  Make one. */
		dtor = &backingFileDtor{path: self.filePath}
		tmp := [0]byte{}
		err := ioutil.WriteFile(self.filePath, tmp[:], 0600)
		if err != nil {
			return err
		}
	}
	self.index = append(self.index, *dtor)

	// look at historical backing files
	for i := 1; i <= self.maxHistories; i++ {
		newStartIdx := self.index[i-1].startIdx +
			self.index[i-1].nbrEntries
		dtor, err = makeBackingFileDtor(
			fmt.Sprintf("%v.%v", self.filePath, i),
			newStartIdx,
		)
		if err != nil {
			return err
		}
		if dtor == nil {
			break
		}
		self.index = append(self.index, *dtor)
	}
	return nil
}

/*
Replace the current backing file with an empty one.  Preserve old
entries according to settings.
*/
func (self *fileRunLog) rotateFiles() error {
	/*
	   Historical backing files are named self.filePath + ".N" where N
	   is a numeral.  Older files have greater N.
	*/

	/*
			We must rename all backing files thus (and in this order):

		        	self.filePath + ".N" => self.filePath + ".N+1"
		        	.
		        	.
		        	.
		        	self.filePath + ".2" => self.filePath + ".3"
		        	self.filePath + ".1" => self.filePath + ".2"
		        	self.filePath => self.filePath + ".1"
	*/

	for oldIdx := len(self.index) - 1; oldIdx >= 0; oldIdx-- {
		newIdx := oldIdx + 1
		if newIdx > self.maxHistories {
			// delete this backing file
			if err := os.Remove(self.index[oldIdx].path); err != nil {
				return err
			}
		} else {
			// rename this backing file
			oldPath := self.index[oldIdx].path
			newPath := fmt.Sprintf("%v.%v", self.filePath, newIdx)
			if err := renameRobust(oldPath, newPath); err != nil {
				return err
			}
		}
	}

	// remake index (and make new current backing file)
	if err := self.makeIndex(); err != nil {
		return err
	}

	return nil
}

type entryIterator struct {
	log         *fileRunLog
	entryRelIdx int // rel idx of next entry
	dtorIdx     int // idx of current dtor
	f           *os.File
}

func (self *entryIterator) done() bool {
	return self.entryRelIdx >= self.log.index[self.dtorIdx].nbrEntries &&
		self.dtorIdx+1 >= len(self.log.index)
}

func (self *entryIterator) next() (*RunLogEntry, error) {
	if self.done() {
		panic("Iterator is done")
	}

	if self.entryRelIdx >= self.log.index[self.dtorIdx].nbrEntries {
		// switch to next historical file
		self.close()
		self.dtorIdx++
		self.entryRelIdx = 0

		return self.next()
	}

	if self.f == nil {
		var err error
		self.f, err = os.Open(self.log.index[self.dtorIdx].path)
		if err != nil {
			return nil, err
		}
	}

	// read entry
	entry, err := self.log.index[self.dtorIdx].readEntry(
		self.entryRelIdx,
		self.f,
	)
	if err != nil {
		return nil, err
	}

	// update state
	self.entryRelIdx++

	return entry, nil
}

func (self *entryIterator) close() {
	if self.f != nil {
		self.f.Close()
		self.f = nil
	}
}

func (self *fileRunLog) iterAt(idx int) entryIterator {
	if idx >= self.Len() {
		panic(fmt.Sprintf("Invalid index: %v", idx))
	}

	iter := entryIterator{log: self, dtorIdx: -1}
	for i := 0; i < len(self.index); i++ {
		dtor := self.index[i]
		if dtor.startIdx <= idx && dtor.startIdx+dtor.nbrEntries > idx {
			iter.dtorIdx = i
			iter.entryRelIdx = idx - dtor.startIdx
			break
		}
	}
	if iter.dtorIdx == -1 {
		panic("Bug!")
	}

	return iter
}

func (self *fileRunLog) GetFromTime(maxTime time.Time,
	timeArr ...time.Time) ([]*RunLogEntry, error) {

	if len(timeArr) > 1 {
		panic("Too many args.")
	}

	if self.Len() == 0 {
		return []*RunLogEntry{}, nil
	}

	var minTime time.Time
	if len(timeArr) >= 1 {
		minTime = timeArr[0]
	} else {
		// set *minTime* to just before the earliest entry's start time
		lastDtor := self.index[len(self.index)-1]
		minTime = lastDtor.earliestTime.Add(-time.Second)
	}

	if maxTime.Before(minTime) {
		panic("maxTime is before minTime")
	}

	// find index of first (latest) entry we should return
	startIdx := -1
	for _, dtor := range self.index {
		if !dtor.earliestTime.After(maxTime) && !dtor.latestTime.Before(maxTime) {
			// open file
			f, err := os.Open(dtor.path)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			// search for first (latest) entry
			relStartIdx, err := dtor.firstEntryNotAfter(maxTime, f)
			if err != nil {
				return nil, err
			}

			startIdx = relStartIdx + dtor.startIdx
			break
		}
	}
	if startIdx == -1 {
		return []*RunLogEntry{}, nil
	}

	// get requested entries
	iter := self.iterAt(startIdx)
	defer iter.close()
	var result []*RunLogEntry
	for {
		if iter.done() {
			break
		}

		entry, err := iter.next()
		if err != nil {
			return nil, err
		}
		if entry.Time.After(minTime) {
			result = append(result, entry)
		} else {
			break
		}
	}
	return result, nil
}

func (self *fileRunLog) GetFromIndex(minIdx int, idxArr ...int) (
	[]*RunLogEntry, error) {

	if len(idxArr) > 1 {
		panic("Too many args.")
	}

	var maxIdx int
	if len(idxArr) >= 1 {
		maxIdx = idxArr[0]
	} else {
		maxIdx = self.Len()
	}

	if minIdx > maxIdx {
		panic("from > to")
	}
	if minIdx >= self.Len() {
		panic(fmt.Sprintf("Invalid 'minIdx' index: %v", minIdx))
	}
	if maxIdx > self.Len() {
		panic(fmt.Sprintf("Invalid 'maxIdx' index: %v", maxIdx))
	}

	// get requested entries
	iter := self.iterAt(minIdx)
	result := make([]*RunLogEntry, 0, maxIdx-minIdx)
	for len(result) < maxIdx-minIdx {
		entry, err := iter.next()
		if err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	return result, nil
}

func (self *fileRunLog) GetAll() ([]*RunLogEntry, error) {
	if self.Len() == 0 {
		return nil, nil
	}

	var result []*RunLogEntry
	iter := self.iterAt(0)
	defer iter.close()
	for !iter.done() {
		entry, err := iter.next()
		if err != nil {
			return nil, err
		}
		result = append(result, entry)
	}
	return result, nil
}

func (self *fileRunLog) Len() int {
	n := 0
	for _, dtor := range self.index {
		n += dtor.nbrEntries
	}
	return n
}

func (self *fileRunLog) Put(entry RunLogEntry) error {
	if len(self.index) == 0 {
		panic("Index is not made")
	}

	if self.index[0].nbrEntries > 0 &&
		self.index[0].earliestTime.After(entry.Time) {
		return &common.Error{What: "Entry's timestamp is too early"}
	}

	// check whether this entry is out of order
	var outOfOrder bool = self.index[0].nbrEntries > 0 &&
		self.index[0].latestTime.After(entry.Time)

	if !outOfOrder {
		// rotate, if necessary
		if self.index[0].isFull(self.maxFileLen) {
			if err := self.rotateFiles(); err != nil {
				return err
			}
		}

		// write entry
		if err := self.index[0].pushEntry(&entry); err != nil {
			return err
		}

		// update other entries in index
		for i := 1; i < len(self.index); i++ {
			self.index[i].startIdx++
		}
	} else {
		if err := self.putOutOfOrder(&entry); err != nil {
			return err
		}
	}

	return nil
}

func (self *fileRunLog) putOutOfOrder(entry *RunLogEntry) error {

	/*
	   Suppose the current backing file looks like this:

	    +----------+
	    |entry n-1 |
	    |.         |                                                                                                                    /gLogEntryLen)
	    |.         |
	    |.         |
	    |entry 0   |
	    +----------+

	   The entry that we need to insert must come after some entry i such
	   that n-1 >= i >= 1.  (i != 0 because this method wouldn't have
	   been called in that case.)

	   If mustRotate == false, then the result will be this:

	    +-----------+
	    |entry n-1  |
	    |.          |
	    |.          |                                                                                          /gLogEntryLen)
	    |.          |
	    |entry i    |
	    |new entry  |
	    |.          |
	    |.          |
	    |.          |
	    |entry 0    |
	    +-----------+

	   But if mustRotate == true, then we must rotate the backing files,
	   thus making a new one, and end up with this:

	    Current file      Historical file 1
	    +----------+       +-----------+
	    |entry 0   |       |entry n-1  |                                                                                                    /gLogEntryLen)
	    |          |       |.          |
	    |          |       |.          |
	    |          |       |.          |
	    |          |       |entry i    |
	    |          |       |new entry  |
	    |          |       |.          |
	    |          |       |.          |
	    |          |       |.          |
	    |          |       |entry 1    |
	    +----------+       +-----------+
	*/

	if self.index[0].earliestTime.After(entry.Time) {
		return &common.Error{What: "Entry is too out-of-order"}
	}

	// open backing file
	f, err := os.OpenFile(
		self.index[0].path,
		os.O_RDWR,
		0,
	)
	if err != nil {
		msg := "Failed to open backing file"
		return &common.Error{What: msg, Cause: err}
	}
	defer f.Close()

	// find place to insert entry
	insertIdx, err := self.index[0].firstEntryNotAfter(entry.Time, f)
	if err != nil {
		return err
	}
	insertIdx -= 1

	// make temp file in which to make new backing file
	tmpF, err := ioutil.TempFile("", "newBackingFile")
	if err != nil {
		msg := "Failed to make temp file"
		return &common.Error{What: msg, Cause: err}
	}
	tmpDtor, err := makeBackingFileDtor(tmpF.Name(), 0)
	if err != nil {
		tmpF.Close()
		os.Remove(tmpF.Name())
		return err
	}

	var mustRotate bool = self.index[0].isFull(self.maxFileLen)

	// copy entries to temp file
	lastIdx := 0
	if mustRotate {
		lastIdx = 1
	}
	for i := self.index[0].nbrEntries - 1; i >= lastIdx; i-- {
		// read current entry
		currEntry, err := self.index[0].readEntry(i, f)
		if err != nil {
			tmpF.Close()
			os.Remove(tmpF.Name())
			return err
		}

		// write entries
		if i == insertIdx {
			if err = tmpDtor.pushEntryF(entry, tmpF); err != nil {
				tmpF.Close()
				os.Remove(tmpF.Name())
				return err
			}
		}
		if err = tmpDtor.pushEntryF(currEntry, tmpF); err != nil {
			tmpF.Close()
			os.Remove(tmpF.Name())
			return err
		}
	}
	tmpF.Close()

	var entry0 *RunLogEntry
	if mustRotate {
		// keep last entry of old file
		entry0, err = self.index[0].readEntry(0, f)
		if err != nil {
			os.Remove(tmpF.Name())
			return err
		}
	}

	// replace old file with new
	f.Close()
	if err = renameRobust(tmpDtor.path, self.index[0].path); err != nil {
		os.Remove(tmpDtor.path)
		msg := "Failed to replace old backing file with new"
		return &common.Error{What: msg, Cause: err}
	}
	tmpDtor.path = self.index[0].path
	self.index[0] = *tmpDtor

	if mustRotate {
		if err = self.rotateFiles(); err != nil {
			return err
		}

		if err = self.index[0].pushEntry(entry0); err != nil {
			return err
		}
	}
	return nil
}

/*
Delete the whole log -- i.e., all the backing files.

This is only used in the unit test.
*/
func (self *fileRunLog) deleteAll() {
	for _, dtor := range self.index {
		os.Remove(dtor.path)
	}
	self.makeIndex()
}

func (self *fileRunLog) debugInfo() string {
	// read current backing file
	bytes, err := ioutil.ReadFile(self.index[0].path)
	if err != nil {
		return fmt.Sprintf("Failed to open current backing file: %v",
			err)
	}
	return "Contents of current backing file:\n" + string(bytes)
}

func (self *fileRunLog) FilePath() string {
	return self.filePath
}

func (self *fileRunLog) MaxFileLen() int64 {
	return self.maxFileLen
}

func (self *fileRunLog) MaxHistories() int {
	return self.maxHistories
}

func NewFileRunLog(filePath string, maxFileLen int64,
	maxHistories int) (RunLog, error) {

	log := fileRunLog{
		filePath:     filePath,
		maxFileLen:   maxFileLen,
		maxHistories: maxHistories,
	}
	if err := log.makeIndex(); err != nil {
		return nil, err
	}
	return &log, nil
}

func encodeRunLogEntry(entry *RunLogEntry) string {
	// truncate job name
	n := int(gMaxJobNameLen)
	if len(entry.JobName) < n {
		n = len(entry.JobName)
	}
	jobName := entry.JobName[:n]

	// encode any newlines in the job name
	jobNameParts := strings.Split(jobName, "\n")
	encodedJobName := strings.Join(jobNameParts, "\\n")

	// encode any tabs in the job name
	jobNameParts = strings.Split(encodedJobName, "\t")
	encodedJobName = strings.Join(jobNameParts, "\\t")

	// encode time as Unix Epoch in nanoseconds
	encodedTime := entry.Time.UnixNano()

	// encode whole thing
	tmp := fmt.Sprintf(
		"%v\t%v\t%v\t%v",
		encodedJobName,
		encodedTime,
		entry.Succeeded,
		entry.Result,
	)
	suffix := strings.Repeat(" ", int(gLogEntryLen)-len(tmp))
	return fmt.Sprintf("%v%v", tmp, suffix)
}

func decodeRunLogEntry(s string) (*RunLogEntry, error) {
	var entry RunLogEntry

	// trim trailing whitespace
	s = strings.TrimRight(s, " ")

	// split string into fields
	fields := strings.Split(s, "\t")
	if len(fields) != 4 {
		msg := "Not enough fields in log entry line."
		return nil, &common.Error{What: msg}
	}

	// decode job name
	jobNameParts := strings.Split(fields[0], "\\n")
	entry.JobName = strings.Join(jobNameParts, "\n")
	jobNameParts = strings.Split(entry.JobName, "\\t")
	entry.JobName = strings.Join(jobNameParts, "\t")

	// decode time
	unixTime, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil, err
	}
	entry.Time = time.Unix(0, unixTime)

	// decode succeeded
	if fields[2] == "true" {
		entry.Succeeded = true
	} else if fields[2] == "false" {
		entry.Succeeded = false
	} else {
		return nil, &common.Error{What: "Invalid 'Succeeded' field."}
	}

	// decode result
	decodedStatus := false
	for _, status := range JobStatuses {
		if fields[3] == status.String() {
			entry.Result = status
			decodedStatus = true
			break
		}
	}
	if !decodedStatus {
		return nil, &common.Error{What: "Invalid 'Result' field."}
	}

	return &entry, nil
}

/*
Rename a file, handling cases where acutal renaming isn't possible by
just copying the contents.
*/
func renameRobust(oldpath, newpath string) error {
	if err := os.Rename(oldpath, newpath); err != nil {
		// open old file
		oldF, err := os.Open(oldpath)
		if err != nil {
			return err
		}
		defer oldF.Close()

		// get old file's perms
		oldFileinfo, err := oldF.Stat()
		if err != nil {
			return err
		}

		/*
			NOTE: We don't run as root, so we can't set the owner of
			the new file.
		*/

		// open new file
		newF, err := os.OpenFile(
			newpath,
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
			oldFileinfo.Mode(),
		)
		defer newF.Close()

		// copy contents
		var buffer [1024]byte
		for {
			// read bytes
			n, err := oldF.Read(buffer[:])
			if n == 0 && err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			// write bytes
			n, err = newF.Write(buffer[:n])
			if err != nil {
				return err
			}
		}

		// delete old file
		os.Remove(oldpath)
	}
	return nil
}
