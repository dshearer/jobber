package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type LoadPrefsTestCase struct {
	Input  string
	Output *prefs
	Err    bool
}

type ShouldIncludeTestCase struct {
	InputPrefs prefs
	InputUser  user.User
	Output     bool
}

var gLoadPrefsTestCases = []LoadPrefsTestCase{
	{
		Input: `
users-include:
    - username: mysql*
    - home: /bin/*blah
    - username: sys*
      home: /something/here`,
		Output: &prefs{
			UsersInclude: []UserSpec{
				{Username: "mysql*"},
				{Home: "/bin/*blah"},
				{Username: "sys*", Home: "/something/here"},
			},
		},
	},
	{
		Input: `
users-include:
    - username: mysql*
    - home: /bin/*blah
    - username: sys*
      home: /something/here
users-exclude:`,
		Output: &prefs{
			UsersInclude: []UserSpec{
				{Username: "mysql*"},
				{Home: "/bin/*blah"},
				{Username: "sys*", Home: "/something/here"},
			},
		},
	},
	{
		Input: `
users-include:
    - username: mysql*
users-exclude:
    - home: /bin/*blah
    - username: sys*
      home: /something/here`,
		Err: true,
	},
}

var gShouldIncludeTestCases = []ShouldIncludeTestCase{
	{
		InputPrefs: prefs{
			UsersInclude: []UserSpec{
				{Username: "mysql*"},
				{Username: "sys*", Home: "/a/b"},
			},
		},
		InputUser: user.User{Username: "mysqlX"},
		Output:    true,
	},
	{
		InputPrefs: prefs{
			UsersInclude: []UserSpec{
				{Username: "mysql*"},
				{Username: "sys*", Home: "/a/b"},
			},
		},
		InputUser: user.User{Username: "sys", HomeDir: "/a/b"},
		Output:    true,
	},
	{
		InputPrefs: prefs{
			UsersInclude: []UserSpec{
				{Username: "mysql*"},
				{Username: "sys*", Home: "/a/b"},
			},
		},
		InputUser: user.User{Username: "sys"},
		Output:    false,
	},
	{
		InputPrefs: prefs{
			UsersExclude: []UserSpec{
				{Username: "mysql*"},
				{Username: "sys*", Home: "/a/b"},
			},
		},
		InputUser: user.User{Username: "mysqlX"},
		Output:    false,
	},
	{
		InputPrefs: prefs{
			UsersExclude: []UserSpec{
				{Username: "mysql*"},
				{Username: "sys*", Home: "/a/b"},
			},
		},
		InputUser: user.User{Username: "sys", HomeDir: "/a/b"},
		Output:    false,
	},
	{
		InputPrefs: prefs{
			UsersExclude: []UserSpec{
				{Username: "mysql*"},
				{Username: "sys*", Home: "/a/b"},
			},
		},
		InputUser: user.User{Username: "sys"},
		Output:    true,
	},
	{
		InputPrefs: prefs{},
		InputUser:  user.User{Username: "sys"},
		Output:     true,
	},
}

func TestLoadPrefs(t *testing.T) {
	for _, testCase := range gLoadPrefsTestCases {
		/*
		 * Set up
		 */
		fmt.Printf("Input:\n%v\n", testCase.Input)

		// make prefs file
		f, err := ioutil.TempFile("", "Testing")
		if err != nil {
			panic(fmt.Sprintf("Failed to make tempfile: %v", err))
		}
		defer os.Remove(f.Name())
		f.WriteString(testCase.Input)
		f.Seek(0, 0)
		defer f.Close()

		/*
		 * Call
		 */
		prefs, err := _loadPrefs(f)

		/*
		 * Test
		 */
		if testCase.Err {
			require.Nil(t, prefs)
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
			require.NotNil(t, prefs)
			require.Equal(t, testCase.Output, prefs)
		}
	}
}

func TestShouldIncludeUser(t *testing.T) {
	for _, testCase := range gShouldIncludeTestCases {
		/*
		 * Call
		 */
		fmt.Printf("%v\n", testCase.InputPrefs)
		fmt.Printf("%v\n", testCase.InputUser)
		result :=
			testCase.InputPrefs.jobberShouldRunForUser(&testCase.InputUser)

		/*
		 * Test
		 */
		require.Equal(t, testCase.Output, result)
	}
}

func TestParseDefaultPrefs(t *testing.T) {
	/*
	 * Set up
	 */
	// make prefs file
	f, err := ioutil.TempFile("", "Testing")
	if err != nil {
		panic(fmt.Sprintf("Failed to make tempfile: %v", err))
	}
	defer os.Remove(f.Name())
	f.WriteString(gDefaultPrefsStr)
	f.Seek(0, 0)
	defer f.Close()

	/*
	 * Call
	 */
	prefs, err := _loadPrefs(f)

	/*
	 * Test
	 */
	require.Nil(t, err)
	require.NotNil(t, prefs)
}

func TestParseDefaultPrefsAfterUncommenting(t *testing.T) {
	/*
	 * Set up
	 */
	f, err := ioutil.TempFile("", "Testing")
	if err != nil {
		panic(fmt.Sprintf("Failed to make tempfile: %v", err))
	}
	defer os.Remove(f.Name())
	lines := strings.Split(gDefaultPrefsStr, "\n")
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
	prefs, err := _loadPrefs(f)

	/*
	 * Test
	 */
	require.Nil(t, err)
	require.NotNil(t, prefs)
}
