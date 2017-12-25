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

/*
Get all users that have home dirs.
*/
func listUsers() ([]*user.User, error) {
	users := make([]*user.User, 0)

	// open passwd
	f, err := os.Open("/etc/passwd")
	if err != nil {
		common.ErrLogger.Printf("Failed to open /etc/passwd: %v\n", err)
		return users, err
	}
	defer f.Close()

	// look for users with home dirs
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) == 0 {
			continue
		}
		usr, err := user.Lookup(parts[0])
		if err != nil {
			continue
		}
		if len(usr.HomeDir) == 0 {
			continue
		}
		users = append(users, usr)
	}
	return users, nil
}

func jobfileForUser(user *user.User) *string {
	/*
	 * Not all users listed in /etc/passwd have their own
	 * jobber file.  E.g., some of them may share a home dir.
	 * When this happens, we say that the jobber file belongs
	 * to the user who owns that file.
	 */

	// make path to jobber file
	jobfilePath, err := common.JobfilePath(user)
	if err != nil {
		common.ErrLogger.Printf("%v", err)
		return nil
	}

	// open it
	f, err := os.Open(jobfilePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	// check owner
	owns, err := common.UserOwnsFile(user, jobfilePath)
	if !owns || err != nil {
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
			"Failed to make dir at %v: %v",
			common.VarDirPath,
			err)
		os.Exit(1)
	}

	// get all users
	users, err := listUsers()
	if err != nil {
		os.Exit(1)
	}

	mainCtx, mainCtxCtl :=
		common.MakeChildContext(context.Background())
	for _, usr := range users {
		// look for jobfile
		jobfilePath := filepath.Join(usr.HomeDir, common.JobFileName)

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
