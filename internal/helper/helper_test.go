package helper

import (
	"os"
	"testing"
	"time"

	"github.com/lucasassuncao/gopaper/internal/filters"
	"github.com/lucasassuncao/gopaper/internal/models"
)

// mockDirEntry implements os.DirEntry for testing purposes.
type mockDirEntry struct {
	name    string
	isDir   bool
	size    int64
	modTime time.Time
}

func (m mockDirEntry) Name() string      { return m.name }
func (m mockDirEntry) IsDir() bool       { return m.isDir }
func (m mockDirEntry) Type() os.FileMode { return 0 }
func (m mockDirEntry) Info() (os.FileInfo, error) {
	return mockFileInfo{name: m.name, size: m.size, modTime: m.modTime}, nil
}

// mockFileInfo implements os.FileInfo for testing purposes.
type mockFileInfo struct {
	name    string
	size    int64
	modTime time.Time
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() os.FileMode  { return 0 }
func (m mockFileInfo) ModTime() time.Time { return m.modTime }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() any           { return nil }

// --- GetEnabledCategories ---

func TestGetEnabledCategories_AllEnabled(t *testing.T) {
	input := []*models.Categories{
		{Name: "a", Enabled: true},
		{Name: "b", Enabled: true},
	}
	result := GetEnabledCategories(input)
	if len(result) != 2 {
		t.Errorf("expected 2 enabled categories, got %d", len(result))
	}
}

func TestGetEnabledCategories_NoneEnabled(t *testing.T) {
	input := []*models.Categories{
		{Name: "a", Enabled: false},
		{Name: "b", Enabled: false},
	}
	result := GetEnabledCategories(input)
	if len(result) != 0 {
		t.Errorf("expected 0 enabled categories, got %d", len(result))
	}
}

func TestGetEnabledCategories_Mixed(t *testing.T) {
	input := []*models.Categories{
		{Name: "a", Enabled: true},
		{Name: "b", Enabled: false},
		{Name: "c", Enabled: true},
	}
	result := GetEnabledCategories(input)
	if len(result) != 2 {
		t.Errorf("expected 2 enabled categories, got %d", len(result))
	}
	for _, c := range result {
		if !c.Enabled {
			t.Errorf("expected only enabled categories, got disabled: %s", c.Name)
		}
	}
}

func TestGetEnabledCategories_Empty(t *testing.T) {
	result := GetEnabledCategories([]*models.Categories{})
	if len(result) != 0 {
		t.Errorf("expected 0 categories, got %d", len(result))
	}
}

// --- GetRandomCategory ---

func TestGetRandomCategory_Empty(t *testing.T) {
	result := GetRandomCategory([]*models.Categories{})
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestGetRandomCategory_Single(t *testing.T) {
	cat := &models.Categories{Name: "only"}
	result := GetRandomCategory([]*models.Categories{cat})
	if result != cat {
		t.Errorf("expected the only category to be returned")
	}
}

func TestGetRandomCategory_ReturnsValidEntry(t *testing.T) {
	input := []*models.Categories{
		{Name: "x"},
		{Name: "y"},
		{Name: "z"},
	}
	for range 50 {
		result := GetRandomCategory(input)
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		found := false
		for _, c := range input {
			if c == result {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("returned category not in input list: %v", result)
		}
	}
}

// --- GetRandomFile ---

func TestGetRandomFile_Empty(t *testing.T) {
	_, err := GetRandomFile([]os.DirEntry{}, nil)
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestGetRandomFile_OnlyDirectories(t *testing.T) {
	entries := []os.DirEntry{
		mockDirEntry{name: "subdir1", isDir: true},
		mockDirEntry{name: "subdir2", isDir: true},
	}
	_, err := GetRandomFile(entries, nil)
	if err == nil {
		t.Error("expected error when only directories are present, got nil")
	}
}

func TestGetRandomFile_OnlyUnsupportedExtensions(t *testing.T) {
	entries := []os.DirEntry{
		mockDirEntry{name: "readme.txt", isDir: false},
		mockDirEntry{name: "data.json", isDir: false},
		mockDirEntry{name: "image.bmp", isDir: false},
	}
	_, err := GetRandomFile(entries, nil)
	if err == nil {
		t.Error("expected error when no supported image files are present, got nil")
	}
}

func TestGetRandomFile_ReturnsImageFile(t *testing.T) {
	for _, name := range []string{"image.jpg", "photo.jpeg", "wall.png", "bg.webp"} {
		entries := []os.DirEntry{mockDirEntry{name: name, isDir: false}}
		got, err := GetRandomFile(entries, nil)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", name, err)
		}
		if got != name {
			t.Errorf("expected '%s', got '%s'", name, got)
		}
	}
}

func TestGetRandomFile_CaseInsensitiveExtension(t *testing.T) {
	entries := []os.DirEntry{
		mockDirEntry{name: "wall.JPG", isDir: false},
		mockDirEntry{name: "photo.PNG", isDir: false},
	}
	_, err := GetRandomFile(entries, nil)
	if err != nil {
		t.Errorf("expected .JPG/.PNG to be accepted, got error: %v", err)
	}
}

func TestGetRandomFile_NeverReturnsDirectoryOrUnsupported(t *testing.T) {
	entries := []os.DirEntry{
		mockDirEntry{name: "subdir", isDir: true},
		mockDirEntry{name: "readme.txt", isDir: false},
		mockDirEntry{name: "wallpaper.png", isDir: false},
		mockDirEntry{name: "another.jpg", isDir: false},
	}
	for range 50 {
		name, err := GetRandomFile(entries, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name == "subdir" || name == "readme.txt" {
			t.Errorf("GetRandomFile returned non-image entry: %s", name)
		}
	}
}

func TestGetRandomFile_MatchFilterExcludesNonMatching(t *testing.T) {
	entries := []os.DirEntry{
		mockDirEntry{name: "screenshot_01.png"},
		mockDirEntry{name: "wallpaper.png"},
	}
	filter, err := filters.Compile(&models.Filter{Match: &models.MatchFilter{Glob: "screenshot_*"}})
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	for range 20 {
		name, err := GetRandomFile(entries, filter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != "screenshot_01.png" {
			t.Errorf("expected only screenshot_01.png to match, got %s", name)
		}
	}
}

func TestGetRandomFile_SizeFilterExcludesTooSmall(t *testing.T) {
	entries := []os.DirEntry{
		mockDirEntry{name: "small.jpg", size: 1_000},
		mockDirEntry{name: "large.jpg", size: 5_000_000},
	}
	filter, err := filters.Compile(&models.Filter{Size: &models.SizeFilter{Min: "1MB"}})
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	for range 20 {
		name, err := GetRandomFile(entries, filter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != "large.jpg" {
			t.Errorf("expected only large.jpg to pass the size filter, got %s", name)
		}
	}
}

func TestGetRandomFile_AgeFilterExcludesTooNew(t *testing.T) {
	entries := []os.DirEntry{
		mockDirEntry{name: "new.jpg", modTime: time.Now()},
		mockDirEntry{name: "old.jpg", modTime: time.Now().Add(-48 * time.Hour)},
	}
	filter, err := filters.Compile(&models.Filter{Age: &models.AgeFilter{Min: 24 * time.Hour}})
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	for range 20 {
		name, err := GetRandomFile(entries, filter)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name != "old.jpg" {
			t.Errorf("expected only old.jpg to pass the age filter, got %s", name)
		}
	}
}

func TestGetRandomFile_FilterExcludingEverythingReturnsError(t *testing.T) {
	entries := []os.DirEntry{mockDirEntry{name: "wallpaper.jpg"}}
	filter, err := filters.Compile(&models.Filter{Match: &models.MatchFilter{Literal: "nomatch.jpg"}})
	if err != nil {
		t.Fatalf("unexpected compile error: %v", err)
	}

	if _, err := GetRandomFile(entries, filter); err == nil {
		t.Error("expected error when the filter excludes every candidate, got nil")
	}
}
