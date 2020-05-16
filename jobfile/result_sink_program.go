package jobfile

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/dshearer/jobber/common"
)

const _PROGRAM_RESULT_SINK_NAME = "program"

type ProgramResultSink struct {
	Path                string `yaml:"path"`
	RunRecFormatVersion SemVer `yaml:"runRecFormatVersion"`
}

func (self ProgramResultSink) CheckParams() error {
	if len(self.Path) == 0 {
		return &common.Error{What: "Program result sink needs 'path' param"}
	}
	if self.RunRecFormatVersion.IsZero() {
		self.RunRecFormatVersion = SemVer{Major: 1, Minor: 4}
	}
	return nil
}

func (self ProgramResultSink) String() string {
	return _PROGRAM_RESULT_SINK_NAME
}

func (self ProgramResultSink) Equals(other ResultSink) bool {
	otherProgResultSink, ok := other.(ProgramResultSink)
	if !ok {
		return false
	}
	if otherProgResultSink.Path != self.Path {
		return false
	}
	return true
}

func serializeRunRec_oldFormat(rec RunRec) []byte {
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
		stdoutStr, stdoutBase64 := SafeBytesToStr(rec.Stdout)
		recJson["stdout"] = stdoutStr
		recJson["stdout_base64"] = stdoutBase64
	}
	if rec.Stderr == nil {
		recJson["stderr"] = nil
	} else {
		stderrStr, stderrBase64 := SafeBytesToStr(rec.Stderr)
		recJson["stderr"] = stderrStr
		recJson["stderr_base64"] = stderrBase64
	}
	recJsonStr, err := json.Marshal(recJson)
	if err != nil {
		panic(fmt.Sprintf("Failed to make RunRec JSON: %v\n", err))
	}
	return recJsonStr
}

func (self ProgramResultSink) Handle(rec RunRec) {
	/*
	 Here we make a JSON document with the data in rec, and then pass it
	 to a user-specified program.
	*/

	var recStr []byte
	if self.RunRecFormatVersion.Compare(SemVer{Major: 1, Minor: 4}) < 0 {
		recStr = serializeRunRec_oldFormat(rec)
	} else {
		recStr = SerializeRunRec(rec, RESULT_SINK_DATA_STDOUT|RESULT_SINK_DATA_STDERR)
	}

	// call program
	execResult, err2 := common.ExecAndWait(exec.Command(self.Path), recStr)
	defer execResult.Close()
	if err2 != nil {
		common.ErrLogger.Printf("Failed to call %v: %v\n", self.Path, err2)
	} else if !execResult.Succeeded {
		stderrBytes, _ := execResult.ReadStderr(RunRecOutputMaxLen)
		errMsg, _ := SafeBytesToStr(stderrBytes)
		common.ErrLogger.Printf(
			"%v failed: %v\n",
			self.Path,
			errMsg,
		)
	} else {
		stdoutBytes, _ := execResult.ReadStdout(RunRecOutputMaxLen)
		stderrBytes, _ := execResult.ReadStderr(RunRecOutputMaxLen)
		stdout, _ := SafeBytesToStr(stdoutBytes)
		stderr, _ := SafeBytesToStr(stderrBytes)
		common.Logger.Print(stdout)
		common.ErrLogger.Print(stderr)
	}
}
