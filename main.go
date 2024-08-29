package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/alyu/configparser"
	"github.com/logrusorgru/aurora/v4"
	"github.com/reujab/wallpaper"
)

const configFile = "D:\\Scripts\\go\\conf\\wallpaperChanger.ini"

func getConfSections() (*configparser.Configuration, []string) {
	configparser.Delimiter = "="

	config, err := configparser.Read(configFile)
	if err != nil {
		logger.Errorf("Failed to read configuration file: %s. Error: %s", configFile, err)
		return nil, nil
	}

	allsections, err := config.AllSections()
	if err != nil {
		logger.Errorf("Failed to read sections in the configuration file. Error: %s", err)
		return nil, nil
	}

	var availableSectionNames []string
	for _, section := range allsections {
		if !strings.HasPrefix(section.Name(), "#") && !strings.Contains(section.Name(), "global") {
			availableSectionNames = append(availableSectionNames, section.Name())
		}
	}
	return config, availableSectionNames
}

func countFilesInSections(config *configparser.Configuration, sectionNames []string) (map[string]int, error) {
	fileCounts := make(map[string]int)
	extensions := []string{".jpg", ".png"}

	for _, sectionName := range sectionNames {
		section, err := config.Section(sectionName)
		if err != nil {
			return nil, fmt.Errorf("failed to get section %s: %w", sectionName, err)
		}

		src := section.ValueOf("source")
		files, err := os.ReadDir(src)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory %s: %w", src, err)
		}

		for _, file := range files {
			if file.Type().IsRegular() && isValidExtension(file.Name(), extensions) {
				fileCounts[sectionName]++
			}
		}
	}

	return fileCounts, nil
}

func isValidExtension(fileName string, extensions []string) bool {
	for _, ext := range extensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}
	return false
}

func getRandomFileFromSection(config *configparser.Configuration, sectionName string) (string, error) {
	section, err := config.Section(sectionName)
	if err != nil {
		return "", fmt.Errorf("failed to get section %s: %w", sectionName, err)
	}

	src := section.ValueOf("source")
	files, err := os.ReadDir(src)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", src, err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no files in section %s", sectionName)
	}

	retries := 0
	for {
		randomIndex := rand.Intn(len(files))
		if !files[randomIndex].IsDir() {
			return filepath.Join(src, files[randomIndex].Name()), nil
		}

		retries++
		if retries >= 5 {
			return wallpaper.Get()
		}
	}
}

func main() {
	logger = GetLogger("main")

	config, sectionNames := getConfSections()
	if config == nil || len(sectionNames) == 0 {
		fmt.Println("")
		logger.Warning(aurora.Black("There are no sections to look for wallpapers. Exiting...").BgBrightWhite())
		fmt.Println("")
		os.Exit(0)
	}

	fileCounts, err := countFilesInSections(config, sectionNames)
	if err != nil {
		logger.Errorf("Error counting files: %s", err)
		os.Exit(1)
	}

	for sectionName, count := range fileCounts {
		logger.Infof("Section %v has %v files", aurora.Blue(sectionName), aurora.Blue(count))
	}
	fmt.Println("")

	randomSection := sectionNames[rand.Intn(len(sectionNames))]
	if count, ok := fileCounts[randomSection]; ok && count > 0 {
		wallpaperPath, err := getRandomFileFromSection(config, randomSection)
		if err != nil {
			logger.Errorf("Error getting random wallpaper: %s", err)
			os.Exit(1)
		}

		logger.Infof("Random Section Name: %v", randomSection)
		background, err := wallpaper.Get()
		if err != nil {
			logger.Errorf("Failed to get current wallpaper. Error: %s", err)
		}
		logger.Infof("Current wallpaper: %s", background)

		err = wallpaper.SetFromFile(wallpaperPath)
		if err != nil {
			logger.Errorf("Failed to set new wallpaper. Error: %s", err)
		}
		logger.Infof("New wallpaper set: %s", wallpaperPath)
		wallpaper.SetMode(wallpaper.Crop)
	} else {
		logger.Infof("There are no files in the selected section: %s", randomSection)
	}

	fmt.Println("")
	fmt.Println("Thanks for using this software ❤️")
	fmt.Println("")
}
