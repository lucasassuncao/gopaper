package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConditionsAbsentReturnsEmptyMap(t *testing.T) {
	v := viper.New()
	conditions, err := LoadConditions(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conditions) != 0 {
		t.Errorf("got %d conditions, want 0", len(conditions))
	}
}

func TestLoadConditionsParsesEntries(t *testing.T) {
	v := viper.New()
	v.Set("configuration.conditions.morning.hours", "06:00-11:59")
	v.Set("configuration.conditions.rainy.weather", []string{"rain", "drizzle"})
	v.Set("configuration.conditions.rainy.priority", 10)

	conditions, err := LoadConditions(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conditions["morning"].Hours != "06:00-11:59" {
		t.Errorf("morning.hours = %q, want 06:00-11:59", conditions["morning"].Hours)
	}
	if conditions["rainy"].Priority != 10 {
		t.Errorf("rainy.priority = %d, want 10", conditions["rainy"].Priority)
	}
	if len(conditions["rainy"].Weather) != 2 {
		t.Errorf("rainy.weather = %v, want 2 entries", conditions["rainy"].Weather)
	}
}

func TestLoadWeatherConfigAbsentReturnsNil(t *testing.T) {
	v := viper.New()
	wc, err := LoadWeatherConfig(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wc != nil {
		t.Errorf("got %+v, want nil", wc)
	}
}

func TestLoadWeatherConfigParsesEntries(t *testing.T) {
	v := viper.New()
	v.Set("configuration.weather.provider", "open-meteo")
	v.Set("configuration.weather.latitude", -23.55)
	v.Set("configuration.weather.longitude", -46.63)

	wc, err := LoadWeatherConfig(v)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wc == nil {
		t.Fatal("got nil, want a WeatherConfig")
	}
	if wc.Provider != "open-meteo" || wc.Latitude != -23.55 || wc.Longitude != -46.63 {
		t.Errorf("got %+v, want provider=open-meteo latitude=-23.55 longitude=-46.63", wc)
	}
}
