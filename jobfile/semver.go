package jobfile

import (
	"strconv"
	"strings"

	"github.com/dshearer/jobber/common"
)

type SemVer struct {
	Major uint
	Minor uint
	Patch uint
}

func ParseSemVer(s string) (*SemVer, error) {
	defaultError := common.Error{What: "Invalid Semantic Version"}

	parts := strings.Split(s, ".")
	if len(parts) > 3 || len(parts) == 0 {
		return nil, &defaultError
	}

	var ver SemVer
	if len(parts) > 0 {
		val, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return nil, err
		}
		ver.Major = uint(val)
	}
	if len(parts) > 1 {
		val, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return nil, err
		}
		ver.Minor = uint(val)
	}
	if len(parts) > 2 {
		val, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			return nil, err
		}
		ver.Patch = uint(val)
	}

	return &ver, nil
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
