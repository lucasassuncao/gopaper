package helper

import (
	"os"
	"testing"

	"github.com/lucasassuncao/gopaper/internal/models"
)

// mockDirEntry implements os.DirEntry for testing purposes.
type mockDirEntry struct {
	name  string
	isDir bool
}

func (m mockDirEntry) Name() string               { return m.name }
func (m mockDirEntry) IsDir() bool                { return m.isDir }
func (m mockDirEntry) Type() os.FileMode          { return 0 }
func (m mockDirEntry) Info() (os.FileInfo, error) { return nil, nil }

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
	_, err := GetRandomFile([]os.DirEntry{})
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestGetRandomFile_OnlyDirectories(t *testing.T) {
	entries := []os.DirEntry{
		mockDirEntry{name: "subdir1", isDir: true},
		mockDirEntry{name: "subdir2", isDir: true},
	}
	_, err := GetRandomFile(entries)
	if err == nil {
		t.Error("expected error when only directories are present, got nil")
	}
}

func TestGetRandomFile_ReturnsFile(t *testing.T) {
	entries := []os.DirEntry{
		mockDirEntry{name: "image.jpg", isDir: false},
	}
	name, err := GetRandomFile(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "image.jpg" {
		t.Errorf("expected 'image.jpg', got '%s'", name)
	}
}

func TestGetRandomFile_NeverReturnsDirectory(t *testing.T) {
	entries := []os.DirEntry{
		mockDirEntry{name: "subdir", isDir: true},
		mockDirEntry{name: "wallpaper.png", isDir: false},
		mockDirEntry{name: "another.jpg", isDir: false},
	}
	for range 50 {
		name, err := GetRandomFile(entries)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if name == "subdir" {
			t.Error("GetRandomFile returned a directory")
		}
	}
}
