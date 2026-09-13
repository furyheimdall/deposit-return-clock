package clock

import (
	"fmt"
	"time"
)

// Date is a timezone-independent civil calendar date (year-month-day).
// Possession-clock math is in calendar days; time-of-day is not used.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// NewDate returns the civil date for year, month, day.
// Values are normalized the same way as time.Date (e.g. January 32 → February 1).
func NewDate(year int, month time.Month, day int) Date {
	t := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return Date{Year: t.Year(), Month: t.Month(), Day: t.Day()}
}

// DateFromTime returns the UTC civil date of t.
func DateFromTime(t time.Time) Date {
	u := t.UTC()
	return Date{Year: u.Year(), Month: u.Month(), Day: u.Day()}
}

// IsZero reports whether d is the zero Date.
func (d Date) IsZero() bool {
	return d == Date{}
}

// Equal reports whether d and o are the same civil date.
func (d Date) Equal(o Date) bool {
	return d.Year == o.Year && d.Month == o.Month && d.Day == o.Day
}

// Before reports whether d is before o.
func (d Date) Before(o Date) bool {
	return d.UTC().Before(o.UTC())
}

// After reports whether d is after o.
func (d Date) After(o Date) bool {
	return d.UTC().After(o.UTC())
}

// AddDays returns d plus n calendar days. n may be negative.
func (d Date) AddDays(n int) Date {
	return DateFromTime(d.UTC().AddDate(0, 0, n))
}

// UTC returns midnight UTC on d.
func (d Date) UTC() time.Time {
	if d.IsZero() {
		return time.Time{}
	}
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}

// String returns d in YYYY-MM-DD form, or "" if zero.
func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, int(d.Month), d.Day)
}
