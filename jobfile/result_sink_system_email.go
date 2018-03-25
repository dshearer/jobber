package jobfile

import (
	"fmt"
	"os/exec"

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
	cmd := exec.Command("sendmail", rec.Job.User)
	execResult, err := common.ExecAndWait(cmd, &msgBytes)
	if err != nil {
		common.ErrLogger.Printf("Failed to send mail: %v\n", err)
	} else if !execResult.Succeeded {
		errMsg, _ := SafeBytesToStr(execResult.Stderr)
		common.ErrLogger.Printf("Failed to send mail: %v\n", errMsg)
	}
}
