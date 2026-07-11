package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/lucasassuncao/gopaper/internal/config"
	"github.com/lucasassuncao/gopaper/internal/history"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// HistoryCmd opens an interactive TUI listing the wallpaper history; Enter
// reapplies the selected entry.
func HistoryCmd() *cobra.Command {
	var categoryFilter string

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Browse the wallpaper history and reapply any entry",
		Long: `Open an interactive list of every wallpaper recorded in history.

Arrow keys (or j/k) navigate, "/" filters by name or category, Enter
reapplies the selected wallpaper (with the configured transition), and
q quits without changing anything.`,
		Example: `  # Browse the full history
  gopaper history

  # Only entries from one category
  gopaper history --category "Saltern Study"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHistoryTUI(categoryFilter)
		},
	}

	cmd.Flags().StringVar(&categoryFilter, "category", "", "Only show entries from this category")
	return cmd
}

// runHistoryTUI loads the history (optionally pre-filtered by category),
// runs the list TUI, and reapplies the entry chosen with Enter, if any.
func runHistoryTUI(categoryFilter string) error {
	v := viper.GetViper()
	if err := config.LoadDefault(v); err != nil {
		var notFound config.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return fmt.Errorf("could not load configuration: %w", err)
		}
		// No config file found: fall back to the built-in history defaults.
	}

	histPath, err := config.HistoryPath(v)
	if err != nil {
		return fmt.Errorf("could not determine history path: %w", err)
	}
	h, err := history.Load(histPath, config.HistoryLimit(v))
	if err != nil {
		return fmt.Errorf("could not load history: %w", err)
	}

	entries := h.Entries
	if categoryFilter != "" {
		var filtered []history.Entry
		for _, e := range entries {
			if e.Category == categoryFilter {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}
	if len(entries) == 0 {
		logger.Warn(history.ErrHistoryEmpty.Error())
		return nil
	}

	items := make([]list.Item, len(entries))
	for i, e := range entries {
		items[i] = historyItem{entry: e}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "gopaper history (newest first) — Enter applies, q quits"
	l.SetShowStatusBar(false)

	m := historyModel{list: l}
	res, err := tea.NewProgram(&m, tea.WithAltScreen()).Run()
	if err != nil {
		return fmt.Errorf("could not run history TUI: %w", err)
	}

	final, ok := res.(*historyModel)
	if !ok || final.chosen == nil {
		return nil
	}
	entry := *final.chosen

	if err := applyHistoryEntry(v, entry); err != nil {
		return err
	}

	// Move the history cursor to the reapplied entry so a subsequent
	// prev/next continues from it, then persist.
	for i, e := range h.Entries {
		if e.Path == entry.Path && e.Timestamp.Equal(entry.Timestamp) {
			h.CurrentIndex = i
			break
		}
	}
	if err := history.Save(histPath, h); err != nil {
		logger.Warn("could not save history", logger.Args("error", err))
	}

	logger.Info("Wallpaper changed successfully.",
		logger.Args("wallpaper", entry.Path, "category", entry.Category, "mode", entry.Mode),
	)
	return nil
}

// historyItem adapts a history.Entry to the bubbles list item interface.
type historyItem struct {
	entry history.Entry
}

func (i historyItem) Title() string { return filepath.Base(i.entry.Path) }

func (i historyItem) Description() string {
	desc := fmt.Sprintf("%s · %s · %s", i.entry.Category, i.entry.Mode, i.entry.Timestamp.Format("2006-01-02 15:04"))
	if n := len(i.entry.Monitors); n > 0 {
		desc += fmt.Sprintf(" · %d monitors", n)
	}
	return desc
}

func (i historyItem) FilterValue() string { return i.entry.Category + " " + i.entry.Path }

// historyModel is the Bubble Tea model for the history browser: a plain
// list where Enter records the selection and quits; the applying happens
// after the program exits (runHistoryTUI), keeping the TUI side-effect-free.
type historyModel struct {
	list   list.Model
	chosen *history.Entry
}

func (m *historyModel) Init() tea.Cmd { return nil }

func (m *historyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		// While the list's fuzzy filter is capturing input, keys belong to it.
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(historyItem); ok {
				m.chosen = &item.entry
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *historyModel) View() string { return m.list.View() }
