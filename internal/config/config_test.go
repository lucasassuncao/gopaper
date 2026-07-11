package config

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestTransitionEnabledForCategory(t *testing.T) {
	cases := []struct {
		name     string
		global   string
		category string
		want     bool
	}{
		{"both unset defaults to fade", "", "", true},
		{"global none, no override", "none", "", false},
		{"global none, category fade wins", "none", "fade", true},
		{"global fade, category none wins", "fade", "none", false},
		{"category none alone", "", "none", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := viper.New()
			if tc.global != "" {
				v.Set("configuration.behavior.transition", tc.global)
			}
			if got := TransitionEnabledForCategory(v, tc.category); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMultiMonitorMode(t *testing.T) {
	v := viper.New()
	if got := MultiMonitorMode(v); got != "same" {
		t.Errorf("default: got %q, want same", got)
	}
	v.Set("configuration.behavior.multi-monitor", "per-monitor")
	if got := MultiMonitorMode(v); got != "per-monitor" {
		t.Errorf("explicit: got %q, want per-monitor", got)
	}
	v.Set("configuration.behavior.multi-monitor", "banana")
	if got := MultiMonitorMode(v); got != "same" {
		t.Errorf("unknown value: got %q, want same", got)
	}
}

func TestMultiMonitorModeForCategory(t *testing.T) {
	v := viper.New()
	v.Set("configuration.behavior.multi-monitor", "per-monitor")

	if got := MultiMonitorModeForCategory(v, ""); got != "per-monitor" {
		t.Errorf("no override: got %q, want the global per-monitor", got)
	}
	if got := MultiMonitorModeForCategory(v, "same"); got != "same" {
		t.Errorf("category same: got %q, want same", got)
	}

	v2 := viper.New() // global default (same)
	if got := MultiMonitorModeForCategory(v2, "per-monitor"); got != "per-monitor" {
		t.Errorf("category per-monitor over default: got %q, want per-monitor", got)
	}
	if got := MultiMonitorModeForCategory(v2, ""); got != "same" {
		t.Errorf("all defaults: got %q, want same", got)
	}
}

func TestWallhavenCacheDirOverride(t *testing.T) {
	v := viper.New()
	dir, err := WallhavenCacheDir(v, "My Category", `C:\walls\wh-cache`)
	if err != nil {
		t.Fatal(err)
	}
	if dir != `C:\walls\wh-cache` {
		t.Errorf("got %q, want the override as-is", dir)
	}
}

func TestWallhavenCacheDirDefault(t *testing.T) {
	v := viper.New()
	v.Set("configuration.history.file", filepath.Join(t.TempDir(), "gopaper.json"))
	dir, err := WallhavenCacheDir(v, "Wallhaven Landscapes!", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(dir, filepath.Join("wallhaven-cache", "wallhaven-landscapes")) {
		t.Errorf("got %q, want a wallhaven-cache/wallhaven-landscapes suffix", dir)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Wallhaven Landscapes": "wallhaven-landscapes",
		"  Weird -- Name!!":    "weird-name",
		"UPPER123":             "upper123",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
