package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const defaultMaxEntries = 50

var (
	ErrHistoryEmpty  = errors.New("wallpaper history is empty")
	ErrAlreadyOldest = errors.New("already at the oldest wallpaper in history")
	ErrAlreadyNewest = errors.New("already at the most recent wallpaper in history")
)

// Entry represents a single wallpaper that was applied. For a per-monitor
// change, Path/Category mirror the primary monitor's selection and Monitors
// holds every monitor's wallpaper; for a regular change Monitors is empty.
type Entry struct {
	Path      string         `json:"path"`
	Category  string         `json:"category"`
	Mode      string         `json:"mode"`
	Timestamp time.Time      `json:"timestamp"`
	Monitors  []MonitorEntry `json:"monitors,omitempty"`
}

// MonitorEntry is one monitor's wallpaper within a per-monitor change.
// Monitor is 1-based, matching the categories[].monitor config field.
type MonitorEntry struct {
	Monitor  int    `json:"monitor"`
	Path     string `json:"path"`
	Category string `json:"category"`
}

// History is the persistent navigation state for wallpaper history.
// Entries are stored newest-first (index 0 = most recently applied).
// CurrentIndex tracks which entry is currently displayed.
type History struct {
	Entries      []Entry `json:"entries"`
	CurrentIndex int     `json:"current_index"`
	MaxEntries   int     `json:"max_entries"`
}

// DefaultPath returns the path to the history file, located in a "history"
// subdirectory next to the executable: <exe_dir>/history/gopaper.json
func DefaultPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not determine executable path: %w", err)
	}
	return filepath.Join(filepath.Dir(ex), "history", "gopaper.json"), nil
}

// Load reads the history file from path. limit overrides the stored
// MaxEntries when positive; pass 0 to keep the previously saved value (or
// the package default for a file that doesn't exist yet).
func Load(path string, limit int) (*History, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is derived from os.Executable() or configuration.history.file, not user input
	if os.IsNotExist(err) {
		h := &History{MaxEntries: defaultMaxEntries}
		if limit > 0 {
			h.MaxEntries = limit
		}
		return h, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read history file: %w", err)
	}

	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("could not parse history file: %w", err)
	}

	if limit > 0 {
		h.MaxEntries = limit
	} else if h.MaxEntries == 0 {
		h.MaxEntries = defaultMaxEntries
	}

	return &h, nil
}

// Save writes the history to path, creating parent directories as needed.
func Save(path string, h *History) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("could not create history directory: %w", err)
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("could not serialize history: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("could not write history file: %w", err)
	}

	return nil
}

// Append adds a new entry at the front of the history (newest-first),
// resets CurrentIndex to 0, and trims the list to MaxEntries.
func Append(h *History, entry Entry) {
	h.Entries = append([]Entry{entry}, h.Entries...)
	h.CurrentIndex = 0

	if len(h.Entries) > h.MaxEntries {
		h.Entries = h.Entries[:h.MaxEntries]
	}
}

// Prev moves CurrentIndex to an older entry and returns it.
func Prev(h *History) (Entry, error) {
	if len(h.Entries) == 0 {
		return Entry{}, ErrHistoryEmpty
	}
	if h.CurrentIndex >= len(h.Entries)-1 {
		return Entry{}, ErrAlreadyOldest
	}

	h.CurrentIndex++
	return h.Entries[h.CurrentIndex], nil
}

// Next moves CurrentIndex to a newer entry and returns it.
func Next(h *History) (Entry, error) {
	if len(h.Entries) == 0 {
		return Entry{}, ErrHistoryEmpty
	}
	if h.CurrentIndex <= 0 {
		return Entry{}, ErrAlreadyNewest
	}

	h.CurrentIndex--
	return h.Entries[h.CurrentIndex], nil
}
