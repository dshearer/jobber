// +build !darwin

package main

import (
	"bufio"
	"os"
	"os/user"
	"strings"

	"github.com/dshearer/jobber/common"
)

/*
Get all users.
*/
func getAllUsers() ([]*user.User, error) {
	users := make([]*user.User, 0)

	// open passwd
	f, err := os.Open("/etc/passwd")
	if err != nil {
		common.ErrLogger.Printf("Failed to open /etc/passwd: %v\n", err)
		return users, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// look up user
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) == 0 {
			continue
		}
		usr, err := user.Lookup(parts[0])
		if err != nil {
			continue
		}

		users = append(users, usr)
	}
	return users, nil
}

func shouldRunForUser_platform(usr *user.User) bool {
	return true
}
