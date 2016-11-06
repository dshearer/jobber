package jobfile

import (
    "github.com/dshearer/jobber/common"
    "fmt"
    "strings"
    "strconv"
)

const (
    TimeWildcard = "*"
)

type TimeSpec interface {
	String() string
	Satisfied(int) bool
}

type FullTimeSpec struct {
	Sec  TimeSpec
	Min  TimeSpec
	Hour TimeSpec
	Mday TimeSpec
	Mon  TimeSpec
	Wday TimeSpec
}

func (self FullTimeSpec) String() string {
    return fmt.Sprintf("%v %v %v %v %v %v",
                       self.Sec,
                       self.Min,
                       self.Hour,
                       self.Mday,
                       self.Mon,
                       self.Wday)
}

type WildcardTimeSpec struct {
}

func (s WildcardTimeSpec) String() string {
	return "*"
}

func (s WildcardTimeSpec) Satisfied(v int) bool {
	return true
}

type OneValTimeSpec struct {
	val int
}

func (s OneValTimeSpec) String() string {
	return fmt.Sprintf("%v", s.val)
}

func (s OneValTimeSpec) Satisfied(v int) bool {
	return s.val == v
}

type SetTimeSpec struct {
	desc string
	vals []int
}

func (s SetTimeSpec) String() string {
	return s.desc
}

func (s SetTimeSpec) Satisfied(v int) bool {
	for _, v2 := range s.vals {
		if v == v2 {
			return true
		}
	}
	return false
}

func ParseFullTimeSpec(s string) (*FullTimeSpec, error) {
	var fullSpec FullTimeSpec
	fullSpec.Sec = WildcardTimeSpec{}
	fullSpec.Min = WildcardTimeSpec{}
	fullSpec.Hour = WildcardTimeSpec{}
	fullSpec.Mday = WildcardTimeSpec{}
	fullSpec.Mon = WildcardTimeSpec{}
	fullSpec.Wday = WildcardTimeSpec{}

	var timeParts []string = strings.Fields(s)

	// sec
	if len(timeParts) > 0 {
		spec, err := parseTimeSpec(timeParts[0], "sec", 0, 59)
		if err != nil {
			return nil, err
		}
		fullSpec.Sec = spec
	}

	// min
	if len(timeParts) > 1 {
		spec, err := parseTimeSpec(timeParts[1], "minute", 0, 59)
		if err != nil {
			return nil, err
		}
		fullSpec.Min = spec
	}

	// hour
	if len(timeParts) > 2 {
		spec, err := parseTimeSpec(timeParts[2], "hour", 0, 23)
		if err != nil {
			return nil, err
		}
		fullSpec.Hour = spec
	}

	// mday
	if len(timeParts) > 3 {
		spec, err := parseTimeSpec(timeParts[3], "month day", 1, 31)
		if err != nil {
			return nil, err
		}
		fullSpec.Mday = spec
	}

	// month
	if len(timeParts) > 4 {
		spec, err := parseTimeSpec(timeParts[4], "month", 1, 12)
		if err != nil {
			return nil, err
		}
		fullSpec.Mon = spec
	}

	// wday
	if len(timeParts) > 5 {
		spec, err := parseTimeSpec(timeParts[5], "weekday", 0, 6)
		if err != nil {
			return nil, err
		}
		fullSpec.Wday = spec
	}

	if len(timeParts) > 6 {
		return nil, &common.Error{"Excess elements in 'time' field.", nil}
	}

	return &fullSpec, nil
}

func parseTimeSpec(s string, fieldName string, min int, max int) (TimeSpec, error) {
	errMsg := fmt.Sprintf("Invalid '%v' value", fieldName)

	if s == TimeWildcard {
		return WildcardTimeSpec{}, nil
	} else if strings.HasPrefix(s, "*/") {
		// parse step
		stepStr := s[2:]
		step, err := strconv.Atoi(stepStr)
		if err != nil {
			return nil, &common.Error{errMsg, err}
		}

		// make set of valid values
		vals := make([]int, 0)
		for v := min; v <= max; v = v + step {
			vals = append(vals, v)
		}

		// make spec
		spec := SetTimeSpec{vals: vals, desc: s}
		return spec, nil

	} else if strings.Contains(s, ",") {
		// split step
		stepStrs := strings.Split(s, ",")

		// make set of valid values
		vals := make([]int, 0)
		for _,stepStr := range stepStrs {
			step, err := strconv.Atoi(stepStr)
			if err != nil {
				return nil, &common.Error{errMsg, err}
			}
			vals = append(vals, step)
		}

		// make spec
		spec := SetTimeSpec{vals: vals, desc: s}
		return spec, nil
	} else if strings.Contains(s, "-") {
		// get range extremes
		extremes := strings.Split(s, "-")
		begin, err := strconv.Atoi(extremes[0])

		if err != nil {
			return nil, &common.Error{errMsg, err}
		}

		end, err := strconv.Atoi(extremes[1])

		if err != nil {
			return nil, &common.Error{errMsg, err}
		}

		// make set of valid values
		vals := make([]int, 0)

		for v := begin; v <= end; v++ {
			vals = append(vals, v)
		}

		// make spec
		spec := SetTimeSpec{vals: vals, desc: s}
		return spec, nil
	} else {
		// convert to int
		val, err := strconv.Atoi(s)
		if err != nil {
			return nil, &common.Error{errMsg, err}
		}

		// make TimeSpec
		spec := OneValTimeSpec{val}

		// check range
		if val < min {
			errMsg := fmt.Sprintf("%s: cannot be less than %v.", errMsg, min)
			return nil, &common.Error{errMsg, nil}
		} else if val > max {
			errMsg := fmt.Sprintf("%s: cannot be greater than %v.", errMsg, max)
			return nil, &common.Error{errMsg, nil}
		}

		return spec, nil
	}
}
