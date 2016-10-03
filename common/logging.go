package common

import (
    "log"
    "os"
)

var Logger *log.Logger = log.New(os.Stdout, "", 0)
var ErrLogger *log.Logger = log.New(os.Stderr, "", 0)