package common

import (
	"os/user"
	"path/filepath"
)

const (
	VarDirPath = "/var/jobber"
)

func PerUserDirPath(usr *user.User) string {
	return filepath.Join(VarDirPath, usr.Uid)
}

func SocketPath(usr *user.User) string {
	return filepath.Join(PerUserDirPath(usr), "socket")
}
