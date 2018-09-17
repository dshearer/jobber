package jobfile

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func makeRange(start, end int) []int {
	arr := make([]int, 0, end-start)
	for i := start; i < end; i++ {
		arr = append(arr, i)
	}
	return arr
}

func TestParseFullTimeSpec(t *testing.T) {
	evens := []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22}
	threes := []int{1, 4, 7, 10, 13, 16, 19, 22}
	cases := []struct {
		str  string
		spec FullTimeSpec
	}{
		{"0 0 14", FullTimeSpec{
			OneValTimeSpec{0},
			OneValTimeSpec{0},
			OneValTimeSpec{14},
			WildcardTimeSpec{},
			WildcardTimeSpec{},
			WildcardTimeSpec{}}},
		{"0 0 14 * * 1", FullTimeSpec{
			OneValTimeSpec{0},
			OneValTimeSpec{0},
			OneValTimeSpec{14},
			WildcardTimeSpec{},
			WildcardTimeSpec{},
			OneValTimeSpec{1}}},
		{"0 0 */2 * * 1", FullTimeSpec{
			OneValTimeSpec{0},
			OneValTimeSpec{0},
			SetTimeSpec{"*/2", evens},
			WildcardTimeSpec{},
			WildcardTimeSpec{},
			OneValTimeSpec{1}}},
		{"0 0 1,4,7,10,13,16,19,22 * * 1", FullTimeSpec{
			OneValTimeSpec{0},
			OneValTimeSpec{0},
			SetTimeSpec{"1,4,7,10,13,16,19,22", threes},
			WildcardTimeSpec{},
			WildcardTimeSpec{},
			OneValTimeSpec{1}}},
		{"10,20,30,40 0 14 1 8 0-5", FullTimeSpec{
			SetTimeSpec{"10,20,30,40", []int{10, 20, 30, 40}},
			OneValTimeSpec{0},
			OneValTimeSpec{14},
			OneValTimeSpec{1},
			OneValTimeSpec{8},
			SetTimeSpec{"0-5", makeRange(0, 6)}}},
		{"0 0 R * * 1", FullTimeSpec{
			OneValTimeSpec{0},
			OneValTimeSpec{0},
			&RandomTimeSpec{desc: "R", vals: makeRange(0, 24)},
			WildcardTimeSpec{},
			WildcardTimeSpec{},
			OneValTimeSpec{1}}},
		{"0 0 R2-4 * * 1", FullTimeSpec{
			OneValTimeSpec{0},
			OneValTimeSpec{0},
			&RandomTimeSpec{desc: "R2-4", vals: makeRange(2, 5)},
			WildcardTimeSpec{},
			WildcardTimeSpec{},
			OneValTimeSpec{1}}},
	}

	for _, c := range cases {
		/*
		 * Call
		 */
		var result *FullTimeSpec
		var err error
		result, err = ParseFullTimeSpec(c.str)

		/*
		 * Test
		 */
		if err != nil {
			t.Fatalf("Got error: %v", err)
		}
		require.Equal(t, c.spec, *result)
	}
}
