package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

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

type prefs struct {
	VarDir     *string `yaml:"var-dir"`
	LibexecDir *string `yaml:"libexec-dir"`
	TempDir    *string `yaml:"temp-dir"`

	UsersInclude []UserSpec `yaml:"users-include"`
	UsersExclude []UserSpec `yaml:"users-exclude"`
}

type settingOrigin int

const (
	settingOriginDefault settingOrigin = iota
	settingOriginSet     settingOrigin = iota
)

func (self settingOrigin) String() string {
	switch self {
	case settingOriginDefault:
		return "default"
	case settingOriginSet:
		return "set"
	default:
		panic("invalid value")
	}
}

type setting struct {
	origin   string
	setValue *string
	def      string
}

func newSetting(def string) setting {
	var s setting
	s.def = def
	s.clear()
	return s
}

func (self *setting) get() (string, string) {
	if self.setValue == nil {
		return self.def, self.origin
	}
	return *self.setValue, self.origin
}

func (self *setting) set(val string, origin string) {
	self.origin = origin
	self.setValue = &val
}

func (self *setting) clear() {
	self.origin = "default"
}

func (self setting) String() string {
	v, o := self.get()
	return fmt.Sprintf("%v (%v)", v, o)
}

// This can be set at compile-time with ""-ldflags""
var etcDirPath = "/etc"

var gVarDir = newSetting("/var")
var gLibexecDir = newSetting("/usr/libexec")
var gTempDir = newSetting(os.TempDir())
var gEtcDir = newSetting(etcDirPath)
var gPrfs prefs

type InitSettingsParams struct {
	VarDir     *string
	LibexecDir *string
	TempDir    *string
	EtcDir     *string
}

func InitSettings(params InitSettingsParams) error {
	if params.EtcDir != nil {
		gEtcDir.set(*params.EtcDir, "cmdline")
	} else {
		gEtcDir.clear()
	}

	prfs, err := loadPrefs()
	if err != nil {
		return err
	}
	gPrfs = *prfs

	if params.VarDir != nil {
		gVarDir.set(*params.VarDir, "cmdline")
	} else if gPrfs.VarDir != nil {
		gVarDir.set(*gPrfs.VarDir, "config")
	} else {
		gVarDir.clear()
	}
	if params.LibexecDir != nil {
		gLibexecDir.set(*params.LibexecDir, "cmdline")
	} else if gPrfs.LibexecDir != nil {
		gLibexecDir.set(*gPrfs.LibexecDir, "config")
	} else {
		gLibexecDir.clear()
	}
	if params.TempDir != nil {
		gTempDir.set(*params.TempDir, "cmdline")
	} else if gPrfs.TempDir != nil {
		gTempDir.set(*gPrfs.TempDir, "config")
	} else {
		gTempDir.clear()
	}

	return nil
}

func VarDirPath() string {
	basePath, _ := gVarDir.get()
	return filepath.Join(basePath, "jobber")
}

func LibexecDirPath() string {
	path, _ := gLibexecDir.get()
	return path
}

func EtcDirPath() string {
	path, _ := gEtcDir.get()
	return path
}

func TempDirPath() string {
	path, _ := gTempDir.get()
	return path
}

func PerUserDirPath(usr *user.User) string {
	return filepath.Join(VarDirPath(), usr.Uid)
}

func CmdSocketPath(usr *user.User) string {
	const cmdSocketFileName = "cmd.sock"
	return filepath.Join(PerUserDirPath(usr), cmdSocketFileName)
}

func LibexecProgramPath(name string) string {
	return filepath.Join(LibexecDirPath(), name)
}

/*
 Get a list of all users for whom there is a jobberrunner process.
*/
func AllUsersWithSockets() ([]*user.User, error) {
	// get list of per-user dirs
	files, err := ioutil.ReadDir(VarDirPath())
	if err != nil {
		return nil, err
	}

	// make list of users
	users := make([]*user.User, 0)
	for _, file := range files {
		if file.IsDir() {
			usr, err := user.LookupId(file.Name())
			if err != nil {
				continue
			}
			users = append(users, usr)
		}
	}
	return users, nil
}

func PrintPaths() {
	fmt.Printf("var: %v\n", gVarDir)
	fmt.Printf("libexec: %v\n", gLibexecDir)
	fmt.Printf("etc: %v\n", gEtcDir)
	fmt.Printf("temp: %v\n", gTempDir)
}

const gPrefsFileName = "jobber.conf"
const gYamlStarter = "---"

const gDefaultPrefsStr = `## Here, you can control which users can use Jobber to run jobs.  You
## can either specify which users should be able to use Jobber, or
## which users should NOT be able to use Jobber --- not both.
##
## NOTE: Users without home directories, or who do not own their home
## directories, will not be able to use Jobber, no matter what you
## specify in this file.

## EXAMPLE: With the following, the only users that can use jobber are
## (1) root and (2) all users whose home directories are in
## /home/svcusers.
#users-include:
#    - username: root
#    - home: /home/svcusers/*

## EXAMPLE: With the following, the users postfix and mysql and all
## users whose usernames end with "nobody" cannot use Jobber.  (In the
## last rule, "*nobody" is quoted because "*" has a special meaning in
## YAML when unquoted.)
#users-exclude:
#    - username: postfix
#    - username: mysql
#    - username: '*nobody'
`

func MakeDefaultPrefs(params InitSettingsParams) string {
	var paths []string
	if params.VarDir != nil {
		paths = append(paths, fmt.Sprintf("var-dir: %s", *params.VarDir))
	}
	if params.LibexecDir != nil {
		paths = append(paths, fmt.Sprintf("libexec-dir: %s", *params.LibexecDir))
	}
	if params.TempDir != nil {
		paths = append(paths, fmt.Sprintf("temp-dir: %s", *params.TempDir))
	}
	allPaths := strings.Join(paths, "\n")
	return fmt.Sprintf("%s\n\n%s", allPaths, gDefaultPrefsStr)
}

func loadPrefs() (*prefs, error) {
	path := filepath.Join(EtcDirPath(), gPrefsFileName)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			Logger.Println("No prefs file.  Using defaults.")
			return &prefs{}, err
		} else {
			return nil, err
		}
	}
	defer f.Close()
	return _loadPrefs(f)
}

func _loadPrefs(f *os.File) (*prefs, error) {
	// read file
	bytes, err := ioutil.ReadAll(f)

	// add YAML prefix
	contents := string(bytes)
	if !strings.HasPrefix(contents, gYamlStarter+"\n") {
		contents = gYamlStarter + "\n" + contents
	}

	// parse YAML
	var prfs prefs
	err = yaml.Unmarshal([]byte(contents), &prfs)
	if err != nil {
		return nil, err
	}

	// check specs
	if len(prfs.UsersInclude) > 0 && len(prfs.UsersExclude) > 0 {
		msg := "Cannot use both \"users-include\" and " +
			"\"users-include\" in prefs."
		return nil, &Error{What: msg}
	}

	// check spec patterns
	var patterns []string
	specs := append(prfs.UsersInclude, prfs.UsersExclude...)
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

	return &prfs, nil
}

func JobberShouldRunForUser(usr *user.User) bool {
	return gPrfs.jobberShouldRunForUser(usr)
}

func (self *prefs) jobberShouldRunForUser(usr *user.User) bool {
	if len(self.UsersInclude) > 0 {
		return self._jobberShouldRunForUser(usr, self.UsersInclude, true)
	} else if len(self.UsersExclude) > 0 {
		return self._jobberShouldRunForUser(usr, self.UsersExclude, false)
	} else {
		return true
	}
}

func (self *prefs) _jobberShouldRunForUser(usr *user.User, specs []UserSpec,
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
