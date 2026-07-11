package helper

import (
	"testing"
	"time"

	"github.com/lucasassuncao/gopaper/internal/models"
	"github.com/lucasassuncao/gopaper/internal/weather"
)

func TestResolveSourcePlainCategory(t *testing.T) {
	cat := &models.Categories{Source: `C:\walls\plain`}
	src, ok := ResolveSource(cat, time.Now(), nil, nil, "")
	if !ok || src != `C:\walls\plain` {
		t.Fatalf("got (%q, %v), want (C:\\walls\\plain, true)", src, ok)
	}
}

func TestResolveSourceWallhaven(t *testing.T) {
	cat := &models.Categories{Wallhaven: &models.WallhavenSource{Query: "landscape"}}

	src, ok := ResolveSource(cat, time.Now(), nil, nil, `C:\cache\wh`)
	if !ok || src != `C:\cache\wh` {
		t.Errorf("got (%q, %v), want the wallhaven cache dir", src, ok)
	}

	// No resolved cache dir (e.g. the path lookup failed) → ineligible.
	if _, ok := ResolveSource(cat, time.Now(), nil, nil, ""); ok {
		t.Error("wallhaven category without a cache dir should be ineligible")
	}
}

func TestResolveSourceInlineHours(t *testing.T) {
	cat := &models.Categories{Variants: []models.Variant{
		{Source: `C:\walls\day`, Hours: "06:00-17:59"},
		{Source: `C:\walls\night`, Hours: "18:00-05:59"},
	}}

	noon := time.Date(2026, 7, 10, 12, 0, 0, 0, time.Local)
	if src, ok := ResolveSource(cat, noon, nil, nil, ""); !ok || src != `C:\walls\day` {
		t.Errorf("noon: got (%q, %v), want day variant", src, ok)
	}

	night := time.Date(2026, 7, 10, 23, 0, 0, 0, time.Local)
	if src, ok := ResolveSource(cat, night, nil, nil, ""); !ok || src != `C:\walls\night` {
		t.Errorf("23:00: got (%q, %v), want night variant", src, ok)
	}
}

func TestResolveSourceNoActiveVariant(t *testing.T) {
	cat := &models.Categories{Variants: []models.Variant{
		{Source: `C:\walls\morning`, Hours: "08:00-08:59"},
	}}
	tenPM := time.Date(2026, 7, 10, 22, 0, 0, 0, time.Local)
	if src, ok := ResolveSource(cat, tenPM, nil, nil, ""); ok {
		t.Errorf("got (%q, %v), want ineligible (ok=false)", src, ok)
	}
}

func TestResolveSourceNamedCondition(t *testing.T) {
	conditions := map[string]models.Condition{
		"morning": {Hours: "06:00-11:59"},
		"night":   {Hours: "22:00-05:59"},
	}
	cat := &models.Categories{
		Source: `C:\walls\dynamic`,
		Variants: []models.Variant{
			{Source: `./day`, Condition: "morning"},
			{Source: `./night`, Condition: "night"},
		},
	}

	morning := time.Date(2026, 7, 10, 8, 0, 0, 0, time.Local)
	src, ok := ResolveSource(cat, morning, nil, conditions, "")
	if !ok || src != `C:\walls\dynamic\day` {
		t.Errorf("got (%q, %v), want (C:\\walls\\dynamic\\day, true)", src, ok)
	}
}

func TestResolveSourceRelativeSourceWithoutCategorySourceIsIneligible(t *testing.T) {
	conditions := map[string]models.Condition{"morning": {Hours: "00:00-23:59"}}
	cat := &models.Categories{
		Variants: []models.Variant{{Source: "./day", Condition: "morning"}},
	}
	if _, ok := ResolveSource(cat, time.Now(), nil, conditions, ""); ok {
		t.Error("expected ineligible when variant source is relative and category has no source")
	}
}

func TestResolveSourceWeatherPriorityBeatsHours(t *testing.T) {
	conditions := map[string]models.Condition{
		"afternoon": {Hours: "12:00-17:59", Priority: 0},
		"rainy":     {Weather: []string{"rain"}, Priority: 10},
	}
	cat := &models.Categories{Variants: []models.Variant{
		{Source: `C:\walls\afternoon`, Condition: "afternoon"},
		{Source: `C:\walls\rainy`, Condition: "rainy"},
	}}

	threePM := time.Date(2026, 7, 10, 15, 0, 0, 0, time.Local)
	ws := &weather.Snapshot{Code: 61} // rain
	src, ok := ResolveSource(cat, threePM, ws, conditions, "")
	if !ok || src != `C:\walls\rainy` {
		t.Errorf("got (%q, %v), want rainy variant to win on priority", src, ok)
	}
}

func TestResolveSourceNoWeatherSnapshotSkipsWeatherConditions(t *testing.T) {
	conditions := map[string]models.Condition{
		"afternoon": {Hours: "12:00-17:59"},
		"rainy":     {Weather: []string{"rain"}, Priority: 10},
	}
	cat := &models.Categories{Variants: []models.Variant{
		{Source: `C:\walls\afternoon`, Condition: "afternoon"},
		{Source: `C:\walls\rainy`, Condition: "rainy"},
	}}

	threePM := time.Date(2026, 7, 10, 15, 0, 0, 0, time.Local)
	src, ok := ResolveSource(cat, threePM, nil, conditions, "")
	if !ok || src != `C:\walls\afternoon` {
		t.Errorf("got (%q, %v), want afternoon variant (no weather data available)", src, ok)
	}
}

func TestResolveSourceWindSpeedThreshold(t *testing.T) {
	min30 := 30.0
	conditions := map[string]models.Condition{
		"windy": {WindSpeedMin: &min30, Priority: 10},
	}
	cat := &models.Categories{Variants: []models.Variant{
		{Source: `C:\walls\windy`, Condition: "windy"},
	}}

	calm := &weather.Snapshot{WindSpeed: 10}
	if _, ok := ResolveSource(cat, time.Now(), calm, conditions, ""); ok {
		t.Error("expected windy condition to not hold at 10 km/h with min 30")
	}

	gusty := &weather.Snapshot{WindSpeed: 35}
	if src, ok := ResolveSource(cat, time.Now(), gusty, conditions, ""); !ok || src != `C:\walls\windy` {
		t.Errorf("got (%q, %v), want windy variant to hold at 35 km/h", src, ok)
	}
}

func floatPtr(f float64) *float64 { return &f }

func TestResolveSourceDateRangeSameYear(t *testing.T) {
	conditions := map[string]models.Condition{
		"christmas": {DateRange: &models.DateRange{Start: "12-24", End: "12-26"}},
	}
	cat := &models.Categories{Variants: []models.Variant{
		{Source: `C:\walls\christmas`, Condition: "christmas"},
	}}

	inRange := time.Date(2026, 12, 25, 10, 0, 0, 0, time.Local)
	if src, ok := ResolveSource(cat, inRange, nil, conditions, ""); !ok || src != `C:\walls\christmas` {
		t.Errorf("got (%q, %v), want christmas variant active on Dec 25", src, ok)
	}

	outOfRange := time.Date(2026, 6, 1, 10, 0, 0, 0, time.Local)
	if _, ok := ResolveSource(cat, outOfRange, nil, conditions, ""); ok {
		t.Error("expected ineligible outside the date range")
	}
}

func TestResolveSourceDateRangeCrossesYear(t *testing.T) {
	conditions := map[string]models.Condition{
		"summer": {DateRange: &models.DateRange{Start: "12-21", End: "03-20"}},
	}
	cat := &models.Categories{Variants: []models.Variant{
		{Source: `C:\walls\summer`, Condition: "summer"},
	}}

	newYearsEve := time.Date(2026, 12, 31, 23, 0, 0, 0, time.Local)
	if _, ok := ResolveSource(cat, newYearsEve, nil, conditions, ""); !ok {
		t.Error("expected summer to hold on Dec 31")
	}

	earlyJan := time.Date(2026, 1, 15, 10, 0, 0, 0, time.Local)
	if _, ok := ResolveSource(cat, earlyJan, nil, conditions, ""); !ok {
		t.Error("expected summer to hold on Jan 15 (wrapped from the previous year)")
	}

	midYear := time.Date(2026, 7, 1, 10, 0, 0, 0, time.Local)
	if _, ok := ResolveSource(cat, midYear, nil, conditions, ""); ok {
		t.Error("expected summer to not hold on Jul 1")
	}
}

func TestResolveSourceTemperatureOnly(t *testing.T) {
	conditions := map[string]models.Condition{
		"cold": {TemperatureMax: floatPtr(15), Priority: 10},
	}
	cat := &models.Categories{Variants: []models.Variant{
		{Source: `C:\walls\cold`, Condition: "cold"},
	}}

	warm := &weather.Snapshot{Temperature: 20}
	if _, ok := ResolveSource(cat, time.Now(), warm, conditions, ""); ok {
		t.Error("expected cold condition to not hold at 20C with max 15")
	}

	chilly := &weather.Snapshot{Temperature: 10}
	if src, ok := ResolveSource(cat, time.Now(), chilly, conditions, ""); !ok || src != `C:\walls\cold` {
		t.Errorf("got (%q, %v), want cold variant to hold at 10C", src, ok)
	}
}

func TestResolveSourceTemperatureCombinedWithSkyAndWind(t *testing.T) {
	conditions := map[string]models.Condition{
		"perfect-storm": {
			Weather:        []string{"thunderstorm"},
			WindSpeedMin:   floatPtr(40),
			TemperatureMax: floatPtr(20),
			Priority:       25,
		},
	}
	cat := &models.Categories{Variants: []models.Variant{
		{Source: `C:\walls\perfect-storm`, Condition: "perfect-storm"},
	}}

	all := &weather.Snapshot{Code: 95, WindSpeed: 45, Temperature: 18}
	if src, ok := ResolveSource(cat, time.Now(), all, conditions, ""); !ok || src != `C:\walls\perfect-storm` {
		t.Errorf("got (%q, %v), want perfect-storm to hold when all three sub-fields match", src, ok)
	}

	tooWarm := &weather.Snapshot{Code: 95, WindSpeed: 45, Temperature: 25}
	if _, ok := ResolveSource(cat, time.Now(), tooWarm, conditions, ""); ok {
		t.Error("expected perfect-storm to not hold when temperature fails its own sub-check")
	}
}

func TestResolveSourcePriorityChainWeatherCombinations(t *testing.T) {
	conditions := map[string]models.Condition{
		"mild":          {TemperatureMin: floatPtr(18), TemperatureMax: floatPtr(26), Priority: 8},
		"stormy":        {Weather: []string{"thunderstorm"}, Priority: 15},
		"stormy-windy":  {Weather: []string{"thunderstorm"}, WindSpeedMin: floatPtr(40), Priority: 20},
		"perfect-storm": {Weather: []string{"thunderstorm"}, WindSpeedMin: floatPtr(40), TemperatureMax: floatPtr(20), Priority: 25},
	}
	cat := &models.Categories{Variants: []models.Variant{
		{Source: `C:\walls\mild`, Condition: "mild"},
		{Source: `C:\walls\stormy`, Condition: "stormy"},
		{Source: `C:\walls\stormy-windy`, Condition: "stormy-windy"},
		{Source: `C:\walls\perfect-storm`, Condition: "perfect-storm"},
	}}

	ws := &weather.Snapshot{Code: 95, WindSpeed: 45, Temperature: 19}
	if src, ok := ResolveSource(cat, time.Now(), ws, conditions, ""); !ok || src != `C:\walls\perfect-storm` {
		t.Errorf("got (%q, %v), want perfect-storm (priority 25) to win", src, ok)
	}

	ws2 := &weather.Snapshot{Code: 95, WindSpeed: 45, Temperature: 25}
	if src, ok := ResolveSource(cat, time.Now(), ws2, conditions, ""); !ok || src != `C:\walls\stormy-windy` {
		t.Errorf("got (%q, %v), want stormy-windy (priority 20) to win", src, ok)
	}

	ws3 := &weather.Snapshot{Code: 95, WindSpeed: 5, Temperature: 25}
	if src, ok := ResolveSource(cat, time.Now(), ws3, conditions, ""); !ok || src != `C:\walls\stormy` {
		t.Errorf("got (%q, %v), want stormy (priority 15) to win", src, ok)
	}
}
