package history

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/lucasassuncao/gopaper/internal/models"
)

func entry(path string) models.HistoryEntry {
	return models.HistoryEntry{Path: path, Category: "test", Mode: "crop", Timestamp: time.Now()}
}

// --- Append ---

func TestAppend_AddsToFront(t *testing.T) {
	h := &models.History{MaxEntries: 50}
	Append(h, entry("a.jpg"))
	Append(h, entry("b.jpg"))

	if h.Entries[0].Path != "b.jpg" {
		t.Errorf("expected b.jpg at index 0, got %s", h.Entries[0].Path)
	}
	if h.Entries[1].Path != "a.jpg" {
		t.Errorf("expected a.jpg at index 1, got %s", h.Entries[1].Path)
	}
}

func TestAppend_ResetsCurrentIndex(t *testing.T) {
	h := &models.History{MaxEntries: 50, CurrentIndex: 3}
	Append(h, entry("x.jpg"))

	if h.CurrentIndex != 0 {
		t.Errorf("expected CurrentIndex=0 after Append, got %d", h.CurrentIndex)
	}
}

func TestAppend_TrimsToMaxEntries(t *testing.T) {
	h := &models.History{MaxEntries: 3}
	Append(h, entry("a.jpg"))
	Append(h, entry("b.jpg"))
	Append(h, entry("c.jpg"))
	Append(h, entry("d.jpg"))

	if len(h.Entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(h.Entries))
	}
	if h.Entries[len(h.Entries)-1].Path != "b.jpg" {
		t.Errorf("expected oldest entry to be b.jpg, got %s", h.Entries[len(h.Entries)-1].Path)
	}
}

// --- Prev ---

func TestPrev_EmptyHistory(t *testing.T) {
	h := &models.History{MaxEntries: 50}
	_, err := Prev(h)
	if err != ErrHistoryEmpty {
		t.Errorf("expected ErrHistoryEmpty, got %v", err)
	}
}

func TestPrev_AlreadyAtOldest(t *testing.T) {
	h := &models.History{MaxEntries: 50}
	Append(h, entry("a.jpg"))

	_, err := Prev(h)
	if err != ErrAlreadyOldest {
		t.Errorf("expected ErrAlreadyOldest, got %v", err)
	}
}

func TestPrev_MovesToOlderEntry(t *testing.T) {
	h := &models.History{MaxEntries: 50}
	Append(h, entry("a.jpg"))
	Append(h, entry("b.jpg"))
	// Entries: [b, a], CurrentIndex=0

	got, err := Prev(h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Path != "a.jpg" {
		t.Errorf("expected a.jpg, got %s", got.Path)
	}
	if h.CurrentIndex != 1 {
		t.Errorf("expected CurrentIndex=1, got %d", h.CurrentIndex)
	}
}

// --- Next ---

func TestNext_EmptyHistory(t *testing.T) {
	h := &models.History{MaxEntries: 50}
	_, err := Next(h)
	if err != ErrHistoryEmpty {
		t.Errorf("expected ErrHistoryEmpty, got %v", err)
	}
}

func TestNext_AlreadyAtNewest(t *testing.T) {
	h := &models.History{MaxEntries: 50}
	Append(h, entry("a.jpg"))
	// CurrentIndex=0, only 1 entry

	_, err := Next(h)
	if err != ErrAlreadyNewest {
		t.Errorf("expected ErrAlreadyNewest, got %v", err)
	}
}

func TestNext_MovesToNewerEntry(t *testing.T) {
	h := &models.History{MaxEntries: 50}
	Append(h, entry("a.jpg"))
	Append(h, entry("b.jpg"))
	// Entries: [b, a], CurrentIndex=0

	Prev(h) // move to index 1 (a.jpg)

	got, err := Next(h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Path != "b.jpg" {
		t.Errorf("expected b.jpg, got %s", got.Path)
	}
	if h.CurrentIndex != 0 {
		t.Errorf("expected CurrentIndex=0, got %d", h.CurrentIndex)
	}
}

// --- Load / Save round-trip ---

func TestLoadSave_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")

	h := &models.History{MaxEntries: 10}
	Append(h, entry("wall1.jpg"))
	Append(h, entry("wall2.jpg"))

	if err := Save(path, h); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Path != "wall2.jpg" {
		t.Errorf("expected wall2.jpg at index 0, got %s", loaded.Entries[0].Path)
	}
}

func TestLoad_ReturnsEmptyWhenFileNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")

	h, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(h.Entries) != 0 {
		t.Errorf("expected empty entries, got %d", len(h.Entries))
	}
	if h.MaxEntries != defaultMaxEntries {
		t.Errorf("expected MaxEntries=%d, got %d", defaultMaxEntries, h.MaxEntries)
	}
}
