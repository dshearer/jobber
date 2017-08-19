package main

import (
	"bufio"
	"github.com/dshearer/jobber/common"
	"github.com/dshearer/jobber/jobfile"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

/*

1. Determine which users need a jobberrunner process.
2. Make dir for each user at /var/jobber/{uid}.
3. Launch jobberrunner for each user.
4. Monitor jobberrunner processes.

*/

type RunnerProcInfo struct {
	user        *user.User
	socketPath  string
	jobfilePath string
	proc        *exec.Cmd
}

func runnerThread(ctx *common.NewContext,
	usr *user.User,
	jobfilePath string) {

	common.Logger.Printf("Entered thread for %v", usr.Username)

Loop:
	for {
		// spawn runner process
		common.Logger.Println("Launching runner")
		proc, err := LaunchRunner(usr, jobfilePath)
		if err != nil {
			common.ErrLogger.Printf(
				"Failed to launch runner for %v: %v",
				usr.Username,
				err,
			)
			return
		}

		// wait for process to exit or context to be cancelled
		select {
		case err := <-proc.ExitedChan:
			if err != nil {
				common.ErrLogger.Printf(
					"Runner for %v exited with error: %v",
					usr.Username,
					err,
				)
				if _, flag := err.(*exec.ExitError); !flag {
					break Loop
				}
			}
			common.Logger.Printf(
				"Restarting runner for %v",
				usr.Username,
			)

		case <-ctx.CancelledChan():
			common.Logger.Printf("%v thread cancelled", usr.Username)
			proc.Kill()
			break Loop
		}
	}

	ctx.Finish()
	common.Logger.Printf("Exiting thread for %v", usr.Username)
}

func listUsers() ([]string, error) {
	usernames := make([]string, 0)
	f, err := os.Open("/etc/passwd")
	if err != nil {
		common.ErrLogger.Printf("Failed to open /etc/passwd: %v\n", err)
		return usernames, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) > 0 {
			usernames = append(usernames, parts[0])
		}
	}
	return usernames, nil
}

func jobfileForUser(user *user.User) *string {
	/*
	 * Not all users listed in /etc/passwd have their own
	 * jobber file.  E.g., some of them may share a home dir.
	 * When this happens, we say that the jobber file belongs
	 * to the user who owns that file.
	 */

	// make path to jobber file
	jobfilePath := filepath.Join(user.HomeDir, jobfile.JobFileName)

	// open it
	f, err := os.Open(jobfilePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	// check owner
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil
	}
	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return nil
	}
	if uint32(uid) != info.Sys().(*syscall.Stat_t).Uid {
		return nil
	}

	return &jobfilePath
}

func mkdirp(path string, perm os.FileMode) error {
	if err := os.Mkdir(path, perm); err != nil {
		if err.(*os.PathError).Err.(syscall.Errno) != 17 {
			return err
		}
	}
	return nil
}

func main() {
	// make var dir
	if err := mkdirp(common.VarDirPath, 0775); err != nil {
		// already exists
		common.ErrLogger.Printf(
			"Failed to make dir at %v: %t",
			common.VarDirPath,
			err)
		os.Exit(1)
	}

	// get all users and jobfiles by reading passwd
	usernames, err := listUsers()
	if err != nil {
		os.Exit(1)
	}

	mainCtx := common.BackgroundContext().MakeChild()
	for _, username := range usernames {
		// look up user
		usr, err := user.Lookup(username)
		if err != nil {
			continue
		}

		// look for jobfile
		jobfilePath := jobfileForUser(usr)
		if jobfilePath == nil {
			// no jobfile for this user
			continue
		}

		// make dir that will contain socket
		dirPath := common.PerUserDirPath(usr)
		if err := mkdirp(dirPath, 0770); err != nil {
			common.ErrLogger.Printf(
				"Failed to make dir at %v: %t",
				dirPath,
				err)
			continue
		}

		// set its owner
		uid, _ := strconv.Atoi(usr.Uid)
		gid, _ := strconv.Atoi(usr.Gid)
		if err := os.Chown(dirPath, uid, gid); err != nil {
			common.ErrLogger.Printf(
				"Failed to chown dir at %v: %t",
				dirPath,
				err)
			continue
		}

		// launch thread to monitor runner process
		subctx := mainCtx.MakeChild()
		go runnerThread(subctx, usr, *jobfilePath)
	}

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, os.Kill)
	<-c

	// kill threads
	common.Logger.Printf("Killing threads")
	mainCtx.Cancel()
	common.Logger.Printf("Waiting for threads")
	mainCtx.Finish()
	common.Logger.Printf("Done waiting for threads")
}
