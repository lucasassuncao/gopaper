package schedule

import (
	"testing"
	"time"
)

// atDate uses 2028 (a leap year) so Feb 29 test cases are real dates, not
// normalized away by time.Date.
func atDate(month, day int) time.Time {
	return time.Date(2028, time.Month(month), day, 12, 0, 0, 0, time.Local)
}

func TestParseDateRangeValid(t *testing.T) {
	cases := [][2]string{
		{"12-21", "03-20"},
		{"12-24", "12-26"},
		{"02-29", "03-01"}, // leap day must parse
		{"01-01", "12-31"},
	}
	for _, c := range cases {
		if _, err := ParseDateRange(c[0], c[1]); err != nil {
			t.Errorf("ParseDateRange(%q, %q) unexpected error: %v", c[0], c[1], err)
		}
	}
}

func TestParseDateRangeInvalid(t *testing.T) {
	cases := [][2]string{
		{"13-01", "01-01"},
		{"02-30", "03-01"},
		{"1-01", "02-01"},
		{"", "02-01"},
		{"12-21", ""},
	}
	for _, c := range cases {
		if _, err := ParseDateRange(c[0], c[1]); err == nil {
			t.Errorf("ParseDateRange(%q, %q) expected error, got nil", c[0], c[1])
		}
	}
}

func TestDateWindowContainsSameYear(t *testing.T) {
	dw, err := ParseDateRange("12-24", "12-26")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		t    time.Time
		want bool
	}{
		{atDate(12, 24), true},
		{atDate(12, 25), true},
		{atDate(12, 26), true},
		{atDate(12, 23), false},
		{atDate(12, 27), false},
		{atDate(1, 1), false},
	} {
		if got := dw.Contains(tc.t); got != tc.want {
			t.Errorf("Contains(%v) = %v, want %v", tc.t, got, tc.want)
		}
	}
}

func TestDateWindowContainsYearCrossing(t *testing.T) {
	dw, err := ParseDateRange("12-21", "03-20")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		t    time.Time
		want bool
	}{
		{atDate(12, 21), true},
		{atDate(12, 31), true},
		{atDate(1, 1), true},
		{atDate(2, 29), true}, // leap day inside the range
		{atDate(3, 20), true},
		{atDate(3, 21), false},
		{atDate(12, 20), false},
		{atDate(7, 1), false},
	} {
		if got := dw.Contains(tc.t); got != tc.want {
			t.Errorf("Contains(%v) = %v, want %v", tc.t, got, tc.want)
		}
	}
}
