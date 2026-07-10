package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var (
	apiBaseURL = "https://api.open-meteo.com/v1/forecast"
	httpClient = &http.Client{Timeout: 5 * time.Second}
)

// Snapshot is a point-in-time weather reading.
type Snapshot struct {
	Code        int     // WMO weather code
	WindSpeed   float64 // km/h
	Temperature float64 // Celsius
}

// Sky maps the snapshot's code to a Sky category. ok is false for an
// unrecognized code.
func (s Snapshot) Sky() (Sky, bool) {
	return CodeToSky(s.Code)
}

// Config is the location and cache policy used to fetch a Snapshot.
type Config struct {
	Latitude  float64
	Longitude float64
	CacheTTL  time.Duration
}

type cacheEntry struct {
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Code        int       `json:"code"`
	WindSpeed   float64   `json:"wind_speed"`
	Temperature float64   `json:"temperature"`
	FetchedAt   time.Time `json:"fetched_at"`
}

// Fetch returns the current weather for cfg's location, using the cache at
// cachePath when it is fresh (within cfg.CacheTTL) and for the same
// location. On a live-fetch failure it falls back to a stale cache entry
// for the same location if one exists; only when there is neither a fresh
// fetch nor any usable cache does it return an error.
func Fetch(cfg Config, cachePath string) (Snapshot, error) {
	if entry, ok := readCache(cachePath); ok && sameLocation(entry, cfg) && time.Since(entry.FetchedAt) < cfg.CacheTTL {
		return Snapshot{Code: entry.Code, WindSpeed: entry.WindSpeed, Temperature: entry.Temperature}, nil
	}

	snap, err := fetchLive(cfg)
	if err != nil {
		if entry, ok := readCache(cachePath); ok && sameLocation(entry, cfg) {
			return Snapshot{Code: entry.Code, WindSpeed: entry.WindSpeed, Temperature: entry.Temperature}, nil
		}
		return Snapshot{}, err
	}

	writeCache(cachePath, cacheEntry{
		Latitude:    cfg.Latitude,
		Longitude:   cfg.Longitude,
		Code:        snap.Code,
		WindSpeed:   snap.WindSpeed,
		Temperature: snap.Temperature,
		FetchedAt:   time.Now(),
	})
	return snap, nil
}

func sameLocation(e cacheEntry, cfg Config) bool {
	return e.Latitude == cfg.Latitude && e.Longitude == cfg.Longitude
}

type apiResponse struct {
	Current struct {
		WeatherCode int     `json:"weather_code"`
		WindSpeed   float64 `json:"wind_speed_10m"`
		Temperature float64 `json:"temperature_2m"`
	} `json:"current"`
}

func fetchLive(cfg Config) (Snapshot, error) {
	url := fmt.Sprintf("%s?latitude=%g&longitude=%g&current=weather_code,wind_speed_10m,temperature_2m", apiBaseURL, cfg.Latitude, cfg.Longitude)
	resp, err := httpClient.Get(url) // #nosec G107 -- URL is built from validated configuration.weather lat/long, not user input
	if err != nil {
		return Snapshot{}, fmt.Errorf("weather request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Snapshot{}, fmt.Errorf("weather request failed: HTTP %d", resp.StatusCode)
	}

	var body apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Snapshot{}, fmt.Errorf("could not parse weather response: %w", err)
	}
	return Snapshot{Code: body.Current.WeatherCode, WindSpeed: body.Current.WindSpeed, Temperature: body.Current.Temperature}, nil
}

func readCache(path string) (cacheEntry, bool) {
	data, err := os.ReadFile(path) // #nosec G304 -- path comes from config.WeatherCachePath(), not user input
	if err != nil {
		return cacheEntry{}, false
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return cacheEntry{}, false
	}
	return entry, true
}

func writeCache(path string, entry cacheEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o600)
}
