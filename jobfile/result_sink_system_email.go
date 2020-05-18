package jobfile

import (
	"fmt"

	"github.com/dshearer/jobber/common"
)

const _SYSTEM_EMAIL_RESULT_SINK_NAME = "system-email"

type SystemEmailResultSink struct{}

func (self SystemEmailResultSink) CheckParams() error {
	return nil
}

func (self SystemEmailResultSink) String() string {
	return _SYSTEM_EMAIL_RESULT_SINK_NAME
}

func (self SystemEmailResultSink) Equals(other ResultSink) bool {
	_, ok := other.(SystemEmailResultSink)
	return ok
}

func (self SystemEmailResultSink) Handle(rec RunRec) {
	headers := fmt.Sprintf("To: %v\r\nFrom: %v\r\nSubject: \"%v\" failed.",
		rec.Job.User,
		rec.Job.User,
		rec.Job.Name)
	body := rec.Describe()
	msg := fmt.Sprintf("%s\r\n\r\n%s.\r\n", headers, body)

	// run sendmail
	msgBytes := []byte(msg)
	execResult, err := common.ExecAndWait([]string{"sendmail", rec.Job.User}, msgBytes)
	defer execResult.Close()
	if err != nil {
		common.ErrLogger.Printf("Failed to send mail: %v\n", err)
	} else if execResult.Fate == common.SubprocFateFailed {
		stdoutBytes, _ := execResult.ReadStderr(RunRecOutputMaxLen)
		errMsg, _ := SafeBytesToStr(stdoutBytes)
		common.ErrLogger.Printf("Failed to send mail: %v\n", errMsg)
	} else if execResult.Fate == common.SubprocFateCancelled {
		panic("Result sink program subproc was somehow cancelled")
	}
}
