package cmd

import (
	"fmt"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/lucasassuncao/gopaper/internal/filters"
	"github.com/lucasassuncao/yedit/editor"
)

// GopaperValidators is the rule set enforced by the edit command at
// validate/save time.
//
// Per-field constraints (required, allowed values, uniqueness) are declared
// once in the hint tree (models.Config.Metadata and friends) and enforced by
// the FromMetadata family — hints are the single source of field metadata.
// Only cross-field rules, which cannot live in per-field metadata, are
// declared here explicitly.
var GopaperValidators = []editor.Validator{
	// Enforce everything the metadata declares.
	editor.RequiredFromMetadata(),
	editor.OneOfFromMetadata(),
	editor.CountFromMetadata(),
	editor.UniqueFromMetadata(),

	// Category names must be unique across the list (presence comes from the
	// hints; NoDuplicates skips unnamed entries).
	editor.NoDuplicates("categories", "name"),

	// within a filter.match block, literal/regex/glob are mutually exclusive.
	editor.MutuallyExclusiveNested("categories.filter.match", "literal", "regex", "glob"),

	// age and size min/max pairs must be ordered.
	editor.CrossFieldOrderedNested("categories.filter.age", "min", "max"),
	editor.CrossFieldOrderedNested("categories.filter.size", "min", "max"),

	// Validate that filter.match.regex/glob compile and filter.size.min/max
	// parse, so a malformed filter is caught here instead of at wallpaper-
	// change time.
	editor.ValidatorFunc(func(in editor.ValidationInput) []editor.Violation {
		var doc struct {
			Categories []struct {
				Filter *struct {
					Match *struct {
						Regex string `yaml:"regex"`
						Glob  string `yaml:"glob"`
					} `yaml:"match"`
					Size *struct {
						Min string `yaml:"min"`
						Max string `yaml:"max"`
					} `yaml:"size"`
				} `yaml:"filter"`
			} `yaml:"categories"`
		}
		if err := yaml.Unmarshal(in.Raw, &doc); err != nil {
			return nil
		}
		var errs []editor.Violation
		for i, c := range doc.Categories {
			if c.Filter == nil {
				continue
			}
			if m := c.Filter.Match; m != nil {
				if m.Regex != "" {
					if _, err := regexp.Compile(m.Regex); err != nil {
						errs = append(errs, editor.Violation{
							Path:    fmt.Sprintf("categories[%d].filter.match.regex", i),
							Message: err.Error(),
						})
					}
				}
				if m.Glob != "" {
					if _, err := filepath.Match(m.Glob, ""); err != nil {
						errs = append(errs, editor.Violation{
							Path:    fmt.Sprintf("categories[%d].filter.match.glob", i),
							Message: err.Error(),
						})
					}
				}
			}
			if s := c.Filter.Size; s != nil {
				if s.Min != "" {
					if _, err := filters.ParseSize(s.Min); err != nil {
						errs = append(errs, editor.Violation{
							Path:    fmt.Sprintf("categories[%d].filter.size.min", i),
							Message: err.Error(),
						})
					}
				}
				if s.Max != "" {
					if _, err := filters.ParseSize(s.Max); err != nil {
						errs = append(errs, editor.Violation{
							Path:    fmt.Sprintf("categories[%d].filter.size.max", i),
							Message: err.Error(),
						})
					}
				}
			}
		}
		return errs
	}),

	// logging.file is required when logging.output is "log", "file", or "both".
	editor.ValidatorFunc(func(in editor.ValidationInput) []editor.Violation {
		var doc struct {
			Configuration struct {
				Logging struct {
					Output string `yaml:"output"`
					File   string `yaml:"file"`
				} `yaml:"logging"`
			} `yaml:"configuration"`
		}
		if err := yaml.Unmarshal(in.Raw, &doc); err != nil {
			return nil
		}
		switch doc.Configuration.Logging.Output {
		case "log", "file", "both":
		default:
			return nil
		}
		if doc.Configuration.Logging.File != "" {
			return nil
		}
		return []editor.Violation{{
			Path:    "configuration.logging.file",
			Message: fmt.Sprintf("required when output is %q", doc.Configuration.Logging.Output),
		}}
	}),
}
