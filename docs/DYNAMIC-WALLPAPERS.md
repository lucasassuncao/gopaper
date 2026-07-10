# Dynamic Wallpapers: Variants and Conditions

A category can switch between multiple source directories automatically — by time of
day, by calendar date, or by live weather — instead of always drawing from one fixed
`source`. This is what powers macOS-style "dynamic wallpaper" packs (day/night renditions
of the same scene) as well as seasonal or weather-reactive collections.

## The basics: `variants`

Replace a category's single `source` with a `variants` list. Each variant provides its own
`source` and a way to decide when it's active:

```yaml
categories:
  - name: "Saltern Study"
    mode: "crop"
    enabled: true
    variants:
      - source: "C:\\Walls\\Saltern Study"
        hours: "06:00-17:59"
      - source: "C:\\Walls\\Saltern Study Night"
        hours: "18:00-05:59"
```

At selection time, gopaper resolves which variant is currently active and draws an image
from *that* directory. If none of a category's variants are active right now, the whole
category is skipped for that run (logged, not an error — unless it empties the candidate
pool entirely, which is the same "enabled categories not found" case as an all-disabled
config).

`hours` uses a 24h `"HH:MM-HH:MM"` window, both ends inclusive, and wraps past midnight —
`"18:00-05:59"` above means 6pm through 5:59am. `mode`, `enabled`, and `filter` stay at the
category level and apply no matter which variant is picked.

## Relative paths and a shared base directory

If your variants live as sibling subfolders, give the category a `source` to act as the
base directory and use relative variant paths:

```yaml
categories:
  - name: "Saltern Study"
    source: "C:\\Walls\\DynamicWallpapers\\Saltern Study"   # base directory
    mode: "crop"
    enabled: true
    variants:
      - { source: "./day", hours: "06:00-17:59" }
      - { source: "./night", hours: "18:00-05:59" }
```

A variant `source` is used as-is when it's absolute (as in the first example); when it's
relative, the category's `source` is required and the two are joined. This means existing
configs with absolute variant paths and no category-level `source` keep working unchanged
— relative paths are an added convenience, not a replacement.

## Named conditions

Instead of repeating `hours: "06:00-17:59"` in every category that wants a "daytime"
variant, declare it once under `configuration.conditions` and reference it by name:

```yaml
configuration:
  conditions:
    day:   { hours: "06:00-17:59" }
    night: { hours: "18:00-05:59" }

categories:
  - name: "Saltern Study"
    source: "C:\\Walls\\DynamicWallpapers\\Saltern Study"
    variants:
      - { source: "./day", condition: day }
      - { source: "./night", condition: night }

  - name: "Another Pack"
    source: "C:\\Walls\\DynamicWallpapers\\Another Pack"
    variants:
      - { source: "./day", condition: day }      # reuses the same condition
      - { source: "./night", condition: night }
```

A variant uses exactly one of `hours` (inline, self-contained) or `condition` (a name
looked up in `configuration.conditions`) — never both.

## Calendar-based conditions: `date-range`

A condition can hold based on the calendar instead of the clock, for seasons, holidays, or
any custom date span:

```yaml
configuration:
  conditions:
    summer:    { date-range: { start: "12-21", end: "03-20" } }   # wraps past New Year
    winter:    { date-range: { start: "06-21", end: "09-22" } }
    christmas: { date-range: { start: "12-24", end: "12-26" }, priority: 5 }

categories:
  - name: "Seasonal Pack"
    source: "C:\\Walls\\DynamicWallpapers\\Seasonal"
    variants:
      - { source: "./summer", condition: summer }
      - { source: "./winter", condition: winter }

  - name: "Holidays"
    source: "C:\\Walls\\DynamicWallpapers\\Holidays"
    variants:
      - { source: "./christmas", condition: christmas }
```

`start`/`end` are `"MM-DD"` (zero-padded), both inclusive. Same wraparound idea as `hours`
crossing midnight: since `12` (December) numerically comes after `03` (March), the
`summer` example above is understood to span December 21 through the end of the year, then
January 1 through March 20 — not the (empty) range strictly between those two dates within
one calendar year. Order `start`/`end` the other way (`03-21`..`12-20`) for a span that
doesn't cross the year boundary.

## Weather-based conditions

Conditions can react to live weather via [Open-Meteo](https://open-meteo.com/) (no API key
needed). First, tell gopaper where to check the weather:

```yaml
configuration:
  weather:
    provider: open-meteo
    latitude: -23.55     # decimal degrees, required
    longitude: -46.63    # decimal degrees, required
    cache-ttl: 15m       # optional, default 15m
```

Then a condition can use any combination of three sub-fields, which combine with **AND**:

```yaml
configuration:
  conditions:
    rainy:         { weather: [rain, drizzle], priority: 10 }
    windy:         { wind-speed-min: 30, priority: 10 }              # km/h
    cold:          { temperature-max: 15, priority: 10 }              # Celsius
    hot:           { temperature-min: 30, priority: 10 }
    mild:          { temperature-min: 18, temperature-max: 26, priority: 8 }
    stormy-windy:  { weather: [thunderstorm], wind-speed-min: 40, priority: 20 }
    hot-and-clear: { weather: [clear], temperature-min: 30, priority: 12 }
```

`weather` accepts one or more of: `clear`, `cloudy`, `fog`, `drizzle`, `rain`, `snow`,
`thunderstorm`. `wind-speed-min`/`wind-speed-max` and `temperature-min`/`temperature-max`
each accept one or both bounds to form a threshold or a range.

**A condition is exactly one of three groups — `hours`, `date-range`, or the weather
bucket above — never mixed.** `gopaper validate` rejects a condition that combines groups
(e.g. `hours` with `weather`), or one with none of the three set.

### Weather is best-effort

If the Open-Meteo request fails and no cached reading is available, weather-based
conditions simply don't hold for that run — gopaper never aborts a wallpaper change
because of a network or API problem. A successful fetch is cached (`cache-ttl`, default
15 minutes) so gopaper doesn't hit the API on every single run.

## Priority: resolving ties

At any given moment, more than one variant's condition can be true at once — e.g. it's
currently both "afternoon" (by hours) and "raining" (by weather). Each condition declares
an optional `priority` (default `0`); among all currently-holding variants in a category,
the one with the **highest priority** wins. Ties (equal priority) fall back to whichever
variant is listed first.

This is how weather naturally overrides a time-of-day default — give weather conditions a
higher priority than your `hours`/`date-range` ones:

```yaml
configuration:
  conditions:
    afternoon:     { hours: "12:00-17:59" }                                          # priority 0
    stormy:        { weather: [thunderstorm], priority: 15 }
    stormy-windy:  { weather: [thunderstorm], wind-speed-min: 40, priority: 20 }
    perfect-storm: { weather: [thunderstorm], wind-speed-min: 40, temperature-max: 20, priority: 25 }

categories:
  - name: "Weather Reactive"
    source: "C:\\Walls\\DynamicWallpapers\\Weather"
    variants:
      - { source: "./afternoon", condition: afternoon }
      - { source: "./stormy", condition: stormy }
      - { source: "./stormy-windy", condition: stormy-windy }
      - { source: "./perfect-storm", condition: perfect-storm }
```

During a thunderstorm with 45 km/h winds and 19°C: `afternoon` (0), `stormy` (15),
`stormy-windy` (20), and `perfect-storm` (25) can all be true simultaneously — the highest
priority, `perfect-storm`, wins. Drop the wind below 40 km/h and only `afternoon` and
`stormy` remain true — `stormy` (15) wins over `afternoon` (0). No storm at all, and only
`afternoon` holds.

## Full example

```yaml
configuration:
  logging:
    output: console
    level: info
  transition: fade

  weather:
    provider: open-meteo
    latitude: -23.55
    longitude: -46.63
    cache-ttl: 15m

  conditions:
    morning:   { hours: "06:00-11:59" }
    afternoon: { hours: "12:00-17:59" }
    evening:   { hours: "18:00-21:59" }
    night:     { hours: "22:00-05:59" }

    summer:    { date-range: { start: "12-21", end: "03-20" } }
    winter:    { date-range: { start: "06-21", end: "09-22" } }
    christmas: { date-range: { start: "12-24", end: "12-26" }, priority: 5 }

    rainy:         { weather: [rain, drizzle], priority: 10 }
    stormy:        { weather: [thunderstorm], priority: 15 }
    stormy-windy:  { weather: [thunderstorm], wind-speed-min: 40, priority: 20 }
    perfect-storm: { weather: [thunderstorm], wind-speed-min: 40, temperature-max: 20, priority: 25 }

categories:
  - name: "Custom Selection"                # a plain category needs none of this
    source: "C:\\Walls\\CustomSelection"
    mode: crop
    enabled: true

  - name: "Saltern Study"                   # day/night, by hours
    source: "C:\\Walls\\DynamicWallpapers\\Saltern Study"
    mode: crop
    enabled: true
    variants:
      - { source: "./day", condition: morning }
      - { source: "./night", condition: night }

  - name: "Seasonal Pack"                   # summer/winter, by date
    source: "C:\\Walls\\DynamicWallpapers\\Seasonal"
    mode: crop
    enabled: true
    variants:
      - { source: "./summer", condition: summer }
      - { source: "./winter", condition: winter }

  - name: "Holidays"                        # fixed date, high priority
    source: "C:\\Walls\\DynamicWallpapers\\Holidays"
    mode: crop
    enabled: true
    variants:
      - { source: "./christmas", condition: christmas }

  - name: "Weather Reactive"                # live weather, priority chain
    source: "C:\\Walls\\DynamicWallpapers\\Weather"
    mode: crop
    enabled: true
    variants:
      - { source: "./default", condition: afternoon }
      - { source: "./rainy", condition: rainy }
      - { source: "./stormy", condition: stormy }
      - { source: "./stormy-windy", condition: stormy-windy }
      - { source: "./perfect-storm", condition: perfect-storm }
```

## Validation

`gopaper validate` (and `gopaper edit`) check:

- Every category has `source`, `variants`, or both — never neither.
- A relative variant `source` requires the category to define `source`.
- Every variant has exactly one of `hours` or `condition`; a `condition` name must exist in
  `configuration.conditions`.
- Every named condition has exactly one of `hours`, `date-range`, or at least one
  weather-bucket field (`weather`/`wind-speed-min`/`wind-speed-max`/`temperature-min`/`temperature-max`).
- `date-range.start`/`end` are both present and parse as real `"MM-DD"` dates.
- `weather` entries are one of the seven known categories.
- `configuration.weather` (with a valid `provider`, `latitude`, `longitude`) is present
  whenever any condition uses a weather-bucket field.
- `--strict` additionally verifies each variant's resolved directory exists on disk (after
  joining a relative `source` against the category's).

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for the exact error messages these produce.
