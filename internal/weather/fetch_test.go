package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func withTestServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	prevURL := apiBaseURL
	apiBaseURL = server.URL
	t.Cleanup(func() { apiBaseURL = prevURL })
}

func jsonWeatherHandler(code int, wind, temp float64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp apiResponse
		resp.Current.WeatherCode = code
		resp.Current.WindSpeed = wind
		resp.Current.Temperature = temp
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func TestFetchLiveSuccess(t *testing.T) {
	withTestServer(t, jsonWeatherHandler(61, 12.5, 22.3))
	cachePath := filepath.Join(t.TempDir(), "weather-cache.json")

	snap, err := Fetch(Config{Latitude: -23.55, Longitude: -46.63, CacheTTL: time.Minute}, cachePath)
	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}
	if snap.Code != 61 || snap.WindSpeed != 12.5 || snap.Temperature != 22.3 {
		t.Errorf("got %+v, want Code=61 WindSpeed=12.5 Temperature=22.3", snap)
	}
}

func TestFetchUsesFreshCacheWithoutHTTPCall(t *testing.T) {
	calls := 0
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		jsonWeatherHandler(95, 40, 18.0)(w, r)
	})
	cachePath := filepath.Join(t.TempDir(), "weather-cache.json")
	cfg := Config{Latitude: 1, Longitude: 2, CacheTTL: time.Hour}

	if _, err := Fetch(cfg, cachePath); err != nil {
		t.Fatalf("first fetch error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 HTTP call after first fetch, got %d", calls)
	}

	snap, err := Fetch(cfg, cachePath)
	if err != nil {
		t.Fatalf("second fetch error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected cache hit (still 1 HTTP call), got %d", calls)
	}
	if snap.Code != 95 || snap.Temperature != 18.0 {
		t.Errorf("got %+v from cache, want Code=95 Temperature=18.0", snap)
	}
}

func TestFetchRefetchesAfterTTLExpires(t *testing.T) {
	calls := 0
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		jsonWeatherHandler(0, 5, 25.0)(w, r)
	})
	cachePath := filepath.Join(t.TempDir(), "weather-cache.json")
	cfg := Config{Latitude: 1, Longitude: 2, CacheTTL: time.Millisecond}

	if _, err := Fetch(cfg, cachePath); err != nil {
		t.Fatalf("first fetch error: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if _, err := Fetch(cfg, cachePath); err != nil {
		t.Fatalf("second fetch error: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 HTTP calls after TTL expiry, got %d", calls)
	}
}

func TestFetchFallsBackToStaleCacheOnAPIFailure(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "weather-cache.json")
	cfg := Config{Latitude: 1, Longitude: 2, CacheTTL: time.Millisecond}

	withTestServer(t, jsonWeatherHandler(3, 8, 16.5))
	if _, err := Fetch(cfg, cachePath); err != nil {
		t.Fatalf("priming fetch error: %v", err)
	}
	time.Sleep(5 * time.Millisecond) // expire the cache TTL

	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	snap, err := Fetch(cfg, cachePath)
	if err != nil {
		t.Fatalf("expected fallback to stale cache, got error: %v", err)
	}
	if snap.Code != 3 || snap.Temperature != 16.5 {
		t.Errorf("got %+v, want Code=3 Temperature=16.5 (from stale cache)", snap)
	}
}

func TestFetchReturnsErrorWithNoCacheAndAPIFailure(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	cachePath := filepath.Join(t.TempDir(), "does-not-exist", "weather-cache.json")
	if _, err := os.Stat(filepath.Dir(cachePath)); err == nil {
		t.Fatal("test setup: cache dir should not exist")
	}

	_, err := Fetch(Config{Latitude: 1, Longitude: 2, CacheTTL: time.Minute}, cachePath)
	if err == nil {
		t.Fatal("expected error when API fails and no cache exists")
	}
}
