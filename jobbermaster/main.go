package main

import (
	"bufio"
	"context"
	"github.com/dshearer/jobber/common"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

/*

1. Determine which users need a jobberrunner process.
2. Make dir for each user at /var/jobber/{uid}.
3. Launch jobberrunner for each user.
4. Monitor jobberrunner processes.

*/

const gJobFileName = ".jobber"

type RunnerProcInfo struct {
	user        *user.User
	socketPath  string
	jobfilePath string
	proc        *exec.Cmd
}

func runnerThread(ctx common.BetterContext,
	usr *user.User,
	jobfilePath string) {

	common.Logger.Printf("Entered thread for %v", usr.Username)
	defer ctx.Finish()

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

		case <-ctx.Done():
			common.Logger.Printf("%v thread cancelled", usr.Username)
			/* No need to kill the child procs. */
			break Loop
		}
	}

	common.Logger.Printf("Exiting thread for %v", usr.Username)
}

func shouldRunForUser(usr *user.User, prefs *Prefs) bool {
	// check prefs
	if !prefs.ShouldIncludeUser(usr) {
		common.Logger.Printf("Excluding %v according to prefs",
			usr.Username)
		return false
	}

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

/*
Get all users that have home dirs.
*/
func listUsers(prefs *Prefs) ([]*user.User, error) {
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

		// check for reasons to exclude
		if !shouldRunForUser(usr, prefs) {
			continue
		}

		users = append(users, usr)
	}
	return users, nil
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
	common.UseSyslog()

	// make var dir
	if err := mkdirp(common.VarDirPath, 0775); err != nil {
		// already exists
		common.ErrLogger.Printf(
			"Failed to make dir at %v: %v",
			common.VarDirPath,
			err)
		os.Exit(1)
	}

	// load prefs
	prefs, err := LoadPrefs()
	if err != nil {
		common.ErrLogger.Printf("Invalid prefs file: %v", err)
		common.Logger.Println("Using default prefs.")
		prefs = &DefaultPrefs
	}

	// get all users
	users, err := listUsers(prefs)
	if err != nil {
		os.Exit(1)
	}

	mainCtx, mainCtxCtl :=
		common.MakeChildContext(context.Background())
	for _, usr := range users {
		// look for jobfile
		jobfilePath := filepath.Join(usr.HomeDir, gJobFileName)

		// make dir that will contain socket
		dirPath := common.PerUserDirPath(usr)
		if err := mkdirp(dirPath, 0770); err != nil {
			common.ErrLogger.Printf(
				"Failed to make dir at %v: %v",
				dirPath,
				err)
			continue
		}

		// set its owner
		if err := common.Chown(dirPath, usr); err != nil {
			common.ErrLogger.Printf(
				"Failed to chown dir at %v: %v",
				dirPath,
				err)
			continue
		}

		// launch thread to monitor runner process
		subctx, _ := common.MakeChildContext(mainCtx)
		go runnerThread(subctx, usr, jobfilePath)
	}

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, os.Kill)
	<-c

	// kill threads
	common.Logger.Printf("Killing threads")
	mainCtxCtl.Cancel()
	common.Logger.Printf("Waiting for threads")
	mainCtx.WaitForChildren()
	common.Logger.Printf("Done waiting for threads")
}
