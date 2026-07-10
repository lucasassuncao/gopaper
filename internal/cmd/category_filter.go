package cmd

import (
	"fmt"
	"strings"

	"github.com/lucasassuncao/gopaper/internal/helper"
	"github.com/lucasassuncao/gopaper/internal/models"

	"github.com/pterm/pterm"
)

// ParseCategoryNames splits a comma-separated category string into a slice of
// trimmed names. Returns nil when raw is empty or contains only separators.
func ParseCategoryNames(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var names []string
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			names = append(names, s)
		}
	}
	return names
}

// FilterCategories returns the subset of all eligible for random selection.
//
// When names is empty, it behaves like today: enabled categories only, unless
// includeDisabled is set. When names is non-empty, each name is looked up in
// all; an unknown name is an error. A disabled category named explicitly is
// skipped with a warning unless includeDisabled is set.
func FilterCategories(all []*models.Categories, names []string, includeDisabled bool, logger *pterm.Logger) ([]*models.Categories, error) {
	if len(names) == 0 {
		if includeDisabled {
			return all, nil
		}
		return helper.GetEnabledCategories(all), nil
	}

	byName := make(map[string]*models.Categories, len(all))
	for _, c := range all {
		byName[c.Name] = c
	}

	var selected []*models.Categories
	for _, name := range names {
		cat, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("unknown category %q", name)
		}
		if !cat.Enabled && !includeDisabled {
			if logger != nil {
				logger.Warn("skipping disabled category", logger.Args("category", cat.Name, "hint", "use --include-disabled to select it anyway"))
			}
			continue
		}
		selected = append(selected, cat)
	}
	return selected, nil
}
