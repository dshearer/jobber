package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"testing"
	"unicode"

	"github.com/dshearer/jobber/jobfile"
	"github.com/stretchr/testify/require"
)

var gUserEx = user.User{Username: "bob", HomeDir: "/home/bob"}

func TestParseDefaultJobfile(t *testing.T) {
	/*
	 * Set up
	 */
	f, err := ioutil.TempFile("", "Testing")
	if err != nil {
		panic(fmt.Sprintf("Failed to make tempfile: %v", err))
	}
	defer f.Close()
	defer os.Remove(f.Name())
	f.Write([]byte(gDefaultJobfile))
	f.Seek(0, 0)

	/*
	 * Call
	 */
	var raw *jobfile.JobFileRaw
	raw, err = jobfile.LoadJobfile(f)
	var file *jobfile.JobFile
	if raw != nil {
		file, err = raw.Activate(&gUserEx)
	}

	/*
	 * Test
	 */
	require.Nil(t, err, "%v", err)
	require.NotNil(t, file)
}

func uncommentLine(line string) string {
	nonWsIdx := strings.IndexFunc(line, func(r rune) bool {
		return !unicode.IsSpace(r)
	})
	if nonWsIdx < 0 || line[nonWsIdx] != '#' {
		return line
	} else if len(line) == nonWsIdx+1 {
		return line[:nonWsIdx]
	} else {
		return line[:nonWsIdx] + line[nonWsIdx+1:]
	}
}

func TestParseDefaultJobfileAfterUncommenting(t *testing.T) {
	/*
	 * Set up
	 */

	// uncomment certain lines
	var jfile string
	lines := strings.Split(gDefaultJobfile, "\n")
	for _, line := range lines {
		jfile += uncommentLine(line) + "\n"
	}
	fmt.Printf("Jobfile:\n%v\n", jfile)

	// write jobfile
	f, err := ioutil.TempFile("", "Testing")
	if err != nil {
		panic(fmt.Sprintf("Failed to make tempfile: %v", err))
	}
	defer f.Close()
	defer os.Remove(f.Name())
	f.WriteString(jfile)
	f.Seek(0, 0)

	/*
	 * Call
	 */
	var raw *jobfile.JobFileRaw
	raw, err = jobfile.LoadJobfile(f)
	var file *jobfile.JobFile
	if raw != nil {
		file, err = raw.Activate(&gUserEx)
	}

	/*
	 * Test
	 */
	require.Nil(t, err, "%v", err)
	require.NotNil(t, file)
}
