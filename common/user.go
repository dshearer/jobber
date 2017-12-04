package common

import (
	"os"
	"os/user"
	"strconv"
	"syscall"
)

func UserOwnsFile(usr *user.User, path string) (bool, error) {
	// convert UID to int
	userUid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		return false, err
	}

	// get file's owner
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	fileUid := stat.Sys().(*syscall.Stat_t).Uid

	return uint32(userUid) == fileUid, nil
}

func Chown(path string, usr *user.User) error {
	// convert UID & GID to int
	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(usr.Gid)
	if err != nil {
		return err
	}

	// chown
	if err := os.Chown(path, uid, gid); err != nil {
		return err
	}

	return nil
}
