package jobfile

import (
	"fmt"
	"time"
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

}
