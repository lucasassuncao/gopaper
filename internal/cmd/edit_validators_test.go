package cmd

import (
	"strings"
	"testing"

	"github.com/lucasassuncao/gopaper/internal/models"
	"github.com/lucasassuncao/yedit/editor"
)

// runValidators runs the full GopaperValidators rule set on a raw YAML doc
// and returns every violation message joined for easy contains-checks.
func runValidators(t *testing.T, raw string) []editor.Violation {
	t.Helper()
	hints, err := buildGopaperHints()
	if err != nil {
		t.Fatalf("building hints: %v", err)
	}
	wired := editor.Wire(GopaperValidators, editor.Config{Schema: &models.Config{}, Metadata: hints})
	return editor.RunAll(wired, []byte(raw), nil)
}

func hasViolation(violations []editor.Violation, pathPart, msgPart string) bool {
	for _, v := range violations {
		if strings.Contains(v.Path, pathPart) && strings.Contains(v.Message, msgPart) {
			return true
		}
	}
	return false
}

const validBase = `
configuration:
  logging:
    output: console
    level: info
categories:
`

func TestValidateWallhavenMutuallyExclusiveWithSource(t *testing.T) {
	raw := validBase + `
  - name: "WH"
    source: "C:\\walls"
    wallhaven:
      query: "landscape"
    mode: crop
    enabled: true
`
	vs := runValidators(t, raw)
	if !hasViolation(vs, "wallhaven", "mutually exclusive") {
		t.Errorf("expected a mutual-exclusivity violation, got: %+v", vs)
	}
}

func TestValidateWallhavenQueryRequired(t *testing.T) {
	raw := validBase + `
  - name: "WH"
    wallhaven:
      purity: sfw
    mode: crop
    enabled: true
`
	vs := runValidators(t, raw)
	if !hasViolation(vs, "wallhaven.query", "required") {
		t.Errorf("expected a missing-query violation, got: %+v", vs)
	}
}

func TestValidateWallhavenPurityNeedsAPIKey(t *testing.T) {
	raw := validBase + `
  - name: "WH"
    wallhaven:
      query: "landscape"
      purity: nsfw
    mode: crop
    enabled: true
`
	vs := runValidators(t, raw)
	if !hasViolation(vs, "wallhaven.purity", "api-key") {
		t.Errorf("expected a purity-needs-api-key violation, got: %+v", vs)
	}
}

func TestValidateWallhavenPurityWithAPIKeyOK(t *testing.T) {
	raw := `
configuration:
  logging:
    output: console
    level: info
  wallhaven:
    api-key: "abc"
categories:
  - name: "WH"
    wallhaven:
      query: "landscape"
      purity: nsfw
    mode: crop
    enabled: true
`
	vs := runValidators(t, raw)
	if hasViolation(vs, "wallhaven.purity", "api-key") {
		t.Errorf("did not expect a purity violation with an api-key set, got: %+v", vs)
	}
}

func TestValidateWallhavenAloneIsValidSourceShape(t *testing.T) {
	raw := validBase + `
  - name: "WH"
    wallhaven:
      query: "landscape"
    mode: crop
    enabled: true
`
	vs := runValidators(t, raw)
	if hasViolation(vs, "categories[0].source", "define one of") {
		t.Errorf("wallhaven alone should satisfy the source shape, got: %+v", vs)
	}
}

func TestValidateCategoryWithNoSourceAtAll(t *testing.T) {
	raw := validBase + `
  - name: "Empty"
    mode: crop
    enabled: true
`
	vs := runValidators(t, raw)
	if !hasViolation(vs, "categories[0].source", "define one of source, variants, or wallhaven") {
		t.Errorf("expected the three-way shape violation, got: %+v", vs)
	}
}
