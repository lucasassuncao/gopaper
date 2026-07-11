# Configuration Reference

`gopaper.yaml` has two top-level sections: `configuration` and `categories`. Run `gopaper show-docs` to browse this same reference — descriptions, defaults, and allowed values — generated live from the schema, directly in the terminal.

```yaml
configuration:
  logging:
    output: "both"          # console | log | file | both | none
    file: "C:\\logs\\gopaper.log"
    level: "info"            # trace | debug | info | warn | error | fatal
    show-caller: false
  history:
    limit: 50
    file: ""
    enabled: true
  behavior:
    transition: fade         # fade | none
    monitor: all              # all | per-monitor | monitor1, monitor2, ...
    mode: crop                # crop | tile | stretch | span | fit | center

categories:
  - name: "default"
    source: "C:\\wallpapers"
    enabled: true
    filter:                  # optional
      match:
        glob: "landscape_*"
      age:
        max: 720h
      size:
        min: "500KB"
```

Both `logging.file`, `history.file`, and every category's `source` accept a leading `~` (or `~\` on Windows), expanded to the user's home directory at load time.

---

## `configuration.logging`

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `output` | string | yes | `console` | One of `console`, `log`, `file`, `both`, `none`. |
| `file` | string | required if `output` is `log`, `file`, or `both` | — | Log destination. Parent directories are created automatically. |
| `level` | string | yes | `info` | One of `trace`, `debug`, `info`, `warn`, `error`, `fatal`. |
| `show-caller` | bool | no | `false` | Includes the source file/line in each log entry. |

`output: none` discards all log output.

## `configuration.history`

Controls the wallpaper history used by `gopaper prev`/`gopaper next`.

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `limit` | int | no | `50` | Maximum number of past wallpapers kept. Oldest entries are dropped once exceeded. |
| `file` | string | no | `<executable_dir>/history/gopaper.json` | Custom history file path. |
| `enabled` | bool | no | `true` | Set `false` to stop recording history entirely (`prev`/`next` then have nothing to navigate). |

## `configuration.behavior` and `categories[].behavior`

`behavior` groups how a wallpaper change is applied. The `configuration`-level block sets
the run defaults; any category can declare its own `behavior` block, whose non-empty fields
override the defaults **when that category wins the draw**.

```yaml
configuration:
  behavior:
    transition: fade         # fade (default) | none
    monitor: all              # all (default) | per-monitor | monitor1, monitor2, ...
    mode: crop                # crop (default) | tile | stretch | span | fit | center

categories:
  - name: "Heavy RAWs"
    source: "C:\\Walls\\RAW"
    behavior:
      transition: none       # this category always swaps instantly
      mode: fit               # this category always uses fit instead of the default
```

### `behavior.transition`

| Value | Effect |
|---|---|
| `fade` (default) | Native Windows crossfade between the old and new wallpaper. Falls back to an instant change on non-Windows platforms, or if the fade path fails for any reason. |
| `none` | Instant change, no transition — the pre-fade behavior. |

### `behavior.mode`

Wallpaper display mode — see [Wallpaper modes](#wallpaper-modes) for the six values. Defaults
to `crop` when unset at both configuration and category level. A category's own
`behavior.mode` overrides the configuration-level default when that category wins the draw.

### `behavior.monitor`

| Value | Effect |
|---|---|
| `all` (default) | One image mirrored on every monitor — the classic behavior; `fade` works. |
| `per-monitor` | Each monitor gets its own category draw and image — always instant. |
| `monitor1`, `monitor2`, ... | Pins this category to that single monitor (1-based, Windows enumeration order); every other monitor is left untouched — always instant. |

Not to be confused with the category-level `monitor` field (an int, e.g. `monitor: 1`), which
only *restricts* a category's eligibility inside a `per-monitor` draw — see below.

**Not sure which index is which screen?** Run [`gopaper monitors`](COMMANDS.md#gopaper-monitors)
— it lists every connected monitor with its 1-based index, EDID name (e.g. `ASUS VG32VQ1B`),
and desktop position/size, so you can match `monitor1`/`monitor2` (here, and in the
category-level `monitor: N` field below) to a physical screen before writing it into the
config.

**The drawn category decides the run.** gopaper first draws one category from the eligible
set; that category's effective `behavior.monitor` (its own override, or the configuration
default) picks the flow:

- Effective `all` → one image from it, mirrored on every monitor, with fade. Categories set
  to `all` never take part in individual per-monitor draws — they only ever appear mirrored.
- Effective `per-monitor` → every monitor gets its own draw among the per-monitor-eligible
  categories. A category can be further restricted to one monitor with the category-level
  `monitor: N` field (1-based, Windows enumeration order); without it, the category is
  eligible for any monitor. This still means competing with other eligible categories for
  that monitor — different from `behavior.monitor: monitorN` below.
- Effective `monitorN` → this category's own image goes straight to monitor `N`; no draw
  against other categories, and every other monitor keeps whatever it already had. If monitor
  `N` isn't connected, gopaper falls back to the normal single-wallpaper flow.

Notes and limitations of per-monitor and `monitorN` changes:

- **Always instant.** The native crossfade cannot target monitors individually (it relies on
  a one-item slideshow that forces the same image everywhere), so `behavior.transition` is
  ignored for them.
- The wallpaper **mode** (`crop`, `fit`, …) is a single global setting in Windows; for
  `per-monitor` the mode of the category chosen for monitor 1 wins, for `monitorN` the pinned
  category's own mode wins.
- On a machine with a single monitor (or if monitor enumeration fails), gopaper falls back to
  the normal single-wallpaper flow automatically.
- History records every monitor's image; `prev`/`next` and `gopaper history` reapply them by
  monitor position, skipping monitors that are no longer connected.

## `configuration.wallhaven` and `categories[].wallhaven`

A category can source its images from the [Wallhaven](https://wallhaven.cc) API instead of a
local directory. Downloads are kept in a local cache directory, which acts as the category's
source — selection, filters, and history work exactly as with a local folder, and the
category keeps working offline from its cache.

```yaml
configuration:
  wallhaven:
    api-key: "xxxxxxxx"        # optional

categories:
  - name: "Wallhaven Landscapes"
    wallhaven:
      query: "landscape"
      purity: sfw              # sfw (default) | sketchy | nsfw
      cache: "~/Pictures/Walls/.wallhaven-cache"   # optional
      cache-limit: 100         # optional, default 100
    enabled: true
```

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `configuration.wallhaven.api-key` | string | no | — | Wallhaven API key. Without it, searches are anonymous and only `sfw` purity is allowed. |
| `configuration.wallhaven.cache` | string | no | — | Base directory for every wallhaven category's cache, each in its own `<category-slug>` subdirectory. A category's own `wallhaven.cache` overrides this. |
| `wallhaven.query` | string | yes | — | Wallhaven search query. |
| `wallhaven.purity` | string | no | `sfw` | `sketchy`/`nsfw` require the API key (validated). |
| `wallhaven.cache` | string | no | `configuration.wallhaven.cache`, or `<history_dir>/wallhaven-cache/<category-slug>` if that's unset too | Directory where downloads are kept. |
| `wallhaven.cache-limit` | int | no | `100` | Oldest images are pruned beyond this count. |

`wallhaven` is mutually exclusive with `source` and `variants`. Each run fetches at most one
new image (random result for the query) before selecting; a network failure just means that
run draws from the existing cache.

## `configuration.weather` and `configuration.conditions`

Optional sections that power **dynamic wallpapers** — categories that switch source
directory by time of day, calendar date, or live weather. See
[DYNAMIC-WALLPAPERS.md](DYNAMIC-WALLPAPERS.md) for the full guide with examples; summary:

```yaml
configuration:
  weather:                              # only needed if a condition uses weather fields
    provider: open-meteo
    latitude: -23.55
    longitude: -46.63
    cache-ttl: 15m
  conditions:
    day:    { hours: "06:00-17:59" }
    summer: { date-range: { start: "12-21", end: "03-20" } }
    rainy:  { weather: [rain, drizzle], priority: 10 }
```

## `categories[]`

Each entry is one wallpaper source.

| Field | Type | Required | Notes |
|---|---|---|---|
| `name` | string | yes, unique | Display name; must not repeat across categories. |
| `source` | string | yes, unless `variants` or `wallhaven` is set | Directory scanned for images (`.jpg`, `.jpeg`, `.png`, `.webp`), not recursive. With `variants`, doubles as the base directory for any relative variant `source`. |
| `variants` | list | no | Time/date/weather-conditioned renditions of this category — see [DYNAMIC-WALLPAPERS.md](DYNAMIC-WALLPAPERS.md). |
| `wallhaven` | object | no | Sources this category from the Wallhaven API — see [`configuration.wallhaven`](#configurationwallhaven-and-categorieswallhaven). Mutually exclusive with `source`/`variants`. |
| `enabled` | bool | no (default `true`) | Disabled categories are skipped unless selected explicitly with `--category --include-disabled`. |
| `behavior` | object | no | Overrides `configuration.behavior` (`transition`, `monitor`, `mode`) when this category wins the draw. |
| `monitor` | int | no | Restricts this category to one monitor (1-based) within `behavior.monitor: per-monitor` draws; ignored otherwise. Different from `behavior.monitor: monitorN`, which pins the category itself — see [`behavior.monitor`](#behaviormonitor). |
| `filter` | object | no | Narrows which files in `source` are eligible — see [FILTERS.md](FILTERS.md). |

A category with `variants` but no `variant` currently active (e.g. outside every `hours`
window) is skipped for that run, same as a disabled category — logged, not an error.

## Wallpaper modes

| Mode | Effect |
|---|---|
| `crop` | Fills the screen, cropping edges that don't fit the aspect ratio. |
| `tile` | Repeats the image at its native size. |
| `stretch` | Stretches to fill the screen, ignoring aspect ratio. |
| `span` | Stretches across all connected monitors. |
| `fit` | Scales to fit within the screen without cropping. |
| `center` | Centers the image at its native size, no scaling. |

## Validating a config file

```pwsh
gopaper validate                 # pretty output, standard lookup
gopaper validate -c ./gopaper.yaml -f json
gopaper validate --strict         # also check that every category's source exists on disk
```

See [COMMANDS.md](COMMANDS.md#gopaper-validate) for the full flag reference.
