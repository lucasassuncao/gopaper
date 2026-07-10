// Package weather provides current weather conditions (via Open-Meteo) used
// by category variants that switch on sky condition or wind speed.
package weather

import "sort"

// Sky is one of the weather categories a condition can match against.
type Sky string

const (
	SkyClear        Sky = "clear"
	SkyCloudy       Sky = "cloudy"
	SkyFog          Sky = "fog"
	SkyDrizzle      Sky = "drizzle"
	SkyRain         Sky = "rain"
	SkySnow         Sky = "snow"
	SkyThunderstorm Sky = "thunderstorm"
)

// codeToSky maps Open-Meteo's WMO weather codes to a Sky category.
var codeToSky = map[int]Sky{
	0:  SkyClear,
	1:  SkyCloudy,
	2:  SkyCloudy,
	3:  SkyCloudy,
	45: SkyFog,
	48: SkyFog,
	51: SkyDrizzle,
	53: SkyDrizzle,
	55: SkyDrizzle,
	56: SkyDrizzle,
	57: SkyDrizzle,
	61: SkyRain,
	63: SkyRain,
	65: SkyRain,
	66: SkyRain,
	67: SkyRain,
	80: SkyRain,
	81: SkyRain,
	82: SkyRain,
	71: SkySnow,
	73: SkySnow,
	75: SkySnow,
	77: SkySnow,
	85: SkySnow,
	86: SkySnow,
	95: SkyThunderstorm,
	96: SkyThunderstorm,
	99: SkyThunderstorm,
}

var validSkyNames = map[string]bool{
	string(SkyClear):        true,
	string(SkyCloudy):       true,
	string(SkyFog):          true,
	string(SkyDrizzle):      true,
	string(SkyRain):         true,
	string(SkySnow):         true,
	string(SkyThunderstorm): true,
}

// IsValidSky reports whether name is one of the known sky category names.
func IsValidSky(name string) bool {
	return validSkyNames[name]
}

// SkyNames returns all known sky category names, sorted.
func SkyNames() []string {
	names := make([]string, 0, len(validSkyNames))
	for n := range validSkyNames {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// CodeToSky maps a WMO weather code to a Sky category. ok is false for an
// unrecognized code.
func CodeToSky(code int) (sky Sky, ok bool) {
	sky, ok = codeToSky[code]
	return
}
