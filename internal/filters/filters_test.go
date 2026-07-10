package filters

import (
	"os"
	"testing"
	"time"

	"github.com/lucasassuncao/gopaper/internal/models"
)

type fakeInfo struct {
	size    int64
	modTime time.Time
}

func (f fakeInfo) Name() string       { return "fake" }
func (f fakeInfo) Size() int64        { return f.size }
func (f fakeInfo) Mode() os.FileMode  { return 0 }
func (f fakeInfo) ModTime() time.Time { return f.modTime }
func (f fakeInfo) IsDir() bool        { return false }
func (f fakeInfo) Sys() any           { return nil }

func TestCompile_NilFilterMatchesEverything(t *testing.T) {
	c, err := Compile(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.Matches("anything.jpg", nil) {
		t.Error("expected nil filter to match everything")
	}
}

func TestCompile_InvalidRegex(t *testing.T) {
	_, err := Compile(&models.Filter{Match: &models.MatchFilter{Regex: "["}})
	if err == nil {
		t.Error("expected error for invalid regex")
	}
}

func TestCompile_InvalidGlob(t *testing.T) {
	_, err := Compile(&models.Filter{Match: &models.MatchFilter{Glob: "["}})
	if err == nil {
		t.Error("expected error for invalid glob")
	}
}

func TestCompile_InvalidSize(t *testing.T) {
	_, err := Compile(&models.Filter{Size: &models.SizeFilter{Min: "not-a-size"}})
	if err == nil {
		t.Error("expected error for invalid size")
	}
}

func TestMatches_Literal(t *testing.T) {
	c, _ := Compile(&models.Filter{Match: &models.MatchFilter{Literal: "Anna's Archive.pdf"}})
	if !c.Matches("Anna's Archive.pdf", nil) {
		t.Error("expected exact literal match")
	}
	if c.Matches("other.pdf", nil) {
		t.Error("expected non-matching literal to fail")
	}
}

func TestMatches_LiteralCaseInsensitiveByDefault(t *testing.T) {
	c, _ := Compile(&models.Filter{Match: &models.MatchFilter{Literal: "Wallpaper.JPG"}})
	if !c.Matches("wallpaper.jpg", nil) {
		t.Error("expected case-insensitive match by default")
	}
}

func TestMatches_LiteralCaseSensitive(t *testing.T) {
	c, _ := Compile(&models.Filter{Match: &models.MatchFilter{Literal: "Wallpaper.JPG", CaseSensitive: true}})
	if c.Matches("wallpaper.jpg", nil) {
		t.Error("expected case-sensitive match to fail on differing case")
	}
}

func TestMatches_Glob(t *testing.T) {
	c, _ := Compile(&models.Filter{Match: &models.MatchFilter{Glob: "screenshot_*.png"}})
	if !c.Matches("screenshot_2024.png", nil) {
		t.Error("expected glob to match")
	}
	if c.Matches("photo.png", nil) {
		t.Error("expected glob to reject non-matching name")
	}
}

func TestMatches_Regex(t *testing.T) {
	c, _ := Compile(&models.Filter{Match: &models.MatchFilter{Regex: `^\d{4}-\d{2}-\d{2}_`}})
	if !c.Matches("2024-01-02_wallpaper.jpg", nil) {
		t.Error("expected regex to match")
	}
	if c.Matches("wallpaper.jpg", nil) {
		t.Error("expected regex to reject non-matching name")
	}
}

func TestMatches_SizeMinMax(t *testing.T) {
	c, _ := Compile(&models.Filter{Size: &models.SizeFilter{Min: "1KB", Max: "1MB"}})

	if !c.NeedsFileInfo() {
		t.Fatal("expected NeedsFileInfo to be true when size is set")
	}
	if !c.Matches("f.jpg", fakeInfo{size: 500_000}) {
		t.Error("expected in-range size to match")
	}
	if c.Matches("f.jpg", fakeInfo{size: 10}) {
		t.Error("expected below-min size to fail")
	}
	if c.Matches("f.jpg", fakeInfo{size: 5_000_000}) {
		t.Error("expected above-max size to fail")
	}
}

func TestMatches_AgeMinMax(t *testing.T) {
	c, _ := Compile(&models.Filter{Age: &models.AgeFilter{Min: 24 * time.Hour, Max: 72 * time.Hour}})

	if !c.NeedsFileInfo() {
		t.Fatal("expected NeedsFileInfo to be true when age is set")
	}
	if !c.Matches("f.jpg", fakeInfo{modTime: time.Now().Add(-48 * time.Hour)}) {
		t.Error("expected in-range age to match")
	}
	if c.Matches("f.jpg", fakeInfo{modTime: time.Now()}) {
		t.Error("expected too-new file to fail min age")
	}
	if c.Matches("f.jpg", fakeInfo{modTime: time.Now().Add(-200 * time.Hour)}) {
		t.Error("expected too-old file to fail max age")
	}
}

func TestParseSize(t *testing.T) {
	cases := map[string]int64{
		"0":     0,
		"100":   100,
		"1KB":   1_000,
		"1MB":   1_000_000,
		"1GB":   1_000_000_000,
		"1KiB":  1024,
		"1MiB":  1024 * 1024,
		"1.5MB": 1_500_000,
	}
	for in, want := range cases {
		got, err := ParseSize(in)
		if err != nil {
			t.Errorf("ParseSize(%q): unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseSize(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestParseSize_Invalid(t *testing.T) {
	for _, in := range []string{"", "-1MB", "abc"} {
		if _, err := ParseSize(in); err == nil {
			t.Errorf("ParseSize(%q): expected error, got nil", in)
		}
	}
}
