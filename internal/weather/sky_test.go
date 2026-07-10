package weather

import "testing"

func TestCodeToSky(t *testing.T) {
	cases := []struct {
		code int
		want Sky
		ok   bool
	}{
		{0, SkyClear, true},
		{2, SkyCloudy, true},
		{45, SkyFog, true},
		{55, SkyDrizzle, true},
		{65, SkyRain, true},
		{82, SkyRain, true},
		{75, SkySnow, true},
		{95, SkyThunderstorm, true},
		{999, "", false},
	}
	for _, c := range cases {
		got, ok := CodeToSky(c.code)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("CodeToSky(%d) = (%q, %v), want (%q, %v)", c.code, got, ok, c.want, c.ok)
		}
	}
}

func TestIsValidSky(t *testing.T) {
	for _, name := range []string{"clear", "cloudy", "fog", "drizzle", "rain", "snow", "thunderstorm"} {
		if !IsValidSky(name) {
			t.Errorf("IsValidSky(%q) = false, want true", name)
		}
	}
	if IsValidSky("sunny") {
		t.Error(`IsValidSky("sunny") = true, want false (not a recognized category name)`)
	}
}

func TestSkyNames(t *testing.T) {
	names := SkyNames()
	if len(names) != 7 {
		t.Fatalf("got %d names, want 7: %v", len(names), names)
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] >= names[i] {
			t.Errorf("SkyNames() not sorted: %v", names)
		}
	}
}
