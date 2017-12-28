package main

import (
	"fmt"
	"github.com/dshearer/jobber/jobfile"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"testing"
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
	var file *jobfile.JobFile
	file, err = jobfile.LoadJobfile(f, &gUserEx)

	/*
	 * Test
	 */
	require.Nil(t, err, "%v", err)
	require.NotNil(t, file)
}

func TestParseDefaultJobfileAfterUncommenting(t *testing.T) {
	/*
	 * Set up
	 */
	f, err := ioutil.TempFile("", "Testing")
	if err != nil {
		panic(fmt.Sprintf("Failed to make tempfile: %v", err))
	}
	defer f.Close()
	defer os.Remove(f.Name())
	lines := strings.Split(gDefaultJobfile, "\n")
	for _, line := range lines {
		if len(line) > 2 && line[0] == '#' && line[1] != '#' {
			line = line[1:]
		}
		f.WriteString(line + "\n")
	}
	f.Seek(0, 0)

	/*
	 * Call
	 */
	var file *jobfile.JobFile
	file, err = jobfile.LoadJobfile(f, &gUserEx)

	/*
	 * Test
	 */
	require.Nil(t, err, "%v", err)
	require.NotNil(t, file)
}
