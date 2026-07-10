// Package filters compiles a category's models.Filter into a form that can be
// evaluated cheaply against many candidate files.
package filters

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lucasassuncao/gopaper/internal/models"
)

// Compiled is a validated, ready-to-evaluate form of models.Filter. Regex is
// compiled and size bounds are parsed once, so repeated evaluation against
// many candidate files doesn't repeat that work.
type Compiled struct {
	matchLiteral     string
	matchGlob        string
	matchRegex       *regexp.Regexp
	caseSensitive    bool
	ageMin, ageMax   time.Duration
	sizeMin, sizeMax int64
}

// Compile validates and compiles f. A nil f compiles to a filter that matches
// every file.
func Compile(f *models.Filter) (*Compiled, error) {
	c := &Compiled{}
	if f == nil {
		return c, nil
	}

	if err := c.compileMatch(f.Match); err != nil {
		return nil, err
	}
	if f.Age != nil {
		c.ageMin, c.ageMax = f.Age.Min, f.Age.Max
	}
	if err := c.compileSize(f.Size); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Compiled) compileMatch(m *models.MatchFilter) error {
	if m == nil {
		return nil
	}
	c.matchLiteral = m.Literal
	c.matchGlob = m.Glob
	c.caseSensitive = m.CaseSensitive

	if m.Regex != "" {
		re, err := regexp.Compile(m.Regex)
		if err != nil {
			return fmt.Errorf("invalid filter.match.regex: %w", err)
		}
		c.matchRegex = re
	}
	if m.Glob != "" {
		if _, err := filepath.Match(m.Glob, ""); err != nil {
			return fmt.Errorf("invalid filter.match.glob: %w", err)
		}
	}
	return nil
}

func (c *Compiled) compileSize(s *models.SizeFilter) error {
	if s == nil {
		return nil
	}
	if s.Min != "" {
		v, err := ParseSize(s.Min)
		if err != nil {
			return fmt.Errorf("invalid filter.size.min: %w", err)
		}
		c.sizeMin = v
	}
	if s.Max != "" {
		v, err := ParseSize(s.Max)
		if err != nil {
			return fmt.Errorf("invalid filter.size.max: %w", err)
		}
		c.sizeMax = v
	}
	return nil
}

// NeedsFileInfo reports whether the filter has any age/size constraints, so
// callers can skip the file-info lookup when it would go unused.
func (c *Compiled) NeedsFileInfo() bool {
	return c.ageMin > 0 || c.ageMax > 0 || c.sizeMin > 0 || c.sizeMax > 0
}

// Matches reports whether a file with the given name passes the filter. info
// may be nil when NeedsFileInfo reports false.
func (c *Compiled) Matches(name string, info os.FileInfo) bool {
	if !c.matchesName(name) {
		return false
	}
	if info == nil {
		return true
	}
	if c.ageMin > 0 && time.Since(info.ModTime()) < c.ageMin {
		return false
	}
	if c.ageMax > 0 && time.Since(info.ModTime()) > c.ageMax {
		return false
	}
	if c.sizeMin > 0 && info.Size() < c.sizeMin {
		return false
	}
	if c.sizeMax > 0 && info.Size() > c.sizeMax {
		return false
	}
	return true
}

func (c *Compiled) matchesName(name string) bool {
	if c.matchRegex != nil && !c.matchRegex.MatchString(name) {
		return false
	}
	if c.matchGlob != "" {
		matched, _ := filepath.Match(normalizeCase(c.matchGlob, c.caseSensitive), normalizeCase(name, c.caseSensitive))
		if !matched {
			return false
		}
	}
	if c.matchLiteral != "" && normalizeCase(name, c.caseSensitive) != normalizeCase(c.matchLiteral, c.caseSensitive) {
		return false
	}
	return true
}

func normalizeCase(s string, caseSensitive bool) string {
	if caseSensitive {
		return s
	}
	return strings.ToLower(s)
}

// ParseSize parses a human-readable size string (e.g. "10MB", "1.5GB",
// "256MiB") into bytes. KB/MB/GB/TB are decimal (powers of 1000); KiB/MiB/
// GiB/TiB are binary (powers of 1024).
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Ordered longest-suffix-first so "B" never matches before "MB" or "GiB".
	suffixes := []struct {
		suffix     string
		multiplier int64
	}{
		{"TIB", 1 << 40},
		{"GIB", 1 << 30},
		{"MIB", 1 << 20},
		{"KIB", 1 << 10},
		{"TB", 1_000_000_000_000},
		{"GB", 1_000_000_000},
		{"MB", 1_000_000},
		{"KB", 1_000},
		{"B", 1},
	}

	upper := strings.ToUpper(s)
	for _, entry := range suffixes {
		if strings.HasSuffix(upper, entry.suffix) {
			numStr := strings.TrimSpace(s[:len(s)-len(entry.suffix)])
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, fmt.Errorf("could not parse numeric value %q", numStr)
			}
			if val < 0 {
				return 0, fmt.Errorf("size must not be negative: %q", s)
			}
			bytes := val * float64(entry.multiplier)
			if bytes > float64(math.MaxInt64) {
				return 0, fmt.Errorf("size out of range: %q", s)
			}
			return int64(bytes), nil
		}
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unrecognised size format %q", s)
	}
	if val < 0 {
		return 0, fmt.Errorf("size must not be negative: %q", s)
	}
	return val, nil
}
