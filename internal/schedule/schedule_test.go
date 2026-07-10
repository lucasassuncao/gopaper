package schedule

import (
	"testing"
	"time"
)

func at(hour, min int) time.Time {
	return time.Date(2026, 7, 10, hour, min, 30, 0, time.Local)
}

func TestParseWindowValid(t *testing.T) {
	cases := []string{"06:00-17:59", "18:00-05:59", "00:00-23:59", "12:00-12:00"}
	for _, c := range cases {
		if _, err := ParseWindow(c); err != nil {
			t.Errorf("ParseWindow(%q) unexpected error: %v", c, err)
		}
	}
}

func TestParseWindowInvalid(t *testing.T) {
	cases := []string{"", "06:00", "6:00-17:59", "06:00-25:00", "06:60-17:59", "banana", "06:00-17:59-18:00"}
	for _, c := range cases {
		if _, err := ParseWindow(c); err == nil {
			t.Errorf("ParseWindow(%q) expected error, got nil", c)
		}
	}
}

func TestContainsSameDayWindow(t *testing.T) {
	w, err := ParseWindow("06:00-17:59")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		t    time.Time
		want bool
	}{
		{at(6, 0), true},   // inclusive start
		{at(17, 59), true}, // inclusive end
		{at(12, 0), true},
		{at(5, 59), false},
		{at(18, 0), false},
		{at(0, 0), false},
	} {
		if got := w.Contains(tc.t); got != tc.want {
			t.Errorf("Contains(%02d:%02d) = %v, want %v", tc.t.Hour(), tc.t.Minute(), got, tc.want)
		}
	}
}

func TestContainsMidnightCrossingWindow(t *testing.T) {
	w, err := ParseWindow("18:00-05:59")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		t    time.Time
		want bool
	}{
		{at(18, 0), true}, // inclusive start
		{at(5, 59), true}, // inclusive end
		{at(23, 30), true},
		{at(0, 0), true},
		{at(6, 0), false},
		{at(17, 59), false},
		{at(12, 0), false},
	} {
		if got := w.Contains(tc.t); got != tc.want {
			t.Errorf("Contains(%02d:%02d) = %v, want %v", tc.t.Hour(), tc.t.Minute(), got, tc.want)
		}
	}
}
