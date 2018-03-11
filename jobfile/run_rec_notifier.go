package jobfile

import (
	"encoding/json"
	"fmt"

	"os/exec"

	"github.com/dshearer/jobber/common"
)

type RunRecNotifier interface {
	Notify(rec *RunRec)
	fmt.Stringer
}

type NopRunRecNotifier struct{}

func (self NopRunRecNotifier) Notify(rec *RunRec) {}

func (self NopRunRecNotifier) String() string {
	return "N/A"
}

type MailRunRecNotifier struct{}

func (self MailRunRecNotifier) Notify(rec *RunRec) {
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

func (self MailRunRecNotifier) String() string {
	return "mail"
}

type ProgramRunRecNotifier struct {
	Program string
}

func (self ProgramRunRecNotifier) Notify(rec *RunRec) {
	/*
	 Here we make a JSON document with the data in rec, and then pass it
	 to a user-specified program.
	*/

	var timeFormat string = "Jan _2 15:04:05 2006"

	// make job JSON
	jobJson := map[string]interface{}{
		"name":    rec.Job.Name,
		"command": rec.Job.Cmd,
		"time":    rec.Job.FullTimeSpec.String(),
		"status":  rec.NewStatus.String()}

	// make rec JSON
	recJson := map[string]interface{}{
		"job":       jobJson,
		"user":      rec.Job.User,
		"startTime": rec.RunTime.Format(timeFormat),
		"succeeded": rec.Succeeded}
	if rec.Stdout == nil {
		recJson["stdout"] = nil
	} else {
		stdoutStr, stdoutBase64 := SafeBytesToStr(*rec.Stdout)
		recJson["stdout"] = stdoutStr
		recJson["stdout_base64"] = stdoutBase64
	}
	if rec.Stderr == nil {
		recJson["stderr"] = nil
	} else {
		stderrStr, stderrBase64 := SafeBytesToStr(*rec.Stderr)
		recJson["stderr"] = stderrStr
		recJson["stderr_base64"] = stderrBase64
	}
	recJsonStr, err := json.Marshal(recJson)
	if err != nil {
		common.ErrLogger.Printf("Failed to make RunRec JSON: %v\n", err)
		return
	}

	// call program
	execResult, err2 := common.ExecAndWait(exec.Command(self.Program),
		&recJsonStr)
	if err2 != nil {
		common.ErrLogger.Printf("Failed to call %v: %v\n", self.Program, err2)
	} else if !execResult.Succeeded {
		errMsg, _ := SafeBytesToStr(execResult.Stderr)
		common.ErrLogger.Printf(
			"%v failed: %v\n",
			self.Program,
			errMsg,
		)
	}
}

func (self ProgramRunRecNotifier) String() string {
	return self.Program
}
