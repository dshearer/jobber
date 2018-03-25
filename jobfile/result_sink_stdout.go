package jobfile

import "os"

const _STDOUT_RESULT_SINK_NAME = "stdout"

/*
This result sink sends run results to jobberrunner's stdout.
*/
type StdoutResultSink struct {
	Data ResultSinkDataParam `yaml:"data"`
}

func (self StdoutResultSink) CheckParams() error {
	return nil
}

func (self StdoutResultSink) String() string {
	return _STDOUT_RESULT_SINK_NAME
}

func (self StdoutResultSink) Equals(other ResultSink) bool {
	otherStdout, ok := other.(StdoutResultSink)
	if !ok {
		return false
	}
	if otherStdout.Data != self.Data {
		return false
	}
	return true
}

func (self StdoutResultSink) Handle(rec RunRec) {
	os.Stdout.Write(SerializeRunRec(rec, self.Data))
}
