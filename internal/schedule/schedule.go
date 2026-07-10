// Package schedule provides daily time windows used by category variants.
package schedule

import (
	"fmt"
	"strings"
	"time"
)

// Window is an inclusive daily time window with minute granularity. It may
// cross midnight (e.g. 18:00-05:59).
type Window struct {
	start int // minutes since midnight, inclusive
	end   int // minutes since midnight, inclusive
}

// ParseWindow parses a "HH:MM-HH:MM" (24h) string into a Window.
func ParseWindow(s string) (Window, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return Window{}, fmt.Errorf("invalid hours %q: expected format \"HH:MM-HH:MM\"", s)
	}
	start, err := parseHHMM(parts[0])
	if err != nil {
		return Window{}, fmt.Errorf("invalid hours %q: %v", s, err)
	}
	end, err := parseHHMM(parts[1])
	if err != nil {
		return Window{}, fmt.Errorf("invalid hours %q: %v", s, err)
	}
	return Window{start: start, end: end}, nil
}

// parseHHMM converts a strict zero-padded "HH:MM" string to minutes since
// midnight. time.Parse alone is too lenient (it accepts "6:00").
func parseHHMM(s string) (int, error) {
	if len(s) != 5 {
		return 0, fmt.Errorf("%q is not in HH:MM format", s)
	}
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}

// Contains reports whether t's time of day falls inside the window.
func (w Window) Contains(t time.Time) bool {
	m := t.Hour()*60 + t.Minute()
	if w.start <= w.end {
		return m >= w.start && m <= w.end
	}
	// window crosses midnight
	return m >= w.start || m <= w.end
}

// DateWindow is an inclusive calendar date range (month/day only, no
// year), using the same wraparound semantics as Window (e.g.
// "12-21"/"03-20" spans New Year's Eve).
type DateWindow struct {
	start int // month*100+day, inclusive
	end   int // month*100+day, inclusive
}

// ParseDateRange parses two "MM-DD" (zero-padded) strings into a DateWindow.
func ParseDateRange(start, end string) (DateWindow, error) {
	s, err := parseMMDD(start)
	if err != nil {
		return DateWindow{}, fmt.Errorf("invalid date-range start %q: %v", start, err)
	}
	e, err := parseMMDD(end)
	if err != nil {
		return DateWindow{}, fmt.Errorf("invalid date-range end %q: %v", end, err)
	}
	return DateWindow{start: s, end: e}, nil
}

// parseMMDD converts a strict zero-padded "MM-DD" string to a
// month*100+day ordinal, validated as a real calendar date (using a leap
// year as the reference year so Feb 29 parses).
func parseMMDD(s string) (int, error) {
	if len(s) != 5 {
		return 0, fmt.Errorf("%q is not in MM-DD format", s)
	}
	t, err := time.Parse("2006-01-02", "2000-"+s)
	if err != nil {
		return 0, err
	}
	return int(t.Month())*100 + t.Day(), nil
}

// Contains reports whether t's month/day falls inside the range.
func (d DateWindow) Contains(t time.Time) bool {
	md := int(t.Month())*100 + t.Day()
	if d.start <= d.end {
		return md >= d.start && md <= d.end
	}
	return md >= d.start || md <= d.end
}
