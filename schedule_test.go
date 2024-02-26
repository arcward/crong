package crong

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"
	"time"
)

type testCase struct {
	name           string
	cron           string
	expectMinutes  []int
	expectHours    []int
	expectDays     []int
	expectMonths   []int
	expectWeekdays []int
	givenTime      time.Time
	nextTime       time.Time
	prevTime       time.Time
	includeTimes   []time.Time
	excludeTimes   []time.Time
}

func TestCronSchedule(t *testing.T) {
	testCases := []testCase{
		{
			name:           "every minute",
			cron:           "* * * * *",
			expectMinutes:  minuteOpts.Allowed,
			expectHours:    hourOpts.Allowed,
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 10, 31, 12, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 10, 31, 12, 31, 0, 0, time.UTC,
			),
			prevTime: time.Date(
				2024, 10, 31, 12, 29, 0, 0, time.UTC,
			),
		},
		{
			name: "every 2nd minute from 0 through 30",
			cron: "0-30/2 * * * *",
			expectMinutes: []int{
				0,
				2,
				4,
				6,
				8,
				10,
				12,
				14,
				16,
				18,
				20,
				22,
				24,
				26,
				28,
				30,
			},
			expectHours:    hourOpts.Allowed,
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 10, 31, 12, 23, 0, 0, time.UTC,
			),
			prevTime: time.Date(
				2024, 10, 31, 12, 22, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 10, 31, 12, 24, 0, 0, time.UTC,
			),
		},
		{
			name: "minutes 15 and 16, and every 2nd minute from 0 through 10",
			cron: "0-10/2,15,16 * * * *",
			expectMinutes: []int{
				0,
				2,
				4,
				6,
				8,
				10,
				15,
				16,
			},
			expectHours:    hourOpts.Allowed,
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 10, 31, 12, 20, 0, 0, time.UTC,
			),
			prevTime: time.Date(
				2024, 10, 31, 12, 16, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 10, 31, 13, 0, 0, 0, time.UTC,
			),
			excludeTimes: []time.Time{
				time.Date(
					2024, 10, 31, 12, 12, 0, 0, time.UTC,
				),
			},
		},
		{
			name: "daily at 00:00",
			cron: Daily,
			expectMinutes: []int{
				0,
			},
			expectHours:    []int{0},
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 10, 31, 12, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 11, 1, 0, 0, 0, 0, time.UTC,
			),
		},
		{
			name: "monthly",
			cron: Monthly,
			expectMinutes: []int{
				0,
			},
			expectHours:    []int{0},
			expectDays:     []int{1},
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2023, 11, 14, 12, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2023, 12, 1, 0, 0, 0, 0, time.UTC,
			),
		},
		{
			name:           "yearly",
			cron:           Yearly,
			expectMinutes:  []int{0},
			expectHours:    []int{0},
			expectDays:     []int{1},
			expectMonths:   []int{januaryInd},
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2023, 10, 31, 12, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 1, 1, 0, 0, 0, 0, time.UTC,
			),
		},
		{
			name:           "hourly",
			cron:           Hourly,
			expectMinutes:  []int{0},
			expectHours:    hourOpts.Allowed,
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 10, 31, 14, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 10, 31, 15, 0, 0, 0, time.UTC,
			),
		},
		{
			name:          "monday-friday",
			cron:          "0 0 * * 1-5",
			expectMinutes: []int{0},
			expectHours:   []int{0},
			expectDays:    dayOpts.Allowed,
			expectMonths:  monthOpts.Allowed,
			expectWeekdays: []int{
				mondayInd,
				tuesdayInd,
				wednesdayInd,
				thursdayInd,
				fridayInd,
			},
			givenTime: time.Date(
				2024, 2, 24, 14, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 2, 26, 0, 0, 0, 0, time.UTC,
			),
		},
		{
			name:           "at minute 30",
			cron:           "30 * * * *",
			expectMinutes:  []int{30},
			expectHours:    hourOpts.Allowed,
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 10, 31, 14, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 10, 31, 15, 30, 0, 0, time.UTC,
			),
		},
		{
			name:           "every quarter",
			cron:           "0 0 1 */3 *",
			expectMinutes:  []int{0},
			expectHours:    []int{0},
			expectDays:     []int{1},
			expectMonths:   []int{januaryInd, aprilInd, julyInd, octoberInd},
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 2, 20, 14, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 4, 1, 0, 0, 0, 0, time.UTC,
			),
		},
		{
			name:           "every even hour",
			cron:           "0 */2 * * *",
			expectMinutes:  []int{0},
			expectHours:    []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22},
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 10, 31, 15, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 10, 31, 16, 0, 0, 0, time.UTC,
			),
		},
		{
			name:          "every 15th minute past every hour from 9-16 on every day of week from monday-friday",
			cron:          "*/15 9-17 * * MON-FRI",
			expectMinutes: []int{0, 15, 30, 45},
			expectHours:   []int{9, 10, 11, 12, 13, 14, 15, 16, 17},
			expectDays:    dayOpts.Allowed,
			expectMonths:  monthOpts.Allowed,
			expectWeekdays: []int{
				mondayInd,
				tuesdayInd,
				wednesdayInd,
				thursdayInd,
				fridayInd,
			},
			givenTime: time.Date(
				2024, 2, 23, 20, 35, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 2, 26, 9, 0, 0, 0, time.UTC,
			),
		},
		{
			name:          "every 15th minute past every hour from 9-16 on every day of week from monday-friday",
			cron:          "*/15 9-17 * * MON-FRI",
			expectMinutes: []int{0, 15, 30, 45},
			expectHours:   []int{9, 10, 11, 12, 13, 14, 15, 16, 17},
			expectDays:    dayOpts.Allowed,
			expectMonths:  monthOpts.Allowed,
			expectWeekdays: []int{
				mondayInd,
				tuesdayInd,
				wednesdayInd,
				thursdayInd,
				fridayInd,
			},
			givenTime: time.Date(
				2024, 2, 23, 10, 35, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 2, 23, 10, 45, 0, 0, time.UTC,
			),
		},
		{
			name:           "leap year",
			cron:           "0 0 29 2 *",
			expectMinutes:  []int{0},
			expectHours:    []int{0},
			expectDays:     []int{29},
			expectMonths:   []int{februaryInd},
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2023, 2, 23, 10, 35, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 2, 29, 0, 0, 0, 0, time.UTC,
			),
		},
		{
			name:           "minute 0 and 30 past hour 14 and 18 on sunday and friday",
			cron:           "0,30 14,18 * * 0,5",
			expectMinutes:  []int{0, 30},
			expectHours:    []int{14, 18},
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: []int{sundayInd, fridayInd},
			givenTime: time.Date(
				2024, 2, 20, 10, 35, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 2, 23, 14, 0, 0, 0, time.UTC,
			),
		},
		{
			name:           "minute 0 and 30 past hour 14 and 18 on sunday and friday",
			cron:           "0,30 14,18 * * 0,5",
			expectMinutes:  []int{0, 30},
			expectHours:    []int{14, 18},
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: []int{sundayInd, fridayInd},
			givenTime: time.Date(
				2024, 2, 23, 14, 0, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 2, 23, 14, 30, 0, 0, time.UTC,
			),
		},
		{
			name:           "minute 0 and 30 past hour 14 and 18 on sunday and friday",
			cron:           "0,30 14,18 * * 0,5",
			expectMinutes:  []int{0, 30},
			expectHours:    []int{14, 18},
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: []int{sundayInd, fridayInd},
			givenTime: time.Date(
				2024, 2, 23, 18, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 2, 25, 14, 0, 0, 0, time.UTC,
			),
		},
		{
			name:           "last day of month",
			cron:           "* * L * *",
			expectMinutes:  minuteOpts.Allowed,
			expectHours:    hourOpts.Allowed,
			expectDays:     []int{},
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 2, 23, 18, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 2, 29, 0, 0, 0, 0, time.UTC,
			),
		},
		{
			name:           "every 2nd minute from min10-20, past every 2nd hour from 10-20, on every 2nd day of month through 20, in every 2nd month from Feb-Aug",
			cron:           "10-20/2 10-20/2 10-20/2 2-8/2 *",
			expectMinutes:  []int{10, 12, 14, 16, 18, 20},
			expectHours:    []int{10, 12, 14, 16, 18, 20},
			expectDays:     []int{10, 12, 14, 16, 18, 20},
			expectMonths:   []int{februaryInd, aprilInd, juneInd, augustInd},
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 5, 23, 18, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 6, 10, 10, 10, 0, 0, time.UTC,
			),
		},
		{
			name: "every min from 20-25 past every hour from 1-3 on " +
				"every day of month from 5-8 and every day of week from mon-tues " +
				"in every month from feb-apr",
			cron:           "20-25 1-3 5-8 2-4 1-2",
			expectMinutes:  []int{20, 21, 22, 23, 24, 25},
			expectHours:    []int{1, 2, 3},
			expectDays:     []int{5, 6, 7, 8},
			expectMonths:   []int{februaryInd, marchInd, aprilInd},
			expectWeekdays: []int{mondayInd, tuesdayInd},
			givenTime: time.Date(
				2024, 2, 20, 18, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 3, 5, 1, 20, 0, 0, time.UTC,
			),
			includeTimes: []time.Time{
				time.Date(
					2024, 2, 5, 1, 20, 0, 0, time.UTC,
				),
				time.Date(
					2024, 3, 5, 2, 24, 0, 0, time.UTC,
				),
			},
		},
		{
			name:           "at min 20 and 21 past hour 15 and 16 on months 10 and 11 in oct and nov",
			cron:           "20-25 15-16 5-8 10-11 1-2",
			expectMinutes:  []int{20, 21, 22, 23, 24, 25},
			expectHours:    []int{15, 16},
			expectDays:     []int{5, 6, 7, 8},
			expectMonths:   []int{octoberInd, novemberInd},
			expectWeekdays: []int{mondayInd, tuesdayInd},
			givenTime:      time.Time{},
			nextTime:       time.Time{},
			includeTimes: []time.Time{
				time.Date(
					2026, 10, 5, 15, 20, 0, 0, time.UTC,
				),
				time.Date(
					2026, 10, 6, 15, 21, 0, 0, time.UTC,
				),
				time.Date(
					2024, 11, 5, 16, 20, 0, 0, time.UTC,
				),
			},
		},
		{
			name:           "every 10th minute from 5 through 59 past hour 4 and 5",
			cron:           "5/10 4,5 * * *",
			expectMinutes:  []int{5, 15, 25, 35, 45, 55},
			expectHours:    []int{4, 5},
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 2, 21, 11, 30, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 2, 22, 4, 5, 0, 0, time.UTC,
			),
			includeTimes: []time.Time{
				time.Date(
					2024, 2, 22, 4, 15, 0, 0, time.UTC,
				),
				time.Date(
					2024, 2, 22, 4, 25, 0, 0, time.UTC,
				),
				time.Date(
					2024, 2, 22, 4, 35, 0, 0, time.UTC,
				),
			},
		},
		{
			name:           "every 10th minute from 3 through 30 past hour 18",
			cron:           "3-30/10 18 * * *",
			expectMinutes:  []int{3, 13, 23},
			expectHours:    []int{18},
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 2, 21, 11, 15, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 2, 21, 18, 3, 0, 0, time.UTC,
			),
			includeTimes: []time.Time{
				time.Date(
					2024, 2, 21, 18, 13, 0, 0, time.UTC,
				),
				time.Date(
					2024, 2, 21, 18, 23, 0, 0, time.UTC,
				),
				time.Date(
					2024, 2, 22, 18, 3, 0, 0, time.UTC,
				),
			},
		},
		{
			name: "every 10th minute from 5-59 past every 5th hour " +
				"from 4 through 23 on every 10th day-of-month from 5 through " +
				"31 in every 3rd month from march through december",
			cron:           "5/10 4/5 5/10 3/3 *",
			expectMinutes:  []int{5, 15, 25, 35, 45, 55},
			expectHours:    []int{4, 9, 14, 19},
			expectDays:     []int{5, 15, 25},
			expectMonths:   []int{marchInd, juneInd, septemberInd, decemberInd},
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 2, 21, 11, 35, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 3, 5, 4, 5, 0, 0, time.UTC,
			),
			includeTimes: []time.Time{
				time.Date(
					2024, 3, 5, 4, 15, 0, 0, time.UTC,
				),
				time.Date(
					2024, 6, 15, 14, 45, 0, 0, time.UTC,
				),
			},
		},
		{
			name:           "12:30 on friday and saturday",
			cron:           "30 12 * * FRI,SAT",
			expectMinutes:  []int{30},
			expectHours:    []int{12},
			expectDays:     dayOpts.Allowed,
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: []int{fridayInd, saturdayInd},
			givenTime: time.Date(
				2024, 2, 21, 11, 35, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 2, 23, 12, 30, 0, 0, time.UTC,
			),
			includeTimes: []time.Time{
				time.Date(
					2024, 2, 24, 12, 30, 0, 0, time.UTC,
				),
				time.Date(
					2024, 3, 1, 12, 30, 0, 0, time.UTC,
				),
			},
		},
		{
			name:           "12:30 on the last day of every month",
			cron:           "30 12 L * *",
			expectMinutes:  []int{30},
			expectHours:    []int{12},
			expectDays:     []int{},
			expectMonths:   monthOpts.Allowed,
			expectWeekdays: weekdayOpts.Allowed,
			givenTime: time.Date(
				2024, 2, 21, 11, 35, 0, 0, time.UTC,
			),
			nextTime: time.Date(
				2024, 2, 29, 12, 30, 0, 0, time.UTC,
			),
			includeTimes: []time.Time{

				time.Date(
					2024, 3, 31, 12, 30, 0, 0, time.UTC,
				),

				time.Date(
					2024, 4, 30, 12, 30, 0, 0, time.UTC,
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("%s [%s]", tc.name, tc.cron), func(t *testing.T) {
				t.Parallel()
				slices.Sort(tc.expectMonths)
				slices.Sort(tc.expectDays)
				slices.Sort(tc.expectHours)
				slices.Sort(tc.expectMinutes)
				slices.Sort(tc.expectWeekdays)
				s, err := New(tc.cron, nil)
				if err != nil {
					t.Fatalf("unexpected error with '%s': %s", tc.cron, err)
				}
				scheduleTest(t, tc, s)
				checkTimes(t, tc, s)
			},
		)
	}
}

func TestInvalidCron(t *testing.T) {
	s, err := New("* * */42 * *", nil)
	requireErr(
		t,
		err,
		fmt.Sprintf("got schedule: %v", s),
		fmt.Sprintf("minutes: %v", s.minutes),
		fmt.Sprintf("hours: %v", s.hours),
		fmt.Sprintf("days: %v", s.days),
		fmt.Sprintf("months: %v", s.months),
		fmt.Sprintf("weekdays: %v", s.weekdays),
	)
}

func checkTimes(t *testing.T, tc testCase, s *Schedule) {
	t.Helper()
	var testCases []time.Time

	for _, month := range tc.expectMonths {
		// L is allowed as a day value, so if we just see
		// L, the day slice should be empty, and we can
		// temporarily replace it here
		if s.Day() == string(Last) {
			targetMonth := month + 1
			nextMonth := time.Date(
				2024,
				time.Month(targetMonth),
				1,
				0,
				0,
				0,
				0,
				time.UTC,
			)
			lastOfMonth := nextMonth.Add(-time.Hour).Day()
			tc.expectDays = []int{lastOfMonth}
		}

		for _, day := range tc.expectDays {
			for _, hour := range tc.expectHours {
				for _, minute := range tc.expectMinutes {
					dt := time.Date(
						2024,
						time.Month(month),
						day,
						hour,
						minute,
						0,
						0,
						time.UTC,
					)
					// accounts for the day entry exceeding the
					// number of days in the current month
					if int(dt.Month()) != month {
						break
					}
					for _, weekday := range tc.expectWeekdays {
						if int(dt.Weekday()) == weekday {
							testCases = append(testCases, dt)
							break
						}
					}
				}
			}
		}
	}
	if testCases == nil {
		t.Errorf(
			"expected permutations, got none for %s (%s)",
			tc.name,
			s.String(),
		)
	}

	// currentSchedule := testCases[0]
	// if !s.Matches(currentSchedule) {
	// 	t.Fatalf(
	// 		"expected %s to match %s",
	// 		currentSchedule,
	// 		s.String(),
	// 	)
	// }

	for i := 0; i < len(testCases); i++ {
		if i+1 == len(testCases) {
			break
		}
		currentSchedule := testCases[i]
		nextSchedule := testCases[i+1]
		if nextSchedule.Before(currentSchedule) {
			t.Fatalf(
				"index %d is before index %d: %s -> %s (surrounded by: index %d (%s) "+
					"and index %d (%s) (months: %d days: %d hours: %d minutes: %d weekdays: %d))",
				i+1,
				i,
				currentSchedule,
				nextSchedule,
				i-1,
				testCases[i-1],
				i+2,
				testCases[i+2],
				len(tc.expectMonths),
				len(tc.expectDays),
				len(tc.expectHours),
				len(tc.expectMinutes),
				len(tc.expectWeekdays),
			)
		}
		n := s.Next(currentSchedule)
		if !n.Equal(nextSchedule) {
			t.Fatalf(
				"cron: %s\nexpected on index %d/%d:\nfrom: %s\nto:   %s\ngot:  %s",
				s.String(),
				i,
				len(testCases),
				currentSchedule,
				nextSchedule,
				n,
			)
		}
	}
}

func TestEmptyCron(t *testing.T) {
	_, err := New("", nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestErrors(t *testing.T) {
	type errorCase struct {
		Name string
		Cron string
	}
	testCases := []errorCase{
		{Name: "too many fields", Cron: "0 0 1 1 1 1"},
		{Name: "60 minutes", Cron: "60 * * * *"},
		{Name: "25 hours", Cron: "* 25 * * *"},
		{Name: "32 days", Cron: "* * 32 * *"},
		{Name: "13 months", Cron: "* * * 13 *"},
		{Name: "negative minute", Cron: "-1 * * * *"},
		{Name: "negative hours", Cron: "* -1 * * *"},
		{Name: "negative days", Cron: "* * -1 * *"},
		{Name: "negative months", Cron: "* * * -1 *"},
		{Name: "invalid minutes", Cron: "wat * * * *"},
		{Name: "invalid hours", Cron: "* wat * * *"},
		{Name: "invalid days", Cron: "* * wat * *"},
		{Name: "invalid months", Cron: "* * * wat *"},
		{Name: "invalid weekday", Cron: "* * * * wat"},
		{Name: "invalid minute step", Cron: "*/wat * * * *"},
		{Name: "invalid hour step", Cron: "* */ * * *"},
		{Name: "invalid blank identifier", Cron: "* ? * * *"},
		{Name: "invalid minute range", Cron: "5- * * * *"},
		{Name: "invalid minute steps", Cron: "/5 * * * *"},
		{Name: "invalid minute range", Cron: "*/-1 * * * *"},
		{Name: "invalid minute post-range", Cron: "5/60 * * * *"},
		{Name: "invalid minute pre-range", Cron: "60/5 * * * *"},
		{Name: "empty cron", Cron: ""},
		{Name: "blank cron", Cron: " "},
		{Name: "minute range backwards", Cron: "30-20 * * * *"},
		{Name: "minute range Start wildcard", Cron: "*-20 * * * *"},
		{Name: "minute range Start invalid", Cron: "F-20 * * * *"},
		{Name: "minute range end wildcard", Cron: "2-* * * * *"},
		{Name: "empty minute end range", Cron: "2- * * * *"},
		{Name: "empty minute Start range", Cron: "L-2 * * * *"},
		{Name: "should not be every 4 hours", Cron: "*/240 * * * *"},
		{Name: "zero step", Cron: "*/0 * * * *"},
	}

	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("%s (%s)", tc.Name, tc.Cron), func(t *testing.T) {
				_, err := New(tc.Cron, nil)
				if err == nil {
					t.Fatalf("expected error for %s", tc.Cron)
				}
			},
		)
	}
}

func TestParse(t *testing.T) {
	type parseCase struct {
		Name  string
		Value string
	}
	testCases := []parseCase{
		{
			Name:  "empty string",
			Value: "",
		},
		{
			Name:  "bad range",
			Value: "-2",
		},
		{
			Name:  "bad range",
			Value: "1-UH",
		},
		{
			Name:  "L in range",
			Value: "1-L",
		},
		{
			Name:  "bad mine",
			Value: "28/27",
		},
		{
			Name:  "bad list",
			Value: "1,2,32",
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.Name, func(t *testing.T) {
				indexes, err := dayOpts.parse(tc.Value)
				if err == nil {
					t.Fatalf("expected error (got days: %#v)", indexes)
				}
			},
		)
	}
}

func TestUntilNext(t *testing.T) {
	s, err := New(Hourly, nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	dt := time.Date(2024, 2, 21, 11, 35, 0, 0, time.UTC)
	next := s.UntilNext(dt)
	if next != 25*time.Minute {
		t.Fatalf("expected next time")
	}
}

func scheduleTest(t *testing.T, tc testCase, s *Schedule) {
	t.Helper()
	// t.Parallel()

	expectCron, ok := cronShortcut[tc.cron]
	if !ok {
		expectCron = tc.cron
	}
	assertEqual(t, s.String(), expectCron)

	if s.allowAnyMinute {
		assertEqual(t, len(s.minutes), 0)
	} else if !slicesEqual(t, s.minutes, tc.expectMinutes) {
		t.Errorf("(minutes) expected %v, got %v", tc.expectMinutes, s.minutes)
	}

	if slicesEqual(t, tc.expectHours, hourOpts.Allowed) {
		assertEqual(t, len(s.hours), 0)
	} else if !slicesEqual(t, s.hours, tc.expectHours) {
		t.Errorf("(hours) expected %v, got %v", tc.expectHours, s.hours)
	}

	if s.allowAnyDay {
		assertEqual(t, len(s.days), 0)
	} else if !slicesEqual(t, s.days, tc.expectDays) {
		t.Errorf("(days) expected %v, got %v", tc.expectDays, s.days)
	}

	if s.allowAnyMonth {
		assertEqual(t, len(s.months), 0)
	} else if !slicesEqual(t, s.months, tc.expectMonths) {
		t.Errorf("(months) expected %v, got %v", tc.expectMonths, s.months)
	}

	if s.allowAnyWeekday {
		assertEqual(t, len(s.weekdays), 0)
	} else if !slicesEqual(t, s.weekdays, tc.expectWeekdays) {
		t.Logf(
			"compared (%v):\n%#v\n%#v\n%#v",
			s.allowAnyWeekday,
			s.weekdays,
			tc.expectWeekdays,
			weekdayOpts.Allowed,
		)
		t.Errorf(
			"(weekdays) expected %v, got %v",
			tc.expectWeekdays,
			s.weekdays,
		)
	}

	if !tc.givenTime.IsZero() && !tc.nextTime.IsZero() {
		givenTime := tc.givenTime
		nextTime := tc.nextTime
		n := s.Next(givenTime)
		// t.Logf("given: %s, next: %s expected: %s", givenTime, n, nextTime)
		assertEqual(t, nextTime, n)
	}

	if !tc.prevTime.IsZero() {
		prevTime := tc.prevTime
		assertEqual(t, prevTime, s.Prev(tc.givenTime))
	}

	if tc.includeTimes != nil {
		for _, givenTime := range tc.includeTimes {
			matches := s.Matches(givenTime)
			if !matches {
				t.Errorf(
					"expected %s to match %s",
					givenTime,
					tc.cron,
				)
			}
		}
	}

	if tc.excludeTimes != nil {
		for _, givenTime := range tc.excludeTimes {
			if s.Matches(givenTime) {
				t.Errorf(
					"expected %s to match %s",
					givenTime,
					tc.cron,
				)
			}
		}
	}

}

func BenchmarkSchedule(b *testing.B) {
	schedules := []string{}

	for i := 0; i < b.N; i++ {
		cronExpr, err := NewRandom(rand.New(rand.NewSource(int64(i))))
		if err != nil {
			b.Fatalf("unexpected error: %s", err)
		}
		schedules = append(schedules, cronExpr)
	}

	b.ResetTimer()

	for i := 0; i < len(schedules); i++ {
		cronExpr := schedules[i]
		s, err := New(cronExpr, nil)
		if err != nil {
			b.Fatalf("unexpected error: %s", err)
		}
		next := s.Next(time.Now())
		b.Logf("next time for (%s): %s", s, next)
	}
}

func BenchmarkRandomSchedules(b *testing.B) {
	schedules := []string{}

	for i := 0; i < b.N; i++ {
		cronExpr, err := NewRandom(rand.New(rand.NewSource(int64(i))))
		if err != nil {
			b.Fatalf("unexpected error: %s", err)
		}
		schedules = append(schedules, cronExpr)
	}

	b.ResetTimer()

	for i := 0; i < len(schedules); i++ {
		cronExpr := schedules[i]
		s, err := New(cronExpr, nil)
		if err != nil {
			b.Fatalf("unexpected error: %s", err)
		}
		_ = s.Next(time.Now())
		// b.Logf("next time for (%s): %s", s, next)
	}
}

func BenchmarkScheduleNext(b *testing.B) {
	// cronExpr, err := NewRandom(rand.New(rand.NewSource(int64(1))))
	// if err != nil {
	// 	b.Fatalf("unexpected error: %s", err)
	// }
	s, err := New(Hourly, nil)
	if err != nil {
		b.Fatalf("unexpected error: %s", err)
	}
	next := time.Now()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		next = s.Next(next)
	}
}

func FuzzSchedule(f *testing.F) {
	for i := range 500 {
		f.Add(int64(i))
	}
	f.Fuzz(
		func(t *testing.T, seed int64) {
			rc, err := NewRandom(rand.New(rand.NewSource(seed)))
			if err != nil {
				t.Fatalf("failed on %d: %s", seed, err)
			}
			t.Logf("%d: %s", seed, rc)

			schedule, scheduleErr := New(rc, nil)
			if scheduleErr != nil {
				t.Fatalf(
					"unexpected error on %d (%s): %s",
					seed,
					rc,
					scheduleErr,
				)
			}
			next := schedule.Next(time.Now())
			t.Logf("%d (%s) next: %s", seed, rc, next)

		},
	)
}

func TestParseRange(t *testing.T) {
	type rangeCase struct {
		Before      string
		After       string
		ExpectRange []int
		ExpectError bool
	}
	cases := []rangeCase{
		{
			Before:      "1",
			After:       "5",
			ExpectRange: []int{1, 2, 3, 4, 5},
		},
		{
			Before:      "0",
			After:       "5",
			ExpectError: true,
		},
		{
			Before:      "5",
			After:       "5",
			ExpectError: true,
		},
		{
			Before:      "5",
			After:       "32",
			ExpectError: true,
		},
		{
			Before:      "10",
			After:       "5",
			ExpectError: true,
		},
		{
			Before:      "1",
			After:       "31",
			ExpectRange: dayOpts.Allowed,
		},
	}

	for _, tc := range cases {
		t.Run(
			fmt.Sprintf("%s-%s", tc.Before, tc.After), func(t *testing.T) {
				r, err := dayOpts.parseRange(tc.Before, tc.After)

				if tc.ExpectError {
					if err == nil {
						t.Errorf("expected error")
					}
				} else if err != nil {
					t.Errorf("unexpected error: %s", err)
				} else if !slicesEqual(t, r, tc.ExpectRange) {
					t.Errorf("expected %v, got %v", tc.ExpectRange, r)
				}
			},
		)
	}
}

func TestParseStep(t *testing.T) {
	type stepCase struct {
		Before      string
		After       string
		ExpectSteps []int
		ExpectError bool
	}
	cases := []stepCase{
		{
			Before:      "5",
			After:       "6",
			ExpectSteps: []int{5, 11, 17, 23, 29},
		},
		{
			Before:      "5-30",
			After:       "4",
			ExpectSteps: []int{5, 9, 13, 17, 21, 25, 29},
		},
		{
			Before:      "",
			After:       "",
			ExpectError: true,
		},
		{
			Before:      "-1",
			After:       "10",
			ExpectError: true,
		},
		{
			Before:      "32",
			After:       "10",
			ExpectError: true,
		},
		{
			Before:      "45-10",
			After:       "10",
			ExpectError: true,
		},
		{
			Before:      "27-28",
			After:       "29",
			ExpectError: true,
		},
		{
			Before:      "1-31",
			After:       "15",
			ExpectSteps: []int{1, 16, 31},
		},
		{
			Before:      "30",
			After:       "31",
			ExpectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(
			fmt.Sprintf("%s-%s", tc.Before, tc.After), func(t *testing.T) {
				r, err := dayOpts.parseStep(tc.Before, tc.After)

				if tc.ExpectError {
					if err == nil {
						t.Errorf("expected error")
					} else {
						t.Logf("got error: %s", err)
					}
				} else if err != nil {
					t.Errorf("unexpected error: %s", err)
				} else if !slicesEqual(t, r, tc.ExpectSteps) {
					t.Errorf("expected %v, got %v", tc.ExpectSteps, r)
				}
			},
		)
	}
}

func TestNewRandom(t *testing.T) {
	r := rand.New(rand.NewSource(1))

	// Macros are returned when a 1 is selected at random from
	// between 0-100. There's a 99.9952% chance we'll see that 1,
	// and a 0.0048% chance this test will fail from now seeing a 1.
	macroSeen := false
	for i := 0; i < 1000; i++ {
		r.Seed(int64(i))
		cronExpr, err := NewRandom(r)
		if err != nil {
			t.Fatalf("unexpected error on %d: %s", i, err)
		}
		_, err = New(cronExpr, nil)
		if err != nil {
			t.Errorf("unexpected error on %d (%s): %s", i, cronExpr, err)
		}
		if !macroSeen {
			for _, m := range macros {
				if cronExpr == m {
					t.Logf("saw macro %s after %d attempts", cronExpr, i)
					macroSeen = true
					break
				}
			}
		}
	}
	if !macroSeen {
		t.Fatalf("didn't see macro schedule")
	}
}
