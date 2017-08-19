package common

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

const (
	VarDirPath = "/var/jobber"
)

var libexecPaths []string = []string{
	"/usr/libexec",
	"/usr/local/libexec",
}

func PerUserDirPath(usr *user.User) string {
	return filepath.Join(VarDirPath, usr.Uid)
}

func SocketPath(usr *user.User) string {
	return filepath.Join(PerUserDirPath(usr), "socket")
}

func RunnerPidFilePath(usr *user.User) string {
	return filepath.Join(PerUserDirPath(usr), "runner_pid")
}

func FindLibexecProgram(name string) (string, *Error) {
	for _, dir := range libexecPaths {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", &Error{
		fmt.Sprintf("Failed to find %v in %v.", name, libexecPaths),
		nil,
	}
}
