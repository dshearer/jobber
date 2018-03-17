package jobfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dshearer/jobber/common"
)

const _FILESYSTEM_RESULT_SINK_NAME = "filesystem"
const _FS_SINK_STDOUT_SUFFIX = "stdout"
const _FS_SINK_STDERR_SUFFIX = "stderr"

/*
This result sink writes a run's stdout or stderr to disk.

Example: Consider FilesystemResultSink{Path: "/some/dir", DataRaw: ["stdout", "stderr"]}.
For a job named "JobOne", this will result in a set of files like this:

  - /some/dir/
    - JobOne/
      - 1521318351.stdout
      - 1521318351.stderr
      - 1521318411.stdout
      - 1521318411.stderr
      - 1521318471.stdout
      - 1521318471.stderr
*/
type FilesystemResultSink struct {
	Path       string              `yaml:"path"`
	Data       ResultSinkDataParam `yaml:"data"`
	MaxAgeDays int                 `yaml:"maxAgeDays"`
}

func (self FilesystemResultSink) String() string {
	return _FILESYSTEM_RESULT_SINK_NAME
}

func (self FilesystemResultSink) Equals(other ResultSink) bool {
	otherResultSink, ok := other.(FilesystemResultSink)
	if !ok {
		return false
	}
	if otherResultSink.Path != self.Path {
		return false
	}
	if otherResultSink.MaxAgeDays != self.MaxAgeDays {
		return false
	}
	if otherResultSink.Data != self.Data {
		return false
	}
	return true
}

func (self FilesystemResultSink) Validate() error {
	if len(self.Path) == 0 {
		return &common.Error{What: "Filesystem result sink needs 'path' param"}
	}
	if self.MaxAgeDays < 1 {
		msg := "Filesystem result sink's 'maxAgeDays' param must be >= 1"
		return &common.Error{What: msg}
	}
	return nil
}

func (self FilesystemResultSink) Handle(rec RunRec) {
	// make sure dir for job exists
	dirPath := filepath.Join(self.Path, rec.Job.Name)
	if err := os.Mkdir(dirPath, 0700); err != nil && !os.IsExist(err) {
		common.ErrLogger.Println(err.Error())
		return
	}

	// write output
	if self.Data.Contains(RESULT_SINK_DATA_STDOUT) {
		fileName := runTimeToFileName(rec.RunTime, _FS_SINK_STDOUT_SUFFIX)
		path := filepath.Join(dirPath, fileName)
		if err := ioutil.WriteFile(path, rec.Stdout, 0600); err != nil {
			common.ErrLogger.Println(err.Error())
		}
	}
	if self.Data.Contains(RESULT_SINK_DATA_STDERR) {
		fileName := runTimeToFileName(rec.RunTime, _FS_SINK_STDERR_SUFFIX)
		path := filepath.Join(dirPath, fileName)
		if err := ioutil.WriteFile(path, rec.Stderr, 0600); err != nil {
			common.ErrLogger.Println(err.Error())
		}
	}

	// clean up
	deleteOldOutputs(dirPath, self.MaxAgeDays)
}

func runTimeToFileName(t time.Time, suffix string) string {
	return fmt.Sprintf("%v.%v", t.Unix(), suffix)
}

func fileNameToRunTime(name string) (time.Time, error) {
	retErr := common.Error{What: "Cannot parse output file name"}
	parts := strings.Split(name, ".")
	if len(parts) != 2 ||
		(parts[1] != _FS_SINK_STDOUT_SUFFIX && parts[1] != _FS_SINK_STDERR_SUFFIX) {
		return time.Time{}, &retErr
	}
	secs, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, &retErr
	}
	return time.Unix(int64(secs), 0), nil
}

func ageOfFileInDays(name string, now time.Time) (int, error) {
	runTime, err := fileNameToRunTime(name)
	if err != nil {
		return 0, err
	}

	day := time.Duration(24) * time.Hour
	age := now.Sub(runTime)
	return int(age / day), nil
}

func deleteOldOutputs(dirPath string, maxAgeDays int) {
	// read directory
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		if !os.IsNotExist(err) {
			common.ErrLogger.Println(err.Error())
		}
		return
	}

	// iterate over output files
	now := time.Now()
	for _, file := range files {
		if !file.Mode().IsRegular() {
			continue
		}
		ageDays, err := ageOfFileInDays(file.Name(), now)
		if err != nil {
			common.ErrLogger.Println(err.Error())
			continue
		}
		if ageDays > int(maxAgeDays) {
			if err := os.Remove(filepath.Join(dirPath, file.Name())); err != nil {
				common.ErrLogger.Println(err.Error())
			}
		}
	}
}
