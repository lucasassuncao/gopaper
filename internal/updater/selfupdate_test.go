package updater

import (
	"strings"
	"testing"
)

func TestIsHexSHA256(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"valid lowercase", "8bc3c68d94a3e4de6ea921270c169243da6fd46dedbff8f9608541e7390f4c4b", true},
		{"valid uppercase", "8BC3C68D94A3E4DE6EA921270C169243DA6FD46DEDBFF8F9608541E7390F4C4B", true},
		{"too short", "abc123", false},
		{"too long", "8bc3c68d94a3e4de6ea921270c169243da6fd46dedbff8f9608541e7390f4c4b00", false},
		{"non-hex characters", strings.Repeat("g", 64), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isHexSHA256(tt.in); got != tt.want {
				t.Errorf("isHexSHA256(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestFindChecksumInManifest(t *testing.T) {
	const hash = "8bc3c68d94a3e4de6ea921270c169243da6fd46dedbff8f9608541e7390f4c4b"

	tests := []struct {
		name     string
		body     string
		fileName string
		want     string
	}{
		{
			name:     "sha256sum format with matching filename",
			body:     hash + "  gopaper_linux_amd64\n" + strings.Repeat("0", 64) + "  other_file\n",
			fileName: "gopaper_linux_amd64",
			want:     hash,
		},
		{
			name:     "sha256sum format with binary-mode asterisk",
			body:     hash + " *gopaper_linux_amd64\n",
			fileName: "gopaper_linux_amd64",
			want:     hash,
		},
		{
			name:     "no matching filename",
			body:     hash + "  some_other_asset\n",
			fileName: "gopaper_linux_amd64",
			want:     "",
		},
		{
			name:     "bare digest (per-asset manifest)",
			body:     hash + "\n",
			fileName: "gopaper_linux_amd64",
			want:     hash,
		},
		{
			name:     "empty manifest",
			body:     "",
			fileName: "gopaper_linux_amd64",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findChecksumInManifest([]byte(tt.body), tt.fileName); got != tt.want {
				t.Errorf("findChecksumInManifest(...) = %q, want %q", got, tt.want)
			}
		})
	}
}
