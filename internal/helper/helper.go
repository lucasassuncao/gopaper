package helper

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/lucasassuncao/gopaper/internal/models"

	"github.com/reujab/wallpaper"
)

// CreateDirectory checks if the specified directory exists, and if not, creates it with full permissions.
func CreateDirectory(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0777)
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

	randomIndex := rand.Intn(categoriesCount)
	return categories[randomIndex]
}

// GetRandomFile returns a random file from the list of files.
func GetRandomFile(files []os.DirEntry) (string, error) {
	filesCount := len(files)

	switch filesCount {
	case 0:
		return "", fmt.Errorf("no files found in the directory")
	default:
		randomIndex := rand.Intn(filesCount)
		return files[randomIndex].Name(), nil
	}
}

// SetWallpaperFromFile sets the wallpaper from the specified file.
func SetWallpaperFromFile(source, file string) error {
	err := wallpaper.SetFromFile(filepath.Join(source, file))
	if err != nil {
		return fmt.Errorf("error setting wallpaper: %v", err)
	}
	return nil
}

// GetPreviousWallpaper returns the path of the previous wallpaper.
func GetPreviousWallpaper() (string, error) {
	return wallpaper.Get()
}

// SetWallpaperMode sets the wallpaper mode based on the user's preference.
func SetWallpaperMode(mode string) {
	switch mode {
	case "center":
		wallpaper.SetMode(wallpaper.Center)
	case "fit":
		wallpaper.SetMode(wallpaper.Fit)
	case "span":
		wallpaper.SetMode(wallpaper.Span)
	case "stretch":
		wallpaper.SetMode(wallpaper.Stretch)
	case "tile":
		wallpaper.SetMode(wallpaper.Tile)
	case "crop":
		fallthrough
	default:
		wallpaper.SetMode(wallpaper.Crop)
	}
}
