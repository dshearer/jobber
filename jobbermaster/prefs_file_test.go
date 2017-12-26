package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"os/user"
	"testing"
)

type LoadPrefsTestCase struct {
	Input  string
	Output *Prefs
	Err    bool
}

type ShouldIncludeTestCase struct {
	InputPrefs Prefs
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
		Output: &Prefs{
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
		Output: &Prefs{
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
		InputPrefs: Prefs{
			UsersInclude: []UserSpec{
				{Username: "mysql*"},
				{Username: "sys*", Home: "/a/b"},
			},
		},
		InputUser: user.User{Username: "mysqlX"},
		Output:    true,
	},
	{
		InputPrefs: Prefs{
			UsersInclude: []UserSpec{
				{Username: "mysql*"},
				{Username: "sys*", Home: "/a/b"},
			},
		},
		InputUser: user.User{Username: "sys", HomeDir: "/a/b"},
		Output:    true,
	},
	{
		InputPrefs: Prefs{
			UsersInclude: []UserSpec{
				{Username: "mysql*"},
				{Username: "sys*", Home: "/a/b"},
			},
		},
		InputUser: user.User{Username: "sys"},
		Output:    false,
	},
	{
		InputPrefs: Prefs{
			UsersExclude: []UserSpec{
				{Username: "mysql*"},
				{Username: "sys*", Home: "/a/b"},
			},
		},
		InputUser: user.User{Username: "mysqlX"},
		Output:    false,
	},
	{
		InputPrefs: Prefs{
			UsersExclude: []UserSpec{
				{Username: "mysql*"},
				{Username: "sys*", Home: "/a/b"},
			},
		},
		InputUser: user.User{Username: "sys", HomeDir: "/a/b"},
		Output:    false,
	},
	{
		InputPrefs: Prefs{
			UsersExclude: []UserSpec{
				{Username: "mysql*"},
				{Username: "sys*", Home: "/a/b"},
			},
		},
		InputUser: user.User{Username: "sys"},
		Output:    true,
	},
	{
		InputPrefs: Prefs{},
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
		prefs, err := loadPrefs(f)

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
			testCase.InputPrefs.ShouldIncludeUser(&testCase.InputUser)

		/*
		 * Test
		 */
		require.Equal(t, testCase.Output, result)
	}
}
