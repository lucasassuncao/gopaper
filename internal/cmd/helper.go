package cmd

import (
	"fmt"
	"gopaper/internal/models"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/reujab/wallpaper"
)

// createDirectory checks if the specified directory exists, and if not, creates it with full permissions.
func createDirectory(dir string) error {
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

// readDirectory reads the contents of a given directory and returns the files.
func readDirectory(path string) ([]os.DirEntry, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// getEnabledCategories returns a list of enabled categories from the list of categories.
func getEnabledCategories(categories []*models.Categories) []*models.Categories {
	var enabledCategories []*models.Categories
	for _, category := range categories {
		if category.Enabled {
			enabledCategories = append(enabledCategories, category)
		}
	}
	return enabledCategories
}

// getRandomCategory returns a random category from the list of categories.
func getRandomCategory(categories []*models.Categories) *models.Categories {
	categoriesCount := len(categories)
	if categoriesCount == 0 {
		return nil
	}

	randomIndex := rand.Intn(categoriesCount)
	return categories[randomIndex]
}

// getRandomFile returns a random file from the list of files.
func getRandomFile(files []os.DirEntry) (string, error) {
	filesCount := len(files)

	switch filesCount {
	case 0:
		return "", fmt.Errorf("no files found in the directory")
	default:
		randomIndex := rand.Intn(filesCount)
		return files[randomIndex].Name(), nil
	}
}

// setWallpaperFromFile sets the wallpaper from the specified file.
func setWallpaperFromFile(source, file string) error {
	err := wallpaper.SetFromFile(filepath.Join(source, file))
	if err != nil {
		return fmt.Errorf("error setting wallpaper: %v", err)
	}
	return nil
}

func getPreviousWallpaper() (string, error) {
	return wallpaper.Get()
}

// setWallpaperMode sets the wallpaper mode based on the user's preference.
func setWallpaperMode(mode string) {
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
