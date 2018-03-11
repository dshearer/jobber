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

type JobOutputHandler interface {
	WriteOutput(output []byte, jobName string, runTime time.Time)
	fmt.Stringer
}

type NopJobOutputHandler struct{}

func (self NopJobOutputHandler) String() string {
	return "N/A"
}

func (self NopJobOutputHandler) WriteOutput(output []byte, jobName string,
	runTime time.Time) {
}

type FileJobOutputHandler struct {
	Where      string
	MaxAgeDays int
	Suffix     string
}

func (self FileJobOutputHandler) String() string {
	return fmt.Sprintf(
		"{where: %v, maxAgeDays: %v}",
		self.Where,
		self.MaxAgeDays,
	)
}

func (self FileJobOutputHandler) WriteOutput(output []byte, jobName string,
	runTime time.Time) {

	// make sure dir for job exists
	dirPath := filepath.Join(self.Where, jobName)
	if err := os.Mkdir(dirPath, 0700); err != nil && !os.IsExist(err) {
		common.ErrLogger.Println(err.Error())
		return
	}

	// write output
	path := filepath.Join(dirPath, self.runTimeToFileName(runTime))
	if err := ioutil.WriteFile(path, output, 0600); err != nil {
		common.ErrLogger.Println(err.Error())
	}

	// clean up
	self.deleteOldOutputs(dirPath)
}

func (self FileJobOutputHandler) runTimeToFileName(t time.Time) string {
	return fmt.Sprintf("%v.%v", t.Unix(), self.Suffix)
}

func (self FileJobOutputHandler) fileNameToRunTime(name string) (time.Time, error) {
	retErr := common.Error{What: "Cannot parse output file name"}
	parts := strings.Split(name, ".")
	if len(parts) != 2 || parts[1] != self.Suffix {
		return time.Time{}, &retErr
	}
	secs, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, &retErr
	}
	return time.Unix(int64(secs), 0), nil
}

func (self FileJobOutputHandler) ageOfFileInDays(name string, now time.Time) (int, error) {
	runTime, err := self.fileNameToRunTime(name)
	if err != nil {
		return 0, err
	}

	day := time.Duration(24) * time.Hour
	age := now.Sub(runTime)
	return int(age / day), nil
}

func (self FileJobOutputHandler) deleteOldOutputs(dirPath string) {
	// read directory
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		if !os.IsNotExist(err) {
			common.ErrLogger.Printf("%v", err)
		}
		return
	}

	// iterate over output files
	now := time.Now()
	for _, file := range files {
		if !file.Mode().IsRegular() {
			continue
		}
		ageDays, err := self.ageOfFileInDays(file.Name(), now)
		if err != nil {
			common.ErrLogger.Println(err.Error())
			continue
		}
		if ageDays > self.MaxAgeDays {
			err = os.Remove(filepath.Join(dirPath, file.Name()))
			if err != nil {
				common.ErrLogger.Println(err.Error())
			}
		}
	}
}
