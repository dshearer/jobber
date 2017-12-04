package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

const (
	VarDirPath         = "/var/jobber"
	CmdSocketFileName  = "cmd.sock"
	QuitSocketFileName = "quit.sock"
)

var libexecPaths []string = []string{
	"/usr/libexec",
	"/usr/local/libexec",
}

func PerUserDirPath(usr *user.User) string {
	return filepath.Join(VarDirPath, usr.Uid)
}

func CmdSocketPath(usr *user.User) string {
	return filepath.Join(PerUserDirPath(usr), CmdSocketFileName)
}

func QuitSocketPath(usr *user.User) string {
	return filepath.Join(PerUserDirPath(usr), QuitSocketFileName)
}

/*
 Get a list of all users for whom there is a jobberrunner process.
*/
func AllUsersWithSockets() ([]*user.User, error) {
	// get list of per-user dirs
	files, err := ioutil.ReadDir(VarDirPath)
	if err != nil {
		return nil, err
	}

	// make list of users
	users := make([]*user.User, 0)
	for _, file := range files {
		if file.IsDir() {
			usr, err := user.LookupId(file.Name())
			if err != nil {
				continue
			}
			users = append(users, usr)
		}
	}
	return users, nil
}

func RunnerPidFilePath(usr *user.User) string {
	return filepath.Join(PerUserDirPath(usr), "runner_pid")
}

func FindLibexecProgram(name string) (string, error) {
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
