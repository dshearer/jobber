package main

import (
	"fmt"
	"github.com/dshearer/jobber/common"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const gPrefsPath = "/etc/jobber.conf"
const gYamlStarter = "---"

/*
Example:

    users-include:
        - username: mysql*
        - home: /bin/*blah
        - username: sys*
          home: /something/here
*/

type UserSpec struct {
	Username string
	Home     string
}

type Prefs struct {
	UsersInclude []UserSpec `yaml:"users-include"`
	UsersExclude []UserSpec `yaml:"users-exclude"`
}

var DefaultPrefs = Prefs{}

func LoadPrefs() (*Prefs, error) {
	f, err := os.Open(gPrefsPath)
	if err != nil {
		if os.IsNotExist(err) {
			common.Logger.Println("No prefs file.  Using defaults.")
			return &DefaultPrefs, err
		} else {
			return nil, err
		}
	}
	common.Logger.Println("Loaded prefs from %v", gPrefsPath)
	defer f.Close()
	return loadPrefs(f)
}

func loadPrefs(f *os.File) (*Prefs, error) {
	// read file
	bytes, err := ioutil.ReadAll(f)

	// add YAML prefix
	contents := string(bytes)
	if !strings.HasPrefix(contents, gYamlStarter+"\n") {
		contents = gYamlStarter + "\n" + contents
	}

	// parse YAML
	var prefs Prefs
	err = yaml.Unmarshal([]byte(contents), &prefs)
	if err != nil {
		return nil, err
	}

	// check specs
	if len(prefs.UsersInclude) > 0 && len(prefs.UsersExclude) > 0 {
		msg := "Cannot use both \"users-include\" and " +
			"\"users-include\" in prefs."
		return nil, &common.Error{What: msg}
	}

	// check spec patterns
	var patterns []string
	specs := append(prefs.UsersInclude, prefs.UsersExclude...)
	for _, spec := range specs {
		if len(spec.Username) > 0 {
			patterns = append(patterns, spec.Username)
		}
		if len(spec.Home) > 0 {
			patterns = append(patterns, spec.Home)
		}
	}
	for _, pattern := range patterns {
		_, err = filepath.Match(pattern, "something")
		if err != nil {
			return nil, err
		}
	}

	return &prefs, nil
}

func (self *Prefs) ShouldIncludeUser(usr *user.User) bool {
	if len(self.UsersInclude) > 0 {
		return self.shouldIncludeUser(usr, self.UsersInclude, true)
	} else if len(self.UsersExclude) > 0 {
		return self.shouldIncludeUser(usr, self.UsersExclude, false)
	} else {
		return true
	}
}

func (self *Prefs) shouldIncludeUser(usr *user.User, specs []UserSpec,
	includeIfMatch bool) bool {

	for _, spec := range specs {
		if spec.matches(usr) {
			return includeIfMatch
		}
	}
	return !includeIfMatch
}

func (self *UserSpec) matches(usr *user.User) bool {
	match := func(pattern, name string) bool {
		b, err := filepath.Match(pattern, name)
		if err != nil {
			panic(fmt.Sprintf("%v", err))
		}
		return b
	}

	if len(self.Username) > 0 && !match(self.Username, usr.Username) {
		return false
	}
	if len(self.Home) > 0 && !match(self.Home, usr.HomeDir) {
		return false
	}
	return true
}
