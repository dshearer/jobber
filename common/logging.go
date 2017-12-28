package common

import (
	"io"
	"log"
	"log/syslog"
	"os"
)

var Logger *log.Logger = log.New(os.Stdout, "", 0)
var ErrLogger *log.Logger = log.New(os.Stderr, "", 0)

func UseSyslog() error {
	logger, err :=
		syslog.NewLogger(syslog.LOG_INFO|syslog.LOG_DAEMON, 0)
	if err != nil {
		return err
	}
	errLogger, err :=
		syslog.NewLogger(syslog.LOG_ERR|syslog.LOG_DAEMON, 0)
	if err != nil {
		return err
	}
	Logger = logger
	ErrLogger = errLogger
	return nil
}

func SetLogFile(paths ...string) {
	if len(paths) == 1 {
		// open log file
		f, err := os.OpenFile(paths[0],
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			ErrLogger.Printf("Failed to open log file %v", paths[0])
			return
		}

		// make loggers
		stdoutWriter := io.MultiWriter(os.Stdout, f)
		stderrWriter := io.MultiWriter(os.Stderr, f)
		Logger = log.New(stdoutWriter, "", 0)
		ErrLogger = log.New(stderrWriter, "", 0)

	} else if len(paths) == 2 {
		// open log files
		outF, err := os.OpenFile(paths[0],
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			ErrLogger.Printf("Failed to open log file %v", paths[0])
			return
		}
		errF, err := os.OpenFile(paths[1],
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			outF.Close()
			ErrLogger.Printf("Failed to open log file %v", paths[1])
			return
		}

		// make loggers
		stdoutWriter := io.MultiWriter(os.Stdout, outF)
		stderrWriter := io.MultiWriter(os.Stderr, errF)
		Logger = log.New(stdoutWriter, "", 0)
		ErrLogger = log.New(stderrWriter, "", 0)

	} else {
		panic("Invalid paths arg")
	}
}
