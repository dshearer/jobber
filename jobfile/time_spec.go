package jobfile

//go:generate goyacc -o parse_time_spec.go parse_time_spec.y

import (
	"fmt"
	"github.com/dshearer/jobber/common"
	"math/rand"
	"strings"
	"time"
)

const (
	TimeWildcard = "*"
)

func monthToInt(m time.Month) int {
	switch m {
	case time.January:
		return 1
	case time.February:
		return 2
	case time.March:
		return 3
	case time.April:
		return 4
	case time.May:
		return 5
	case time.June:
		return 6
	case time.July:
		return 7
	case time.August:
		return 8
	case time.September:
		return 9
	case time.October:
		return 10
	case time.November:
		return 11
	default:
		return 12
	}
}

func weekdayToInt(d time.Weekday) int {
	switch d {
	case time.Sunday:
		return 0
	case time.Monday:
		return 1
	case time.Tuesday:
		return 2
	case time.Wednesday:
		return 3
	case time.Thursday:
		return 4
	case time.Friday:
		return 5
	default:
		return 6
	}
}

type TimeSpec interface {
	fmt.Stringer
	Satisfied(int) bool
	IsWildcard() bool
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

func (self *FullTimeSpec) Derandomize() {
	timeSpecs := [...]TimeSpec{
		self.Sec,
		self.Min,
		self.Hour,
		self.Mday,
		self.Mon,
		self.Wday,
	}
	for _, spec := range timeSpecs {
		switch randSpec := spec.(type) {
		case *RandomTimeSpec:
			randSpec.Derandomize()
		}
	}
}

func (self *FullTimeSpec) Satisfied(t time.Time) bool {
	nondayMatch := self.Sec.Satisfied(t.Second()) &&
		self.Min.Satisfied(t.Minute()) &&
		self.Hour.Satisfied(t.Hour()) &&
		self.Mon.Satisfied(monthToInt(t.Month()))

	/*
	   - If Mday and Wday are wildcards, then both must be satisfied.
	   - If exactly one of Mday and Wday is a wildcard, then both must
	   be satisfied.
	   - If neither Mday nor Wday is a wildcard, then either must be
	   satisfied.
	*/
	dayMatch := false
	if !self.Mday.IsWildcard() && !self.Wday.IsWildcard() {
		dayMatch = self.Wday.Satisfied(weekdayToInt(t.Weekday())) ||
			self.Mday.Satisfied(t.Day())
	} else {
		dayMatch = self.Wday.Satisfied(weekdayToInt(t.Weekday())) &&
			self.Mday.Satisfied(t.Day())
	}

	return nondayMatch && dayMatch
}

type WildcardTimeSpec struct{}

func (self *WildcardTimeSpec) IsWildcard() bool {
	return true
}

func (self *WildcardTimeSpec) String() string {
	return "*"
}

func (self *WildcardTimeSpec) Satisfied(v int) bool {
	return true
}

type OneValTimeSpec struct {
	val int
}

func (self *OneValTimeSpec) IsWildcard() bool {
	return false
}

func (self *OneValTimeSpec) String() string {
	return fmt.Sprintf("%v", self.val)
}

func (self *OneValTimeSpec) Satisfied(v int) bool {
	return self.val == v
}

type SetTimeSpec struct {
	desc string
	vals []int
}

func (self *SetTimeSpec) IsWildcard() bool {
	return false
}

func (self *SetTimeSpec) String() string {
	return self.desc
}

func (self *SetTimeSpec) Satisfied(v int) bool {
	for _, v2 := range self.vals {
		if v == v2 {
			return true
		}
	}
	return false
}

/*
A time spec that chooses (pseudo-)randomly from a set of values.
Each value in that set has an (approximately) equal chance of getting
picked.
*/
type RandomTimeSpec struct {
	desc      string
	vals      []int
	pickedVal *int
}

func (self *RandomTimeSpec) IsWildcard() bool {
	return false
}

func (self *RandomTimeSpec) String() string {
	if self.pickedVal == nil {
		return self.desc
	} else {
		return fmt.Sprintf("%v->%v", self.desc, *self.pickedVal)
	}
}

/*
Get whether the time spec is satisfied by val.

If Derandomize has never been called, this method will panic.
*/
func (self *RandomTimeSpec) Satisfied(val int) bool {
	if self.pickedVal == nil {
		panic("RandomTimeSpec has never been derandomized")
	}

	return *self.pickedVal == val
}

/*
	Pick a random value, and remember it so that it can be used by
	the method Satisfied.

	The method Satisfied will panic unless this method has been
	called.

	If this method has already been called, calling it again has
	no effect.
*/
func (self *RandomTimeSpec) Derandomize() {
	if self.pickedVal != nil {
		return
	}

	tmp := self.vals[rand.Intn(len(self.vals))]
	self.pickedVal = &tmp
}

/*
   Get the picked value.  If Derandomize has never been called,
   returns nil.
*/
func (self *RandomTimeSpec) PickedValue() *int {
	return self.pickedVal
}

func ParseFullTimeSpec(s string) (*FullTimeSpec, error) {
	var fullSpec FullTimeSpec
	fullSpec.Sec = &WildcardTimeSpec{}
	fullSpec.Min = &WildcardTimeSpec{}
	fullSpec.Hour = &WildcardTimeSpec{}
	fullSpec.Mday = &WildcardTimeSpec{}
	fullSpec.Mon = &WildcardTimeSpec{}
	fullSpec.Wday = &WildcardTimeSpec{}

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
		return nil, &common.Error{What: "Excess elements in 'time' field."}
	}

	return &fullSpec, nil
}
