// Package wallhaven fetches wallpapers from the Wallhaven API into a local
// cache directory, which then acts as a regular category source. Like the
// weather package, it is best-effort: a network failure never breaks a
// category that already has cached images.
package wallhaven

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const defaultCacheLimit = 100

var (
	apiBaseURL = "https://wallhaven.cc/api/v1/search"
	httpClient = &http.Client{Timeout: 10 * time.Second}
)

// Config is one refresh request's parameters.
type Config struct {
	Query      string
	Purity     string // sfw (default) | sketchy | nsfw
	APIKey     string
	CacheLimit int // 0 = default (100)
}

// purityCode maps the config purity tier to Wallhaven's three-bit purity
// parameter (single tier, not cumulative).
func purityCode(purity string) (string, error) {
	switch purity {
	case "", "sfw":
		return "100", nil
	case "sketchy":
		return "010", nil
	case "nsfw":
		return "001", nil
	default:
		return "", fmt.Errorf("unknown purity %q", purity)
	}
}

// Refresh fetches one new image matching cfg into cacheDir (creating it if
// needed) and prunes cacheDir down to cfg.CacheLimit (oldest first by
// modification time). Failures leave the existing cache untouched, so a
// category with a populated cache keeps working offline; the returned error
// is for the caller to log.
func Refresh(cfg Config, cacheDir string) error {
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return fmt.Errorf("could not create wallhaven cache directory: %w", err)
	}

	id, imageURL, err := searchRandom(cfg)
	if err != nil {
		return err
	}

	if cached, err := hasImage(cacheDir, id); err != nil {
		return err
	} else if !cached {
		if err := download(imageURL, filepath.Join(cacheDir, id+strings.ToLower(filepath.Ext(imageURL)))); err != nil {
			return err
		}
	}

	return prune(cacheDir, cfg.CacheLimit)
}

type searchResponse struct {
	Data []struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	} `json:"data"`
}

// searchRandom asks the API for a random page of results matching cfg and
// picks one, returning its ID and full image URL.
func searchRandom(cfg Config) (id, imageURL string, err error) {
	code, err := purityCode(cfg.Purity)
	if err != nil {
		return "", "", err
	}

	params := url.Values{}
	params.Set("q", cfg.Query)
	params.Set("purity", code)
	params.Set("sorting", "random")
	if cfg.APIKey != "" {
		params.Set("apikey", cfg.APIKey)
	}

	resp, err := httpClient.Get(apiBaseURL + "?" + params.Encode()) // #nosec G107 -- base URL is a package constant; params are config values, the same trust boundary as the rest of the config
	if err != nil {
		return "", "", fmt.Errorf("wallhaven request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("wallhaven request failed: HTTP %d", resp.StatusCode)
	}

	var body searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", "", fmt.Errorf("could not parse wallhaven response: %w", err)
	}
	if len(body.Data) == 0 {
		return "", "", fmt.Errorf("wallhaven returned no results for query %q", cfg.Query)
	}

	pick := body.Data[rand.Intn(len(body.Data))] // #nosec G404 -- non-security random selection
	if pick.ID == "" || pick.Path == "" {
		return "", "", fmt.Errorf("wallhaven returned a result without id/path")
	}
	return pick.ID, pick.Path, nil
}

// hasImage reports whether an image with this Wallhaven ID is already
// cached (matched by filename without extension).
func hasImage(cacheDir, id string) (bool, error) {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return false, fmt.Errorf("could not read wallhaven cache directory: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())) == id {
			return true, nil
		}
	}
	return false, nil
}

// download saves the image at imageURL to dest, writing to a temp file
// first so a partial download never leaves a corrupt image in the cache.
func download(imageURL, dest string) error {
	resp, err := httpClient.Get(imageURL) // #nosec G107 -- URL comes from the wallhaven API response for a request we made
	if err != nil {
		return fmt.Errorf("wallhaven image download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wallhaven image download failed: HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dest), ".wallhaven-*")
	if err != nil {
		return fmt.Errorf("could not create temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.ReadFrom(resp.Body); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("could not write image: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("could not write image: %w", err)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("could not move image into cache: %w", err)
	}
	return nil
}

// prune removes the oldest files (by modification time) until the cache
// holds at most limit images.
func prune(cacheDir string, limit int) error {
	if limit <= 0 {
		limit = defaultCacheLimit
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return fmt.Errorf("could not read wallhaven cache directory: %w", err)
	}

	type fileAge struct {
		name string
		mod  time.Time
	}
	var files []fileAge
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileAge{name: e.Name(), mod: info.ModTime()})
	}
	if len(files) <= limit {
		return nil
	}

	sort.Slice(files, func(i, j int) bool { return files[i].mod.Before(files[j].mod) })
	for _, f := range files[:len(files)-limit] {
		if err := os.Remove(filepath.Join(cacheDir, f.name)); err != nil {
			return fmt.Errorf("could not prune wallhaven cache: %w", err)
		}
	}
	return nil
}
