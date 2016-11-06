package common

import (
	"os/exec"
)

func sudo_cmd(user string, cmdStr string, shell string) *exec.Cmd {
	var cmd *exec.Cmd = exec.Command(
		"su",
		"-l", // login shell
		"-s", shell,
		"-c", cmdStr,
		user,
	)
	return cmd
}
