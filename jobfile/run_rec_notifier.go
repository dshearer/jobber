package jobfile

import (
	"encoding/json"
	"fmt"
	"github.com/dshearer/jobber/common"
)

type RunRecNotifier func(rec *RunRec)

func MakeMailNotifier() RunRecNotifier {
	return func(rec *RunRec) {
		headers := fmt.Sprintf("To: %v\r\nFrom: %v\r\nSubject: \"%v\" failed.",
			rec.Job.User,
			rec.Job.User,
			rec.Job.Name)
		body := rec.Describe()
		msg := fmt.Sprintf("%s\r\n\r\n%s.\r\n", headers, body)
		sendmailCmd := fmt.Sprintf("sendmail %v", rec.Job.User)

		// run sendmail
		msgBytes := []byte(msg)
		sudoResult, err := common.SudoAndWait(rec.Job.User, sendmailCmd, "/bin/sh", &msgBytes)
		if err != nil {
			common.ErrLogger.Printf("Failed to send mail: %v\n", err)
		} else if !sudoResult.Succeeded {
			errMsg, _ := common.SafeBytesToStr(sudoResult.Stderr)
			common.ErrLogger.Printf("Failed to send mail: %v\n", errMsg)
		}
	}
}

func MakeProgramNotifier(program string) RunRecNotifier {
	return func(rec *RunRec) {
		/*
		   Here we make a JSON document with the data in rec, and then pass it
		   to a user-specified program.
		*/

		var timeFormat string = "Jan _2 15:04:05 2006"

		// make job JSON
		jobJson := map[string]interface{}{
			"name":            rec.Job.Name,
			"command":         rec.Job.Cmd,
			"time":            rec.Job.FullTimeSpec.String(),
			"onError":         rec.Job.ErrorHandler.String(),
			"notifyOnError":   rec.Job.NotifyOnError,
			"notifyOnFailure": rec.Job.NotifyOnFailure,
			"status":          rec.NewStatus.String()}

		// make rec JSON
		recJson := map[string]interface{}{
			"job":       jobJson,
			"user":      rec.Job.User,
			"startTime": rec.RunTime.Format(timeFormat),
			"succeeded": rec.Succeeded}
		if rec.Stdout == nil {
			recJson["stdout"] = nil
		} else {
			stdoutStr, stdoutBase64 := common.SafeBytesToStr(*rec.Stdout)
			recJson["stdout"] = stdoutStr
			recJson["stdout_base64"] = stdoutBase64
		}
		if rec.Stderr == nil {
			recJson["stderr"] = nil
		} else {
			stderrStr, stderrBase64 := common.SafeBytesToStr(*rec.Stderr)
			recJson["stderr"] = stderrStr
			recJson["stderr_base64"] = stderrBase64
		}
		recJsonStr, err := json.Marshal(recJson)
		if err != nil {
			common.ErrLogger.Printf("Failed to make RunRec JSON: %v\n", err)
			return
		}

		// call program
		sudoResult, err2 := common.SudoAndWait(rec.Job.User, program, "/bin/sh", &recJsonStr)
		if err2 != nil {
			common.ErrLogger.Printf("Failed to call %v: %v\n", program, err2)
		} else if !sudoResult.Succeeded {
			errMsg, _ := common.SafeBytesToStr(sudoResult.Stderr)
			common.ErrLogger.Printf("%v failed: %v\n",
				program,
				errMsg)
		}
	}
}
