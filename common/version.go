package common

var jobberVersion string

func ShortVersionStr() string {
    return jobberVersion
}

func LongVersionStr() string {
    return ShortVersionStr()
}
