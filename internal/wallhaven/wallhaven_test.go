package wallhaven

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// newServer stubs both the search endpoint and the image CDN. imageID and
// imageBody control what the single search result points at.
func newServer(t *testing.T, imageID, imageBody string) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			fmt.Fprintf(w, `{"data":[{"id":%q,"path":"%s/img/%s.jpg"}]}`, imageID, srv.URL, imageID)
		case "/img/" + imageID + ".jpg":
			fmt.Fprint(w, imageBody)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func withAPIBase(t *testing.T, base string) {
	t.Helper()
	old := apiBaseURL
	apiBaseURL = base
	t.Cleanup(func() { apiBaseURL = old })
}

func TestRefreshDownloadsIntoCache(t *testing.T) {
	srv := newServer(t, "abc123", "fake-image-bytes")
	withAPIBase(t, srv.URL+"/")

	dir := filepath.Join(t.TempDir(), "cache")
	if err := Refresh(Config{Query: "landscape"}, dir); err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "abc123.jpg"))
	if err != nil {
		t.Fatalf("expected abc123.jpg in cache: %v", err)
	}
	if string(data) != "fake-image-bytes" {
		t.Errorf("unexpected image content: %q", data)
	}
}

func TestRefreshDuplicateIDIsNoOp(t *testing.T) {
	srv := newServer(t, "abc123", "new-bytes")
	withAPIBase(t, srv.URL+"/")

	dir := t.TempDir()
	existing := filepath.Join(dir, "abc123.jpg")
	if err := os.WriteFile(existing, []byte("old-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := Refresh(Config{Query: "landscape"}, dir); err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}
	data, _ := os.ReadFile(existing)
	if string(data) != "old-bytes" {
		t.Errorf("cached file was overwritten: %q", data)
	}
}

func TestRefreshNetworkFailureKeepsCache(t *testing.T) {
	withAPIBase(t, "http://127.0.0.1:1/") // nothing listening

	dir := t.TempDir()
	existing := filepath.Join(dir, "cached.jpg")
	if err := os.WriteFile(existing, []byte("keep-me"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := Refresh(Config{Query: "landscape"}, dir)
	if err == nil {
		t.Fatal("expected an error from the failed fetch")
	}
	if _, statErr := os.Stat(existing); statErr != nil {
		t.Errorf("existing cache file was disturbed: %v", statErr)
	}
}

func TestRefreshPrunesOldestBeyondLimit(t *testing.T) {
	srv := newServer(t, "new001", "img")
	withAPIBase(t, srv.URL+"/")

	dir := t.TempDir()
	// Two pre-existing files with distinct mtimes; limit 2 means after
	// downloading the third, the oldest must go.
	oldest := filepath.Join(dir, "oldest.jpg")
	newer := filepath.Join(dir, "newer.jpg")
	for _, f := range []string{oldest, newer} {
		if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	past := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldest, past, past); err != nil {
		t.Fatal(err)
	}

	if err := Refresh(Config{Query: "landscape", CacheLimit: 2}, dir); err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	if _, err := os.Stat(oldest); !os.IsNotExist(err) {
		t.Error("oldest.jpg should have been pruned")
	}
	for _, f := range []string{newer, filepath.Join(dir, "new001.jpg")} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("%s should have been kept: %v", filepath.Base(f), err)
		}
	}
}

func TestRefreshEmptyResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[]}`)
	}))
	t.Cleanup(srv.Close)
	withAPIBase(t, srv.URL+"/")

	err := Refresh(Config{Query: "nothing-matches-this"}, t.TempDir())
	if err == nil {
		t.Fatal("expected an error for empty results")
	}
}

func TestPurityCode(t *testing.T) {
	cases := map[string]string{"": "100", "sfw": "100", "sketchy": "010", "nsfw": "001"}
	for in, want := range cases {
		got, err := purityCode(in)
		if err != nil || got != want {
			t.Errorf("purityCode(%q) = (%q, %v), want (%q, nil)", in, got, err, want)
		}
	}
	if _, err := purityCode("banana"); err == nil {
		t.Error("expected an error for unknown purity")
	}
}
