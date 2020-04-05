package common

import "fmt"

var jobberVersion string

func ShortVersionStr() string {
	return jobberVersion
}

func LongVersionStr() string {
	return fmt.Sprintf("Jobber %s", ShortVersionStr())
}
