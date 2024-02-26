package crong

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	// Cron special characters

	Any           = '*'
	ListSeparator = ','
	Range         = '-'
	Step          = '/'
	Blank         = '?'
	Last          = 'L'

	// Cron macros

	Yearly   = "@yearly"
	Annually = "@annually"
	Monthly  = "@monthly"
	Weekly   = "@weekly"
	Daily    = "@daily"
	Midnight = "@midnight"
	Hourly   = "@hourly"

	// String representations for weekdays

	Sunday    = "SUN"
	Monday    = "MON"
	Tuesday   = "TUE"
	Wednesday = "WED"
	Thursday  = "THU"
	Friday    = "FRI"
	Saturday  = "SAT"

	// String representations for months

	January   = "JAN"
	February  = "FEB"
	March     = "MAR"
	April     = "APR"
	May       = "MAY"
	June      = "JUN"
	July      = "JUL"
	August    = "AUG"
	September = "SEP"
	October   = "OCT"
	November  = "NOV"
	December  = "DEC"
)

// cron expression positions
const (
	minuteInd int = iota
	hourInd
	dayInd
	monthInd
	weekdayInd
)

// weekday indices
const (
	sundayInd int = iota
	mondayInd
	tuesdayInd
	wednesdayInd
	thursdayInd
	fridayInd
	saturdayInd
)

// month indices
const (
	januaryInd int = iota + 1
	februaryInd
	marchInd
	aprilInd
	mayInd
	juneInd
	julyInd
	augustInd
	septemberInd
	octoberInd
	novemberInd
	decemberInd
)

var (
	macros = []string{
		Yearly,
		Annually,
		Monthly,
		Weekly,
		Daily,
		Midnight,
		Hourly,
	}
	minuteOpts = field{
		Name:  "minute",
		Index: minuteInd,
		Allowed: []int{
			0,
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
			9,
			10,
			11,
			12,
			13,
			14,
			15,
			16,
			17,
			18,
			19,
			20,
			21,
			22,
			23,
			24,
			25,
			26,
			27,
			28,
			29,
			30,
			31,
			32,
			33,
			34,
			35,
			36,
			37,
			38,
			39,
			40,
			41,
			42,
			43,
			44,
			45,
			46,
			47,
			48,
			49,
			50,
			51,
			52,
			53,
			54,
			55,
			56,
			57,
			58,
			59,
		},
	}
	hourOpts = field{
		Name:  "hour",
		Index: hourInd,
		Allowed: []int{
			0,
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
			9,
			10,
			11,
			12,
			13,
			14,
			15,
			16,
			17,
			18,
			19,
			20,
			21,
			22,
			23,
		},
	}
	dayOpts = field{
		Name:  "day",
		Index: dayInd,
		Allowed: []int{
			1,
			2,
			3,
			4,
			5,
			6,
			7,
			8,
			9,
			10,
			11,
			12,
			13,
			14,
			15,
			16,
			17,
			18,
			19,
			20,
			21,
			22,
			23,
			24,
			25,
			26,
			27,
			28,
			29,
			30,
			31,
		},
	}
	monthOpts = field{
		Name:  "month",
		Index: monthInd,
		Allowed: []int{
			januaryInd,
			februaryInd,
			marchInd,
			aprilInd,
			mayInd,
			juneInd,
			julyInd,
			augustInd,
			septemberInd,
			octoberInd,
			novemberInd,
			decemberInd,
		},
		Conversions: map[string]int{
			January:   januaryInd,
			February:  februaryInd,
			March:     marchInd,
			April:     aprilInd,
			May:       mayInd,
			June:      juneInd,
			July:      julyInd,
			August:    augustInd,
			September: septemberInd,
			October:   octoberInd,
			November:  novemberInd,
			December:  decemberInd,
		},
	}
	weekdayOpts = field{
		Name:  "weekday",
		Index: weekdayInd,
		Allowed: []int{
			sundayInd,
			mondayInd,
			tuesdayInd,
			wednesdayInd,
			thursdayInd,
			fridayInd,
			saturdayInd,
		},
		Conversions: map[string]int{
			Sunday:    sundayInd,
			Monday:    mondayInd,
			Tuesday:   tuesdayInd,
			Wednesday: wednesdayInd,
			Thursday:  thursdayInd,
			Friday:    fridayInd,
			Saturday:  saturdayInd,
		},
	}
	// cronShortcut is a map of cron macros to their
	// respective cron expressions
	cronShortcut = map[string]string{
		Yearly:   "0 0 1 1 *",
		Annually: "0 0 1 1 *",
		Monthly:  "0 0 1 * *",
		Weekly:   "0 0 * * 0",
		Daily:    "0 0 * * *",
		Midnight: "0 0 * * *",
		Hourly:   "0 * * * *",
	}
)

// Schedule is a cron schedule created from a cron expression
//
// # Usage
//
// To create a new Schedule, use the New function:
//
//	  	s, err := crong.New("0 0 * * *", time.UTC)
//		if err != nil {
//			log.Fatal(err)
//		}
//
// To get the next scheduled time after a given time, use the Next method:
//
//	next := s.Next(time.Now())
//
// To check if a time matches the schedule, use the Matches method:
//
//	if s.Matches(time.Now()) {
//		fmt.Println("It's time!")
//	}
type Schedule struct {
	// values holds the parsed cron expression
	values [5]string

	// loc is the timezone/location to use
	loc *time.Location

	// created is the time this cron schedule was initialized
	created time.Time

	// minute is the string value of the minute field
	minute string
	// minutes is the parsed values of the minute field
	minutes     []int
	minutesDesc []int
	// allowAnyMinute indicates a wildcard minute
	allowAnyMinute bool

	// hour is the string value of the hour field
	hour string
	// hours is the parsed values of the hour field
	hours []int
	// allowAnyHour indicates a wildcard hour
	allowAnyHour bool

	// day is the string value of the day field
	day string
	// days is the parsed values of the day field
	days []int
	// allowAnyDay indicates a wildcard day
	allowAnyDay bool

	// month is the string value of the month field
	month string
	// months is the parsed values of the month field
	months []int
	// allowAnyMonth indicates a wildcard month
	allowAnyMonth bool

	// weekday is the string value of the weekday field
	weekday string
	// weekdays is the parsed values of the weekday field
	weekdays []int
	// allowAnyWeekday indicates a wildcard weekday
	allowAnyWeekday bool
}

// New creates a new Schedule from a cron expression. loc is the
// location to use for the schedule (if nil, defaults to time.UTC)
func New(cron string, loc *time.Location) (*Schedule, error) {
	if loc == nil {
		loc = time.UTC
	}

	s := &Schedule{values: [5]string{}, loc: loc}
	s.created = time.Now().In(s.loc)
	cron = strings.TrimSpace(cron)
	cs, ok := cronShortcut[cron]
	if ok {
		cron = cs
	}

	values := strings.Split(cron, " ")
	if len(values) != 5 {
		return nil, fmt.Errorf(
			"invalid cron schedule '%s' (expected 5 values, got %d): %s",
			cron,
			len(values),
			cron,
		)
	}
	for i, v := range values {
		s.values[i] = v
	}

	err := s.validate()
	return s, err
}

// NewRandom creates a new Schedule with a random cron expression
func NewRandom(r *rand.Rand) (string, error) {
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	}

	m := r.Intn(100)
	if m == 1 {
		return macros[r.Intn(len(macros))], nil
	}

	cronFields := make([]string, 5)

	errs := []error{}

	minuteVal, err := minuteOpts.random(r)
	errs = append(errs, err)
	cronFields[minuteInd] = minuteVal

	hourVal, err := hourOpts.random(r)
	errs = append(errs, err)
	cronFields[hourInd] = hourVal

	dayVal, err := dayOpts.random(r)
	errs = append(errs, err)
	cronFields[dayInd] = dayVal

	monthVal, err := monthOpts.random(r)
	errs = append(errs, err)
	cronFields[monthInd] = monthVal

	weekdayVal, err := weekdayOpts.random(r)
	errs = append(errs, err)
	cronFields[weekdayInd] = weekdayVal

	return strings.Join(cronFields, " "), errors.Join(errs...)
}

// Next returns the next scheduled time after the given time
func (s *Schedule) Next(t time.Time) time.Time {
	return s.nextNoTruncate(t.In(s.loc).Truncate(time.Minute))
}

// Prev returns the previous scheduled time before the given time
func (s *Schedule) Prev(t time.Time) time.Time {
	t = t.In(s.loc).Truncate(time.Minute)
	for {
		t = t.Add(-time.Minute)
		if s.Matches(t) {
			return t
		}
	}
}

// nextNoTruncate does the same thing as Next, but assumes
// that the given time had already been truncated to the minute
// and does not truncate it again
func (s *Schedule) nextNoTruncate(t time.Time) time.Time {
	// Given we already know all the months/days/weekdays/hours/minutes
	// in the schedule, there's probably a more efficient or clever
	// way to do a lot of this. For now, I'll stick to checking
	// for the 'heavy' calls (@yearly, @monthly), and increment
	// the rest. Adding a minute is a cheap operation, and
	// I feel like hourly/daily schedules are probably the
	// most common

	switch cronExpr := s.String(); cronExpr {
	case cronShortcut[Yearly]:
		// if the schedule is yearly, we can just add a year
		// to the given time and return it
		return time.Date(
			t.Year()+1,
			time.Month(januaryInd),
			1,
			0,
			0,
			0,
			0,
			t.Location(),
		)
	case Monthly:
		if int(t.Month()) == decemberInd {
			return time.Date(
				t.Year()+1,
				time.Month(januaryInd),
				1,
				0,
				0,
				0,
				0,
				t.Location(),
			)
		}
		return time.Date(
			t.Year(),
			t.Month()+1,
			1,
			0,
			0,
			0,
			0,
			t.Location(),
		)
	}

	// if s.allowAnyMonth {
	// 	maxMonth = decemberInd
	// 	minMonth = januaryInd
	// } else {
	// 	maxMonth = slices.Max(s.months)
	// 	minMonth = slices.Min(s.months)
	// }

	for {
		// if !s.isMonth(t) {
		// 	currentMonth := int(t.Month())
		// 	if currentMonth < minMonth {
		// 		t = time.Date(
		// 			t.Year(),
		// 			time.Month(minMonth),
		// 			1,
		// 			0,
		// 			0,
		// 			0,
		// 			0,
		// 			t.Location(),
		// 		)
		// 	} else if currentMonth > maxMonth {
		// 		t = time.Date(
		// 			t.Year()+1,
		// 			time.Month(minMonth),
		// 			1,
		// 			0,
		// 			0,
		// 			0,
		// 			0,
		// 			t.Location(),
		// 		)
		// 	} else {
		// 		var foundMonth int
		// 		for _, m := range s.months {
		// 			if m > currentMonth {
		// 				foundMonth = m
		// 				break
		//
		// 			}
		// 		}
		// 		if foundMonth == 0 {
		// 			panic("couldn't find month")
		// 		}
		// 		t = time.Date(
		// 			t.Year(),
		// 			time.Month(foundMonth),
		// 			1,
		// 			0,
		// 			0,
		// 			0,
		// 			0,
		// 			t.Location(),
		// 		)s
		// 	}
		// }

		t = t.Add(time.Minute)
		if s.Matches(t) {
			return t
		}
	}
}

// UntilNext returns the duration until the next scheduled time
// after the given time
func (s *Schedule) UntilNext(t time.Time) time.Duration {
	return s.Next(t).Sub(t)
}

// Matches returns true if the schedule matches the given time
func (s *Schedule) Matches(t time.Time) bool {
	// return s.isMinute(t) && s.isHour(t) && s.isDay(t) && s.isMonth(t) && s.isWeekday(t)
	return s.isWeekday(t) && s.isMonth(t) && s.isDay(t) && s.isHour(t) && s.isMinute(t)
}

// String returns the string representation of the schedule
func (s *Schedule) String() string {
	return strings.Join(s.values[:], " ")
}

// Minute returns the minute value of the schedule
func (s *Schedule) Minute() string {
	return s.values[minuteInd]
}

// Hour returns the hour value of the schedule
func (s *Schedule) Hour() string {
	return s.values[hourInd]
}

// Day returns the day value of the schedule
func (s *Schedule) Day() string {
	return s.values[dayInd]
}

// Month returns the month value of the schedule
func (s *Schedule) Month() string {
	return s.values[monthInd]
}

// Weekday returns the weekday value of the schedule
func (s *Schedule) Weekday() string {
	return s.values[weekdayInd]
}

func (s *Schedule) LogValue() slog.Value {
	return slog.StringValue(s.String())
}

// isMinute returns true if the given time is a minute
// included in the schedule
func (s *Schedule) isMinute(t time.Time) bool {
	if s.allowAnyMinute {
		return true
	}
	m := t.Minute()
	for _, includedMinute := range s.minutes {
		if m == includedMinute {
			return true
		}
	}
	return false
}

// isHour returns true if the given time is an hour
// included in the schedule
func (s *Schedule) isHour(t time.Time) bool {
	if s.allowAnyHour {
		return true
	}
	h := t.Hour()
	for _, includedHour := range s.hours {
		if h == includedHour {
			return true
		}
	}
	return false
}

// isDay returns true if the given time is a day
// included in the schedule. If "L" is used as
// the day, it will be interpreted as the last
// day of the month
func (s *Schedule) isDay(t time.Time) bool {
	if s.allowAnyDay {
		return true
	}
	d := t.Day()
	for _, includedDay := range s.days {
		if d == includedDay {
			return true
		}
	}

	if s.Day() == string(Last) {
		targetMonth := t.Month() + 1
		nextMonth := time.Date(
			t.Year(),
			targetMonth,
			1,
			0,
			0,
			0,
			0,
			t.Location(),
		)
		lastOfThisMonth := nextMonth.Add(-24 * time.Hour)
		return lastOfThisMonth.Day() == t.Day()
	}
	return false
}

// isMonth returns true if the given time is a month
// included in the schedule
func (s *Schedule) isMonth(t time.Time) bool {
	if s.allowAnyMonth {
		return true
	}
	m := int(t.Month())
	for _, includedMonth := range s.months {
		if m == includedMonth {
			return true
		}
	}
	return false
}

// isWeekday returns true if the given time is a weekday
// included in the schedule
func (s *Schedule) isWeekday(t time.Time) bool {
	if s.allowAnyWeekday {
		return true
	}
	w := int(t.Weekday())
	for _, includedWeekday := range s.weekdays {
		if w == includedWeekday {
			return true
		}
	}
	return false
}

// validate checks the schedule for errors, and
// assigns the parsed values to the schedule
func (s *Schedule) validate() error {
	errs := make([]error, 0, 5)
	var minutes []int
	var hours []int
	var days []int
	var months []int
	var weekdays []int
	var err error

	anyStr := string(Any)
	blankStr := string(Blank)

	switch ms := s.Minute(); ms {
	case anyStr:
		s.allowAnyMinute = true
	default:
		minutes, err = minuteOpts.parse(ms)
		s.minutes = minutes
		errs = append(errs, err)

		revSlice := make([]int, len(minutes))
		for i, j := 0, len(minutes)-1; i < j; i, j = i+1, j-1 {
			revSlice[i], revSlice[j] = minutes[j], minutes[i]
		}
		s.minutesDesc = revSlice

	}

	switch hs := s.Hour(); hs {
	case anyStr:
		s.allowAnyHour = true
	default:
		hours, err = hourOpts.parse(hs)
		errs = append(errs, err)
		s.hours = hours
	}

	switch ds := s.Day(); ds {
	case anyStr, blankStr:
		s.allowAnyDay = true
	default:
		days, err = dayOpts.parse(ds)
		errs = append(errs, err)
		s.days = days
	}

	switch ms := s.Month(); ms {
	case anyStr, blankStr:
		s.allowAnyMonth = true
	default:
		months, err = monthOpts.parse(ms)
		errs = append(errs, err)
		s.months = months
	}

	switch ws := s.Weekday(); ws {
	case string(Any), string(Blank):
		s.allowAnyWeekday = true
	default:
		weekdays, err = weekdayOpts.parse(ws)
		errs = append(errs, err)
		s.weekdays = weekdays
	}

	return errors.Join(errs...)
}

// field defines a cron field
type field struct {
	// Name is the name of the field
	Name string

	// Index is the field position in the cron expression
	Index int

	// Allowed is the list of allowed values for the field
	// (ex: for days, 1-31, months, 1-12, etc.)
	Allowed []int

	// Conversions is a map of string values to their
	// allowed int values (ex: "JAN" -> 1, "FEB" -> 2, etc.)
	Conversions map[string]int
}

// Min returns the minimum allowed value for the field
func (f field) Min() int {
	return f.Allowed[0]
}

// Max returns the maximum allowed value for the field
func (f field) Max() int {
	return f.Allowed[len(f.Allowed)-1]
}

// error returns an error with a field-specific prefixed message
func (f field) error(msg string) error {
	return fmt.Errorf("invalid %s entry: %s", f.Name, msg)
}

// wrapErr wraps an error with a field-specific message
func (f field) wrapErr(err error) error {
	return fmt.Errorf("invalid %s entry: %w", f.Name, err)
}

// parse parses a string value for the field, returning
// the parsed values (ints to trigger on) or an error
func (f field) parse(s string) ([]int, error) {
	var values []int
	defer func() {
		if values != nil {
			slices.Sort(values)
			values = slices.Compact(values)
		}
	}()

	if s == "" {
		return nil, f.error("empty")
	}

	// if the string is a wildcard, we can just return all
	if s == string(Blank) && f.Index != dayInd && f.Index != monthInd && f.Index != weekdayInd {
		return nil, f.error("wildcard ? only supported for day, month, and weekday fields")
	}

	switch s {
	case string(Any), string(Blank):
		for i := f.Min(); i <= f.Max(); i++ {
			values = append(values, i)
		}
		return values, nil
	}

	// may be a value such as JAN, FEB, FRI, etc., where
	// we need the int equivalent
	s = strings.ToUpper(s)
	if f.Conversions != nil {
		v, ok := f.Conversions[s]
		if ok {
			values = append(values, v)
			return values, nil
		}
	}

	// if we successfully parse the string as an int, we
	// don't have to worry about parsing steps, etc
	m, err := strconv.Atoi(s)
	if err == nil {
		switch {
		case m < f.Min():
			return nil, f.error(fmt.Sprintf("'%s' is less than %d", s, f.Min()))
		case m > f.Max():
			return nil, f.error(
				fmt.Sprintf(
					"'%s' is greater than %d",
					s,
					f.Max(),
				),
			)
		}
		values = append(values, m)
		return values, nil
	} else {
		// but if it fails, we should have a special character in
		// the string
		switch {
		case strings.ContainsRune(s, ListSeparator):
		case strings.ContainsRune(s, Range):
		case strings.ContainsRune(s, Step):
		case strings.ContainsRune(s, Last):
		default:
			return nil, f.wrapErr(err)
		}
	}

	// Check for the list separator next
	// If we have a value like `1,2,3/10`, we want to pull out
	// 1 and 2 first, then parse 3/10
	if strings.ContainsRune(s, ListSeparator) {
		values, err = f.parseList(s)
		return values, err
	}

	// Process step values next, further parsing the string
	// prior to the separator to get the initial value to
	// Start from
	// A minute expression would look like:
	// */10 (every 10th minute, so 4:00, 4:10, 4:20...)
	// 1-40/10 (every 10th minute from 1-50, so 4:01, 4:11, 4:21, 4:31)
	// 5/10 (non-standard, interpreted as every 10th minute from 5-19, so 4:05, 4:15...)
	beforeStep, afterStep, stepFound := strings.Cut(s, string(Step))
	if stepFound {
		values, err = f.parseStep(beforeStep, afterStep)
		return values, err
	}

	before, after, rangeFound := strings.Cut(s, string(Range))
	if rangeFound {
		values, err = f.parseRange(before, after)
		return values, err
	}

	// the above cases may fall through in case of
	// the "L" (Last) special character

	return values, nil
}

// parseStep returns the values specified for the pre-delimiter
// and post-delimiter step entry
func (f field) parseStep(stepRange string, step string) ([]int, error) {
	if stepRange == "" || step == "" {
		return nil, f.error("empty step entry")
	}
	stepVal, err := strconv.Atoi(step)
	if err != nil {
		return nil, f.wrapErr(
			fmt.Errorf(
				"invalid step entry '%s' ('%s')",
				stepRange,
				step,
			),
		)
	}
	if stepVal < 1 {
		return nil, f.error("step must be greater than 0")
	}

	stepRangeValues, err := f.parse(stepRange)
	if err != nil {
		return nil, f.wrapErr(err)
	}

	// Though non-standard, this accounts for cron entries
	// like "5/10 * * * *" which is interpreted here as
	// "every 10th minute from 5-59" as neither a wildcard
	// nor a list was supplied
	if len(stepRangeValues) == 1 {
		minVal := stepRangeValues[0]
		for _, av := range f.Allowed {
			if av > minVal {
				stepRangeValues = append(stepRangeValues, av)
			}
		}
	}

	values := stepValues(stepRangeValues, stepVal)
	if len(values) == 1 {
		return nil, f.wrapErr(fmt.Errorf("step only occurs once"))
	}
	return values, nil
}

// parseRange returns the specified values for the given values
// specified before and after the range delimiter.
// Ex: "1-5" will [1, 2, 3, 4, 5]
func (f field) parseRange(beforeRange string, afterRange string) (
	[]int,
	error,
) {
	if afterRange == "" {
		return nil, f.error("empty end range")
	}

	startMin, err := f.parse(beforeRange)
	if err != nil {
		return nil, f.wrapErr(err)
	}
	if startMin == nil {
		return nil, f.error("empty Start range")
	}
	if len(startMin) > 1 {
		return nil, f.error("multiple Start range values")
	}

	endMin, err := f.parse(afterRange)
	if err != nil {
		return nil, f.wrapErr(err)
	}
	if endMin == nil {
		return nil, f.error("empty end range")
	}
	if len(endMin) > 1 {
		return nil, f.error("multiple end range values")
	}

	startNum := startMin[0]
	endNum := endMin[0]

	if startNum > endNum || startNum == endNum {
		return nil, f.error(
			fmt.Sprintf(
				"Start range '%d' must be less than end range '%d'",
				startNum,
				endNum,
			),
		)
	}
	values := []int{}
	for i := startNum; i <= endNum; i++ {
		values = append(values, i)
	}
	return values, nil
}

// parseList splits the given entry on ListSeparator, parses each individual
// list entry, and returns the fully extracted list of values
func (f field) parseList(s string) ([]int, error) {
	values := []int{}
	for _, ms := range strings.Split(s, string(ListSeparator)) {
		sv, err := f.parse(ms)
		if err != nil {
			return nil, f.wrapErr(err)
		}
		for _, v := range sv {
			values = append(values, v)
		}
	}
	return values, nil
}

// randomStep returns a random step field string value
func (f field) randomStep(r *rand.Rand) string {
	// If this is 1-31, by default rand.Intn(31) would
	// return 0-30. The step range can't be the same number (no 2-2),
	// and we want to leave room for at least step /1, so we
	// want 2-31 for the end value (allowing 1-{2-31} to start).
	// So we add 2 to the end value (so it initially
	// chooses from 0-29), then add 2 (so we end up
	// with 2-31, no zeroes).
	// The start val then chooses between 0-(endval-1) and
	// adds the minimum value, so if the end range is 31,
	// and the min is 1, and the initial random number is
	// 29, we end up with 30-31, which leaves room for a single step.
	endVal := rand.Intn(f.Max()-2) + 2
	startVal := r.Intn(endVal-1) + f.Min()
	step := r.Intn(endVal-startVal) + 1
	return fmt.Sprintf("%d-%d/%d", startVal, endVal, step)
}

// random generates a random value for the given field.
// Extra weight is put on the Any wildcard (*), as it's
// the most common value. Outside of that, the fields
// are randomly selected to be a single digit, a list,
// range or step.
//
// Steps will be created with random ranges, and
// lists will be created with 2, 3, 4 or 5 entries,
// with a slight bias to 2 entries, and those entries
// are biased to single-digit entries, followed by
// ranges, followed by steps (50%, 20%, 10%)
func (f field) random(r *rand.Rand) (string, error) {
	special := []rune{'\n', Any, ListSeparator, Range, Step}
	isAny := r.Intn(10)
	anyStr := string(Any)
	switch f.Index {
	case minuteInd:
		if isAny == 9 {
			return anyStr, nil
		}
	case hourInd:
		if isAny > 6 {
			return anyStr, nil
		}
	case dayInd:
		if isAny > 6 {
			return anyStr, nil
		}
	case monthInd:
		if isAny > 4 {
			return anyStr, nil
		}
	case weekdayInd:
		if isAny > 1 {
			return anyStr, nil
		}
	}

	i := r.Intn(len(special))

	switch c := special[i]; c {
	case Any:
		return string(c), nil
	case Range:
		startInd := r.Intn(len(f.Allowed) - 1)
		start := f.Allowed[startInd]
		tail := f.Allowed[startInd+1:]
		end := tail[r.Intn(len(tail))]
		return fmt.Sprintf("%d%c%d", start, Range, end), nil
	case Step:
		return f.randomStep(r), nil
	case ListSeparator:
		subct := r.Intn(5) + 1
		if subct < 2 {
			subct = 2
		}
		vals := []string{}
		entriesSeen := map[string]bool{}
		for len(vals) < subct {
			v := f.randomNoList(r)
			if _, seen := entriesSeen[v]; seen {
				continue
			}
			vals = append(vals, v)
			entriesSeen[v] = true
		}
		return strings.Join(vals, string(ListSeparator)), nil
	default:
		return strconv.Itoa(f.randomAllowed(r)), nil
	}
}

// randomNoList returns a randomized list value that
// does not contain any nested lists. It prioritizes
// individual values, then ranges, then steps.
func (f field) randomNoList(r *rand.Rand) string {
	special := []rune{
		'\n',
		'\n',
		'\n',
		'\n',
		'\n',
		'\n',
		Range,
		Range,
		Range,
		Step,
	}
	i := r.Intn(len(special))

	switch c := special[i]; c {
	case Range:
		startInd := r.Intn(len(f.Allowed) - 1)
		start := f.Allowed[startInd]
		tail := f.Allowed[startInd+1:]
		end := tail[r.Intn(len(tail))]
		return fmt.Sprintf("%d%c%d", start, Range, end)
	case Step:
		return f.randomStep(r)
	default:
		return strconv.Itoa(f.randomAllowed(r))
	}
}

// randomAllowed returns a random value from field.Allowed
func (f field) randomAllowed(r *rand.Rand) int {
	return f.Allowed[r.Intn(len(f.Allowed))]
}

// stepValues returns a slice of values from the given
// slice, which includes every nth value
func stepValues(values []int, step int) []int {
	steps := []int{}

	for stepInd, stepVal := range values {
		if stepInd%step == 0 {
			steps = append(steps, stepVal)
		}
	}
	return steps
}

func init() {
	slices.Sort(minuteOpts.Allowed)
	slices.Sort(hourOpts.Allowed)
	slices.Sort(dayOpts.Allowed)
	slices.Sort(monthOpts.Allowed)
	slices.Sort(weekdayOpts.Allowed)
}
