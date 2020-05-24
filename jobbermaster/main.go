package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"sync"
	"syscall"

	arg "github.com/alexflint/go-arg"
	"github.com/dshearer/jobber/common"
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

type defprefsArgs struct {
}

type debugArgs struct {
}

type argsS struct {
	Defprefs *defprefsArgs `arg:"subcommand:defprefs"`
	Debug    *debugArgs    `arg:"subcommand:debug"`
	Etc      *string       `arg:"-e"`
	Var      *string       `arg:"-r"`
	Libexec  *string       `arg:"-l"`
	Temp     *string       `arg:"-t"`
}

func (argsS) Version() string {
	return common.LongVersionStr()
}

func runnerThread(ctx context.Context,
	usr *user.User,
	jobfilePath string) {

Loop:
	for {
		// spawn runner process
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
				common.ErrLogger.Println(proc.Stderr())
				if _, flag := err.(*exec.ExitError); !flag {
					break Loop
				}
			}
			common.Logger.Printf(
				"Restarting runner for %v",
				usr.Username,
			)

		case <-ctx.Done():
			/* No need to kill the child procs. */
			break Loop
		}
	}
}

func mkdirp(path string, perm os.FileMode) error {
	if err := os.Mkdir(path, perm); err != nil {
		if err.(*os.PathError).Err.(syscall.Errno) != 17 {
			return err
		}
	}
	return nil
}

func doDefault(args argsS) int {
	common.LogToStdoutStderr()
	if err := initSettings(args); err != nil {
		common.ErrLogger.Println(err)
		return 1
	}

	common.Logger.Printf("etc dir: %v", common.EtcDirPath())
	common.Logger.Printf("var dir: %v", common.VarDirPath())
	common.Logger.Printf("libexec dir: %v", common.LibexecDirPath())
	common.Logger.Printf("temp dir: %v", common.TempDirPath())

	// make var dir
	if err := mkdirp(common.VarDirPath(), 0775); err != nil {
		// already exists
		common.ErrLogger.Printf(
			"Failed to make dir at %v: %v",
			common.VarDirPath(),
			err)
		return 1
	}

	// get all users
	users, err := getAcceptableUsers()
	if err != nil {
		return 1
	}

	ctx, cancelCtx :=
		context.WithCancel(context.Background())
	var runnerWaitGroup sync.WaitGroup
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
		runnerWaitGroup.Add(1)
		go func(u *user.User, p string) {
			defer runnerWaitGroup.Done()
			runnerThread(ctx, u, p)
		}(usr, jobfilePath)
	}

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, os.Kill)
	<-c

	// kill threads
	cancelCtx()
	runnerWaitGroup.Wait()

	return 0
}

func doDefprefs(args argsS) int {
	common.LogAllToStderr()
	s := common.MakeDefaultPrefs(common.InitSettingsParams{
		VarDir:     args.Var,
		LibexecDir: args.Libexec,
		TempDir:    args.Temp,
	})
	fmt.Println(s)
	return 0
}

func doDebug(args argsS) int {
	common.LogAllToStderr()
	if err := initSettings(args); err != nil {
		common.ErrLogger.Println(err)
		return 1
	}

	common.PrintPaths()
	return 0
}

func main() {
	// parse args
	var args argsS
	p := arg.MustParse(&args)

	// do command
	var exitVal int
	if p.Subcommand() == nil {
		exitVal = doDefault(args)
	} else if args.Defprefs != nil {
		exitVal = doDefprefs(args)
	} else if args.Debug != nil {
		exitVal = doDebug(args)
	} else {
		p.Fail("Invalid command")
	}
	os.Exit(exitVal)
}

func initSettings(args argsS) error {
	return common.InitSettings(common.InitSettingsParams{
		VarDir:     args.Var,
		LibexecDir: args.Libexec,
		TempDir:    args.Temp,
		EtcDir:     args.Etc,
	})
}
