package main

import (
	"os/user"
	"path/filepath"

	"github.com/dshearer/jobber/common"
)

func userHasHome(usr *user.User) bool {
	// check if user has home dir
	if len(usr.HomeDir) == 0 || usr.HomeDir == "/dev/null" {
		common.Logger.Printf("Excluding %v: has no home directory",
			usr.Username)
		return false

	}

	// check if home dir path is absolute
	if !filepath.IsAbs(usr.HomeDir) {
		common.Logger.Printf("Excluding %v: home directory path is "+
			"not absolute: %v", usr.Username, usr.HomeDir)
		return false
	}

	// check if user owns home dir
	ownsHomeDir, err := common.UserOwnsFile(usr, usr.HomeDir)
	if err != nil {
		common.Logger.Printf("Excluding %v: %v", usr.Username, err)
		return false
	}
	if !ownsHomeDir {
		common.Logger.Printf("Excluding %v: doesn't own home dir",
			usr.Username)
		return false
	}

	return true
}

func shouldRunForUser(usr *user.User) bool {
	// check prefs
	if !common.JobberShouldRunForUser(usr) {
		common.Logger.Printf("Excluding %v according to prefs",
			usr.Username)
		return false
	}

	if !userHasHome(usr) {
		return false
	}

	if !shouldRunForUser_platform(usr) {
		return false
	}

	return true
}

/*
Get all users for which we should run jobberrunner.
*/
func getAcceptableUsers() ([]*user.User, error) {
	acceptableUsers := make([]*user.User, 0)
	allUsers, err := getAllUsers()
	if err != nil {
		return nil, err
	}
	for _, usr := range allUsers {
		// check for reasons to exclude
		if !shouldRunForUser(usr) {
			continue
		}

		acceptableUsers = append(acceptableUsers, usr)
	}

	return acceptableUsers, nil
}
