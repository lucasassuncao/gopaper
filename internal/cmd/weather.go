package cmd

import (
	"time"

	"github.com/lucasassuncao/gopaper/internal/config"
	"github.com/lucasassuncao/gopaper/internal/models"
	"github.com/lucasassuncao/gopaper/internal/weather"
)

// fetchWeatherSnapshot returns the current weather snapshot for use by
// weather-based conditions, or nil when configuration.weather is not set or
// the fetch/cache both fail. Weather is always best-effort: a failure here
// only means weather-based variants are skipped this run, it never aborts
// the wallpaper change.
func fetchWeatherSnapshot(g *models.Gopaper) *weather.Snapshot {
	weatherCfg, err := config.LoadWeatherConfig(g.Viper)
	if err != nil {
		g.Logger.Warn("invalid weather configuration, weather-based variants will be skipped", g.Logger.Args("error", err))
		return nil
	}
	if weatherCfg == nil {
		return nil
	}

	cachePath, err := config.WeatherCachePath()
	if err != nil {
		g.Logger.Warn("could not determine weather cache path, weather-based variants will be skipped", g.Logger.Args("error", err))
		return nil
	}

	ttl, err := time.ParseDuration(weatherCfg.CacheTTL)
	if err != nil {
		ttl = 15 * time.Minute
	}

	snap, err := weather.Fetch(weather.Config{
		Latitude:  weatherCfg.Latitude,
		Longitude: weatherCfg.Longitude,
		CacheTTL:  ttl,
	}, cachePath)
	if err != nil {
		g.Logger.Warn("could not fetch weather, weather-based variants will be skipped", g.Logger.Args("error", err))
		return nil
	}
	return &snap
}
