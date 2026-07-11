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

func TestMonitorMode(t *testing.T) {
	v := viper.New()
	if got := MonitorMode(v); got != "all" {
		t.Errorf("default: got %q, want all", got)
	}
	v.Set("configuration.behavior.monitor", "per-monitor")
	if got := MonitorMode(v); got != "per-monitor" {
		t.Errorf("explicit: got %q, want per-monitor", got)
	}
	v.Set("configuration.behavior.monitor", "banana")
	if got := MonitorMode(v); got != "all" {
		t.Errorf("unknown value: got %q, want all", got)
	}
}

func TestMonitorModeForCategory(t *testing.T) {
	v := viper.New()
	v.Set("configuration.behavior.monitor", "per-monitor")

	if got := MonitorModeForCategory(v, ""); got != "per-monitor" {
		t.Errorf("no override: got %q, want the global per-monitor", got)
	}
	if got := MonitorModeForCategory(v, "all"); got != "all" {
		t.Errorf("category all: got %q, want all", got)
	}

	v2 := viper.New() // global default (all)
	if got := MonitorModeForCategory(v2, "per-monitor"); got != "per-monitor" {
		t.Errorf("category per-monitor over default: got %q, want per-monitor", got)
	}
	if got := MonitorModeForCategory(v2, ""); got != "all" {
		t.Errorf("all defaults: got %q, want all", got)
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
