package jobfile

import (
	"fmt"
	"strings"

	"github.com/dshearer/jobber/common"
	"gopkg.in/yaml.v2"
)

/*
A result sink is an object that does something with the results of a job run.
*/
type ResultSink interface {
	Handle(runRec RunRec)
	Validate() error
	Equals(other ResultSink) bool
	fmt.Stringer
}

type ResultSinkRaw map[string]interface{}

func MakeResultSinkFromConfig(config ResultSinkRaw) (ResultSink, error) {
	// get type
	typeName, ok := config["type"]
	if !ok {
		msg := fmt.Sprintf("Missing type for result sink: %v", config)
		return nil, &common.Error{What: msg}
	}

	// extract params
	params := make(ResultSinkRaw)
	for key, value := range config {
		if key == "type" {
			continue
		}
		params[key] = value
	}

	// make sink
	switch typeName {
	case _SYSTEM_EMAIL_RESULT_SINK_NAME:
		var sink SystemEmailResultSink
		if err := loadSinkParams(params, &sink); err != nil {
			return nil, err
		}
		return sink, nil

	case _PROGRAM_RESULT_SINK_NAME:
		var sink ProgramResultSink
		if err := loadSinkParams(params, &sink); err != nil {
			return nil, err
		}
		return sink, nil

	case _FILESYSTEM_RESULT_SINK_NAME:
		var sink FilesystemResultSink
		if err := loadSinkParams(params, &sink); err != nil {
			return nil, err
		}
		return sink, nil

	default:
		msg := fmt.Sprintf("No such result sink type: %v", typeName)
		return nil, &common.Error{What: msg}
	}
}

func loadSinkParams(params map[string]interface{}, sink ResultSink) error {
	paramYaml, err := yaml.Marshal(params)
	if err != nil {
		return err
	}
	if err := yaml.UnmarshalStrict(paramYaml, sink); err != nil {
		return err
	}
	if err := sink.Validate(); err != nil {
		return err
	}
	return nil
}

type ResultSinkDataParam uint

const (
	RESULT_SINK_DATA_STDOUT ResultSinkDataParam = 1 << iota
	RESULT_SINK_DATA_STDERR
)

func (self ResultSinkDataParam) Contains(value ResultSinkDataParam) bool {
	return self&value == value
}

func (self *ResultSinkDataParam) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var strs []string
	if err := unmarshal(&strs); err != nil {
		return err
	}

	*self = 0
	for _, s := range strs {
		s = strings.ToLower(s)
		switch s {
		case "stdout":
			*self |= RESULT_SINK_DATA_STDOUT
		case "stderr":
			*self |= RESULT_SINK_DATA_STDERR
		default:
			msg := fmt.Sprintf("Invalid data value: \"%v\"", s)
			return &common.Error{What: msg}
		}
	}

	return nil
}
