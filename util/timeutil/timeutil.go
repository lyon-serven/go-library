// Package timeutil provides utility functions for time operations.
package timeutil

import (
	"fmt"
	"strconv"
	"time"
)

// Common time formats
const (
	// DateFormat represents date format YYYY-MM-DD
	DateFormat = "2006-01-02"
	// TimeFormat represents time format HH:MM:SS
	TimeFormat = "15:04:05"
	// DateTimeFormat represents datetime format YYYY-MM-DD HH:MM:SS
	DateTimeFormat = "2006-01-02 15:04:05"
	// ISO8601Format represents ISO8601 format
	ISO8601Format = "2006-01-02T15:04:05Z07:00"
	// RFC3339Format represents RFC3339 format
	RFC3339Format = time.RFC3339
	// UnixTimestampFormat represents unix timestamp
	UnixTimestampFormat = "unix"
)

// TimeZone constants
var (
	// Common time zones
	UTC      = time.UTC
	Shanghai = mustLoadLocation("Asia/Shanghai")
	Tokyo    = mustLoadLocation("Asia/Tokyo")
	NewYork  = mustLoadLocation("America/New_York")
	London   = mustLoadLocation("Europe/London")
)

// mustLoadLocation loads a time zone location or panics
func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(fmt.Sprintf("failed to load time zone %s: %v", name, err))
	}
	return loc
}

// Now returns the current time
func Now() time.Time {
	return time.Now()
}

// NowInZone returns the current time in the specified time zone
func NowInZone(zone *time.Location) time.Time {
	return time.Now().In(zone)
}

// Today returns today's date at 00:00:00
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// TodayInZone returns today's date in the specified time zone at 00:00:00
func TodayInZone(zone *time.Location) time.Time {
	now := time.Now().In(zone)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, zone)
}

// StartOfWeek returns the start of the week (Monday 00:00:00)
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	days := weekday - 1
	return time.Date(t.Year(), t.Month(), t.Day()-days, 0, 0, 0, 0, t.Location())
}

// EndOfWeek returns the end of the week (Sunday 23:59:59)
func EndOfWeek(t time.Time) time.Time {
	start := StartOfWeek(t)
	return start.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
}

// StartOfMonth returns the start of the month (1st day 00:00:00)
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the end of the month (last day 23:59:59)
func EndOfMonth(t time.Time) time.Time {
	start := StartOfMonth(t)
	nextMonth := start.AddDate(0, 1, 0)
	return nextMonth.Add(-time.Second)
}

// StartOfYear returns the start of the year (Jan 1st 00:00:00)
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns the end of the year (Dec 31st 23:59:59)
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 0, t.Location())
}

// Format formats time to string using the specified format
func Format(t time.Time, format string) string {
	switch format {
	case UnixTimestampFormat:
		return strconv.FormatInt(t.Unix(), 10)
	default:
		return t.Format(format)
	}
}

// Parse parses time string using the specified format
func Parse(timeStr, format string) (time.Time, error) {
	switch format {
	case UnixTimestampFormat:
		timestamp, err := strconv.ParseInt(timeStr, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid unix timestamp: %w", err)
		}
		return time.Unix(timestamp, 0), nil
	default:
		return time.Parse(format, timeStr)
	}
}

// ParseInLocation parses time string in the specified location
func ParseInLocation(timeStr, format string, loc *time.Location) (time.Time, error) {
	switch format {
	case UnixTimestampFormat:
		timestamp, err := strconv.ParseInt(timeStr, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid unix timestamp: %w", err)
		}
		return time.Unix(timestamp, 0).In(loc), nil
	default:
		return time.ParseInLocation(format, timeStr, loc)
	}
}

// AddBusinessDays adds business days (Monday-Friday) to a time
func AddBusinessDays(t time.Time, days int) time.Time {
	result := t
	remaining := days

	if days > 0 {
		for remaining > 0 {
			result = result.AddDate(0, 0, 1)
			if result.Weekday() != time.Saturday && result.Weekday() != time.Sunday {
				remaining--
			}
		}
	} else if days < 0 {
		remaining = -remaining
		for remaining > 0 {
			result = result.AddDate(0, 0, -1)
			if result.Weekday() != time.Saturday && result.Weekday() != time.Sunday {
				remaining--
			}
		}
	}

	return result
}

// IsBusinessDay checks if the given time is a business day (Monday-Friday)
func IsBusinessDay(t time.Time) bool {
	weekday := t.Weekday()
	return weekday != time.Saturday && weekday != time.Sunday
}

// IsWeekend checks if the given time is weekend (Saturday or Sunday)
func IsWeekend(t time.Time) bool {
	return !IsBusinessDay(t)
}

// DaysBetween returns the number of days between two times
func DaysBetween(start, end time.Time) int {
	if end.Before(start) {
		start, end = end, start
	}

	duration := end.Sub(start)
	return int(duration.Hours() / 24)
}

// BusinessDaysBetween returns the number of business days between two times
func BusinessDaysBetween(start, end time.Time) int {
	if end.Before(start) {
		start, end = end, start
	}

	count := 0
	current := start

	for current.Before(end) || current.Equal(end) {
		if IsBusinessDay(current) {
			count++
		}
		current = current.AddDate(0, 0, 1)
	}

	return count
}

// Age calculates age in years from birthdate
func Age(birthdate time.Time) int {
	now := time.Now()
	age := now.Year() - birthdate.Year()

	// Adjust if birthday hasn't occurred this year yet
	if now.Month() < birthdate.Month() ||
		(now.Month() == birthdate.Month() && now.Day() < birthdate.Day()) {
		age--
	}

	return age
}

// IsLeapYear checks if the given year is a leap year
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth returns the number of days in the given month and year
func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// Quarter returns the quarter (1-4) for the given time
func Quarter(t time.Time) int {
	month := int(t.Month())
	return (month-1)/3 + 1
}

// StartOfQuarter returns the start of the quarter
func StartOfQuarter(t time.Time) time.Time {
	quarter := Quarter(t)
	month := time.Month((quarter-1)*3 + 1)
	return time.Date(t.Year(), month, 1, 0, 0, 0, 0, t.Location())
}

// EndOfQuarter returns the end of the quarter
func EndOfQuarter(t time.Time) time.Time {
	start := StartOfQuarter(t)
	nextQuarter := start.AddDate(0, 3, 0)
	return nextQuarter.Add(-time.Second)
}

// ConvertTimeZone converts time from one timezone to another
func ConvertTimeZone(t time.Time, from, to *time.Location) time.Time {
	// First convert to the source timezone if needed
	if t.Location() != from {
		t = t.In(from)
	}
	// Then convert to target timezone
	return t.In(to)
}

// IsSameDay checks if two times are on the same day
func IsSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// IsSameWeek checks if two times are in the same week
func IsSameWeek(t1, t2 time.Time) bool {
	start1 := StartOfWeek(t1)
	start2 := StartOfWeek(t2)
	return IsSameDay(start1, start2)
}

// IsSameMonth checks if two times are in the same month and year
func IsSameMonth(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month()
}

// IsSameYear checks if two times are in the same year
func IsSameYear(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year()
}

// Sleep pauses the current goroutine for the given duration
func Sleep(duration time.Duration) {
	time.Sleep(duration)
}

// Elapsed returns the time elapsed since the given start time
func Elapsed(start time.Time) time.Duration {
	return time.Since(start)
}

// Benchmark measures the execution time of a function
func Benchmark(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

// FormatDuration formats duration to human readable string
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

// TimeRange represents a time range
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// NewTimeRange creates a new time range
func NewTimeRange(start, end time.Time) TimeRange {
	return TimeRange{Start: start, End: end}
}

// Duration returns the duration of the time range
func (tr TimeRange) Duration() time.Duration {
	return tr.End.Sub(tr.Start)
}

// Contains checks if the time range contains the given time
func (tr TimeRange) Contains(t time.Time) bool {
	return (t.Equal(tr.Start) || t.After(tr.Start)) &&
		(t.Equal(tr.End) || t.Before(tr.End))
}

// Overlaps checks if two time ranges overlap
func (tr TimeRange) Overlaps(other TimeRange) bool {
	return tr.Start.Before(other.End) && other.Start.Before(tr.End)
}

// Split splits the time range into smaller ranges of the given duration
func (tr TimeRange) Split(duration time.Duration) []TimeRange {
	var ranges []TimeRange
	current := tr.Start

	for current.Before(tr.End) {
		next := current.Add(duration)
		if next.After(tr.End) {
			next = tr.End
		}
		ranges = append(ranges, TimeRange{Start: current, End: next})
		current = next
	}

	return ranges
}
