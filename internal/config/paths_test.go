package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot resolve home directory: %v", err)
	}

	cases := map[string]string{
		"~":                home,
		"~/Pictures/Walls": filepath.Join(home, "Pictures", "Walls"),
		// "~\Pictures\Walls" is recognized as a tilde form on any OS, but the
		// remainder after "~\" is joined as-is: filepath.Join only treats "/"
		// (and "\" on Windows) as a separator, so the expected value must be
		// computed with the same call the implementation uses, not
		// hand-built with forward slashes.
		`~\Pictures\Walls`:   filepath.Join(home, `Pictures\Walls`),
		"C:\\wallpapers":     "C:\\wallpapers",
		"/absolute/path":     "/absolute/path",
		"relative/path":      "relative/path",
		"~user/not-expanded": "~user/not-expanded",
	}

	for in, want := range cases {
		if got := ExpandTilde(in); got != want {
			t.Errorf("ExpandTilde(%q) = %q, want %q", in, got, want)
		}
	}
}
