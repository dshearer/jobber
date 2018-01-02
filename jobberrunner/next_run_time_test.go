package main

import (
	"fmt"
	"github.com/dshearer/jobber/jobfile"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TimeString(t time.Time) string {
	return t.Format("Jan _2 15:04:05 2006")
}

func myDate(year int, month time.Month, day, hour, min, sec int) time.Time {
	loc, _ := time.LoadLocation("Local")
	return time.Date(year, month, day, hour, min, sec, 0, loc)
}

type TestCase struct {
	timeSpec    string
	startTime   time.Time
	expRunTimes []time.Time
}

var TestCases []TestCase = []TestCase{
	TestCase{
		"12 5 14 * * 1",             // every Monday at 2:05:12 PM
		myDate(2016, 1, 1, 0, 0, 0), // start on 1 Jan 2016, which is a Friday
		[]time.Time{
			myDate(2016, 1, 4, 14, 5, 12),
			myDate(2016, 1, 11, 14, 5, 12),
			myDate(2016, 1, 18, 14, 5, 12),
			myDate(2016, 1, 25, 14, 5, 12),
			myDate(2016, 2, 1, 14, 5, 12),
		},
	},

	TestCase{
		"12 5 14 5 * *",             // every 5th day of month at 2:05:12 PM
		myDate(2016, 1, 1, 0, 0, 0), // start on 1 Jan 2016
		[]time.Time{
			myDate(2016, 1, 5, 14, 5, 12),
			myDate(2016, 2, 5, 14, 5, 12),
			myDate(2016, 3, 5, 14, 5, 12),
		},
	},

	TestCase{
		"12 5 14 4 * 1",             // every Monday that is on the 4th at 2:05:12 PM
		myDate(2016, 1, 1, 0, 0, 0), // start on 1 Jan 2016
		[]time.Time{
			myDate(2016, 1, 4, 14, 5, 12),
			myDate(2016, 4, 4, 14, 5, 12),
			myDate(2016, 7, 4, 14, 5, 12),
			myDate(2017, 9, 4, 14, 5, 12),
			myDate(2017, 12, 4, 14, 5, 12),
		},
	},

	TestCase{
		"", // every second
		myDate(2016, 1, 1, 5, 59, 55), // start on 1 Jan 2016
		[]time.Time{
			myDate(2016, 1, 1, 5, 59, 55),
			myDate(2016, 1, 1, 5, 59, 56),
			myDate(2016, 1, 1, 5, 59, 57),
			myDate(2016, 1, 1, 5, 59, 58),
			myDate(2016, 1, 1, 5, 59, 59),
			myDate(2016, 1, 1, 6, 0, 0),
			myDate(2016, 1, 1, 6, 0, 1),
		},
	},

	TestCase{
		"*/2 * * * * 1",             // every 2 seconds on Mondays
		myDate(2016, 1, 1, 0, 0, 0), // start on 1 Jan 2016, a Friday
		[]time.Time{
			myDate(2016, 1, 4, 0, 0, 0),
			myDate(2016, 1, 4, 0, 0, 2),
			myDate(2016, 1, 4, 0, 0, 4),
			myDate(2016, 1, 4, 0, 0, 6),
		},
	},

	TestCase{
		"0 */3 * * * 1",             // every 3 minutes on Mondays
		myDate(2016, 1, 1, 0, 0, 0), // start on 1 Jan 2016, a Friday
		[]time.Time{
			myDate(2016, 1, 4, 0, 0, 0),
			myDate(2016, 1, 4, 0, 3, 0),
			myDate(2016, 1, 4, 0, 6, 0),
			myDate(2016, 1, 4, 0, 9, 0),
		},
	},

	TestCase{
		"0 30 7 * * 1-5",            // every monday to friday at 7:30 am
		myDate(2016, 1, 1, 0, 0, 0), // start on 1 Jan 2016, a Friday
		[]time.Time{
			myDate(2016, 1, 1, 7, 30, 0),
			myDate(2016, 1, 4, 7, 30, 0),
			myDate(2016, 1, 5, 7, 30, 0),
			myDate(2016, 1, 6, 7, 30, 0),
			myDate(2016, 1, 7, 7, 30, 0),
			myDate(2016, 1, 8, 7, 30, 0),
		},
	},
}

func TestNextRunTime(t *testing.T) {

	for _, testCase := range TestCases {
		/*
		 * Set up
		 */
		var job *jobfile.Job = jobfile.NewJob("JobA", "blah", "dude")
		timeSpec, _ := jobfile.ParseFullTimeSpec(testCase.timeSpec)
		require.NotNil(t, timeSpec)
		job.FullTimeSpec = *timeSpec

		var now time.Time = testCase.startTime
		for _, expRunTime := range testCase.expRunTimes {
			fmt.Printf("time spec: %v\n", testCase.timeSpec)
			fmt.Printf("now: %v\n", TimeString(now))

			/*
			 * Call
			 */
			var actualRunTime *time.Time = nextRunTime(job, now)

			/*
			 * Test
			 */
			require.NotNil(t, actualRunTime)
			msg := fmt.Sprintf("Time spec: %v", testCase.timeSpec)
			require.Equal(t, expRunTime, *actualRunTime, msg)

			now = actualRunTime.Add(time.Second)
		}
	}
}

func TestNextRunTimeWithRandom(t *testing.T) {
	/*
	 * Set up
	 */
	var job *jobfile.Job = jobfile.NewJob("JobA", "blah", "dude")
	timeSpec, _ := jobfile.ParseFullTimeSpec("0 0 R2-4")
	require.NotNil(t, timeSpec)
	job.FullTimeSpec = *timeSpec
	job.FullTimeSpec.Derandomize()

	now := myDate(2016, 1, 1, 0, 0, 0)

	/*
	 * Call
	 */
	actualRunTime := nextRunTime(job, now)

	/*
	 * Test
	 */
	require.Contains(t, []int{2, 3, 4}, actualRunTime.Hour())
}
