package common

type IpcArg struct {
	User        string
	ForAllUsers bool

	// For TestCmd, CatCmd:
	Job     string
	JobUser string

	// For PauseCmd, ResumeCmd:
	Jobs []string
}
