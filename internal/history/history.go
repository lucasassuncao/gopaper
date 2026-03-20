package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lucasassuncao/gopaper/internal/models"
)

const defaultMaxEntries = 50

var (
	ErrHistoryEmpty   = errors.New("wallpaper history is empty")
	ErrAlreadyOldest  = errors.New("already at the oldest wallpaper in history")
	ErrAlreadyNewest  = errors.New("already at the most recent wallpaper in history")
)

// DefaultPath returns the path to the history file, located in a "history"
// subdirectory next to the executable: <exe_dir>/history/gopaper.json
func DefaultPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not determine executable path: %w", err)
	}
	return filepath.Join(filepath.Dir(ex), "history", "gopaper.json"), nil
}

// Load reads the history file from path. Returns an empty History if the file does not exist.
func Load(path string) (*models.History, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &models.History{MaxEntries: defaultMaxEntries}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not read history file: %w", err)
	}

	var h models.History
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("could not parse history file: %w", err)
	}

	if h.MaxEntries == 0 {
		h.MaxEntries = defaultMaxEntries
	}

	return &h, nil
}

// Save writes the history to path, creating parent directories as needed.
func Save(path string, h *models.History) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("could not create history directory: %w", err)
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("could not serialize history: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("could not write history file: %w", err)
	}

	return nil
}

// Append adds a new entry at the front of the history (newest-first),
// resets CurrentIndex to 0, and trims the list to MaxEntries.
func Append(h *models.History, entry models.HistoryEntry) {
	h.Entries = append([]models.HistoryEntry{entry}, h.Entries...)
	h.CurrentIndex = 0

	if len(h.Entries) > h.MaxEntries {
		h.Entries = h.Entries[:h.MaxEntries]
	}
}

// Prev moves CurrentIndex to an older entry and returns it.
func Prev(h *models.History) (models.HistoryEntry, error) {
	if len(h.Entries) == 0 {
		return models.HistoryEntry{}, ErrHistoryEmpty
	}
	if h.CurrentIndex >= len(h.Entries)-1 {
		return models.HistoryEntry{}, ErrAlreadyOldest
	}

	h.CurrentIndex++
	return h.Entries[h.CurrentIndex], nil
}

// Next moves CurrentIndex to a newer entry and returns it.
func Next(h *models.History) (models.HistoryEntry, error) {
	if len(h.Entries) == 0 {
		return models.HistoryEntry{}, ErrHistoryEmpty
	}
	if h.CurrentIndex <= 0 {
		return models.HistoryEntry{}, ErrAlreadyNewest
	}

	h.CurrentIndex--
	return h.Entries[h.CurrentIndex], nil
}
