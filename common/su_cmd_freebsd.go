package common

import (
	"os/exec"
)

func su_cmd(user string, cmdStr string, shell string) *exec.Cmd {
	var cmd *exec.Cmd = exec.Command(
		"su",
		"-l", // login shell
		user,
		"-c", cmdStr,
	)
	return cmd
}
