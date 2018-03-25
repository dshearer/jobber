package jobfile

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dshearer/jobber/common"
)

type SemVer struct {
	Major uint
	Minor uint
	Patch uint
}

func (self SemVer) IsZero() bool {
	return self.Major == 0 && self.Minor == 0 && self.Patch == 0
}

func (self SemVer) String() string {
	if self.Patch > 0 {
		return fmt.Sprintf("%v.%v.%v", self.Major, self.Minor, self.Patch)
	} else if self.Minor > 0 {
		return fmt.Sprintf("%v.%v", self.Major, self.Minor)
	} else {
		return fmt.Sprintf("%v", self.Major)
	}
}

func (self SemVer) MarshalJSON() ([]byte, error) {
	return []byte(self.String()), nil
}

func (self SemVer) MarshalYAML() (interface{}, error) {
	return self.String(), nil
}

func (self *SemVer) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return &common.Error{What: "Cannot unmarshal SemVer", Cause: err}
	}
	return self.fromString(s)
}

func (self *SemVer) fromString(s string) error {
	parts := strings.Split(s, ".")
	if len(parts) > 3 || len(parts) == 0 {
		return &common.Error{What: fmt.Sprintf("Invalid Semantic Version: \"%v\"", s)}
	}

	if len(parts) > 0 {
		val, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return err
		}
		self.Major = uint(val)
	}
	if len(parts) > 1 {
		val, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return err
		}
		self.Minor = uint(val)
	}
	if len(parts) > 2 {
		val, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			return err
		}
		self.Patch = uint(val)
	}

	return nil
}

func compareInts(a, b uint) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	} else {
		return 0
	}
}

func (self SemVer) Compare(other SemVer) int {
	if c := compareInts(self.Major, other.Major); c != 0 {
		return c
	} else if c := compareInts(self.Minor, other.Minor); c != 0 {
		return c
	} else {
		return compareInts(self.Patch, other.Patch)
	}
}
