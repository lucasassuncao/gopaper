package helper

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lucasassuncao/gopaper/internal/filters"
	"github.com/lucasassuncao/gopaper/internal/models"
	"github.com/lucasassuncao/gopaper/internal/schedule"
	"github.com/lucasassuncao/gopaper/internal/weather"

	"github.com/reujab/wallpaper"
)

// imageExtensions holds the supported wallpaper file extensions.
var imageExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
}

// CreateDirectory checks if the specified directory exists, and if not, creates it with full permissions.
func CreateDirectory(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0o750)
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

// ReadDirectory reads the contents of a given directory and returns the files.
func ReadDirectory(path string) ([]os.DirEntry, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// GetEnabledCategories returns a list of enabled categories from the list of categories.
func GetEnabledCategories(categories []*models.Categories) []*models.Categories {
	var enabledCategories []*models.Categories
	for _, category := range categories {
		if category.Enabled {
			enabledCategories = append(enabledCategories, category)
		}
	}
	return enabledCategories
}

// GetRandomCategory returns a random category from the list of categories.
func GetRandomCategory(categories []*models.Categories) *models.Categories {
	categoriesCount := len(categories)
	if categoriesCount == 0 {
		return nil
	}

	randomIndex := rand.Intn(categoriesCount) // #nosec G404 -- non-security random selection
	return categories[randomIndex]
}

// GetRandomFile returns a random image file from the list of entries.
// Directories and files with unsupported extensions are excluded. filter may
// be nil to impose no additional constraint beyond the extension check.
func GetRandomFile(files []os.DirEntry, filter *filters.Compiled) (string, error) {
	imageFiles := make([]os.DirEntry, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if _, ok := imageExtensions[ext]; !ok {
			continue
		}
		if filter != nil {
			var info os.FileInfo
			if filter.NeedsFileInfo() {
				fi, err := f.Info()
				if err != nil {
					continue
				}
				info = fi
			}
			if !filter.Matches(f.Name(), info) {
				continue
			}
		}
		imageFiles = append(imageFiles, f)
	}

	if len(imageFiles) == 0 {
		return "", fmt.Errorf("no supported image files found in the directory (.jpg, .jpeg, .png, .webp) matching the configured filter")
	}

	randomIndex := rand.Intn(len(imageFiles)) // #nosec G404 -- non-security random selection
	return imageFiles[randomIndex].Name(), nil
}

// ResolveSource returns the source directory a category should use at time
// now, given the current weather snapshot ws (nil when weather is
// unavailable or not configured) and the named conditions declared in
// configuration.conditions. Plain categories return their source directly.
//
// For a category with variants, every variant whose condition currently
// holds is a candidate; the candidate with the highest priority wins
// (ties broken by position in the variants list). A variant's priority
// comes from its named condition's priority (0 if unset); a variant using
// inline hours has priority 0. ok is false when no variant's condition
// holds, meaning the category is ineligible for this run.
func ResolveSource(cat *models.Categories, now time.Time, ws *weather.Snapshot, conditions map[string]models.Condition) (string, bool) {
	if len(cat.Variants) == 0 {
		return cat.Source, true
	}

	bestIdx := -1
	bestPriority := 0
	for i, v := range cat.Variants {
		holds, priority := variantHolds(v, now, ws, conditions)
		if !holds {
			continue
		}
		if bestIdx == -1 || priority > bestPriority {
			bestIdx = i
			bestPriority = priority
		}
	}
	if bestIdx == -1 {
		return "", false
	}
	return resolveVariantSource(cat, cat.Variants[bestIdx])
}

// variantHolds reports whether v's condition currently holds, and the
// priority to use when comparing against other holding variants.
func variantHolds(v models.Variant, now time.Time, ws *weather.Snapshot, conditions map[string]models.Condition) (holds bool, priority int) {
	if v.Condition != "" {
		cond, ok := conditions[v.Condition]
		if !ok {
			return false, 0
		}
		return conditionHolds(cond, now, ws), cond.Priority
	}
	if v.Hours != "" {
		w, err := schedule.ParseWindow(v.Hours)
		if err != nil {
			return false, 0
		}
		return w.Contains(now), 0
	}
	return false, 0
}

// conditionHolds evaluates a single named condition. A condition holds via
// exactly one of: hours, date-range, or the weather bucket (validation
// enforces this is not mixed); weather-bucket conditions never hold when
// ws is nil.
func conditionHolds(cond models.Condition, now time.Time, ws *weather.Snapshot) bool {
	if cond.Hours != "" {
		w, err := schedule.ParseWindow(cond.Hours)
		if err != nil {
			return false
		}
		return w.Contains(now)
	}

	if cond.DateRange != nil {
		dw, err := schedule.ParseDateRange(cond.DateRange.Start, cond.DateRange.End)
		if err != nil {
			return false
		}
		return dw.Contains(now)
	}

	if ws == nil {
		return false
	}
	hasWeatherFields := len(cond.Weather) > 0 || cond.WindSpeedMin != nil || cond.WindSpeedMax != nil ||
		cond.TemperatureMin != nil || cond.TemperatureMax != nil
	if !hasWeatherFields {
		return false
	}
	if len(cond.Weather) > 0 {
		sky, ok := ws.Sky()
		if !ok {
			return false
		}
		matched := false
		for _, name := range cond.Weather {
			if string(sky) == name {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if cond.WindSpeedMin != nil && ws.WindSpeed < *cond.WindSpeedMin {
		return false
	}
	if cond.WindSpeedMax != nil && ws.WindSpeed > *cond.WindSpeedMax {
		return false
	}
	if cond.TemperatureMin != nil && ws.Temperature < *cond.TemperatureMin {
		return false
	}
	if cond.TemperatureMax != nil && ws.Temperature > *cond.TemperatureMax {
		return false
	}
	return true
}

// resolveVariantSource returns the directory a variant's images live in.
// An absolute source is used as-is; a relative one is resolved against the
// category's source (required in that case — validation enforces this).
func resolveVariantSource(cat *models.Categories, v models.Variant) (string, bool) {
	if filepath.IsAbs(v.Source) {
		return v.Source, true
	}
	if cat.Source == "" {
		return "", false
	}
	return filepath.Join(cat.Source, v.Source), true
}

// SetWallpaperFromFile sets the wallpaper from the specified file.
func SetWallpaperFromFile(source, file string, fade bool) error {
	return SetWallpaperFromPath(filepath.Join(source, file), fade)
}

// SetWallpaperFromPath sets the wallpaper from a pre-built absolute path.
// When fade is true it tries the native Windows crossfade transition first,
// falling back to the instant swap if the fade path fails for any reason.
func SetWallpaperFromPath(fullPath string, fade bool) error {
	if fade {
		if err := setWallpaperFade(fullPath); err == nil {
			return nil
		}
	}
	if err := wallpaper.SetFromFile(fullPath); err != nil {
		return fmt.Errorf("error setting wallpaper: %v", err)
	}
	return nil
}

// GetPreviousWallpaper returns the path of the previous wallpaper.
func GetPreviousWallpaper() (string, error) {
	return wallpaper.Get()
}

// SetWallpaperMode sets the wallpaper mode based on the user's preference.
// On Windows this applies the mode via IDesktopWallpaper directly, which
// avoids the legacy registry write + reapply that would otherwise stomp on
// SetWallpaperFromPath's fade transition; other platforms fall back to the
// standard behavior.
func SetWallpaperMode(mode string) error {
	if err := setWallpaperPosition(mode); err == nil {
		return nil
	}
	switch mode {
	case "center":
		return wallpaper.SetMode(wallpaper.Center)
	case "fit":
		return wallpaper.SetMode(wallpaper.Fit)
	case "span":
		return wallpaper.SetMode(wallpaper.Span)
	case "stretch":
		return wallpaper.SetMode(wallpaper.Stretch)
	case "tile":
		return wallpaper.SetMode(wallpaper.Tile)
	case "crop":
		fallthrough
	default:
		return wallpaper.SetMode(wallpaper.Crop)
	}
}
