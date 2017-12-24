package main

import (
	"fmt"
	"github.com/dshearer/jobber/jobfile"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const UsernameEx string = "bob"

func TestParseDefaultJobfile(t *testing.T) {
	/*
	 * Set up
	 */
	f, err := ioutil.TempFile("", "Testing")
	if err != nil {
		panic(fmt.Sprintf("Failed to make tempfile: %v", err))
	}
	defer os.Remove(f.Name())
	f.Write([]byte(gDefaultJobfile))
	f.Close()

	/*
	 * Call
	 */
	var file *jobfile.JobFile
	file, err = jobfile.LoadJobFile(f.Name(), UsernameEx)

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
	defer os.Remove(f.Name())
	lines := strings.Split(gDefaultJobfile, "\n")
	for _, line := range lines {
		if len(line) > 2 && line[0] == '#' && line[1] != '#' {
			line = line[1:]
		}
		f.WriteString(line + "\n")
	}
	f.Close()

	/*
	 * Call
	 */
	var file *jobfile.JobFile
	file, err = jobfile.LoadJobFile(f.Name(), UsernameEx)

	/*
	 * Test
	 */
	require.Nil(t, err, "%v", err)
	require.NotNil(t, file)
}
