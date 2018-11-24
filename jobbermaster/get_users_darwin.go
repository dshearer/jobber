package main

import (
	"os/exec"
	"os/user"
	"strings"

	"github.com/dshearer/jobber/common"
)

/*
Get all users.
*/
func getAllUsers() ([]*user.User, error) {
	users := make([]*user.User, 0)
	out, err := exec.Command("dscl", ".", "list", "/Users").Output()
	if err != nil {
		return nil, err
	}
	for _, s := range strings.Split(string(out), "\n") {
		// look up user
		usr, err := user.Lookup(s)
		if err != nil {
			continue
		}
		users = append(users, usr)
	}
	return users, nil
}

func shouldRunForUser_platform(usr *user.User) bool {
	res := !strings.HasPrefix(usr.Username, "_")
	if !res {
		common.Logger.Printf("Excluding %v: username starts with \"_\"",
			usr.Username)
	}
	return res
}
