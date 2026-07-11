# Gopaper

Gopaper is a small, cross-platform command-line tool written in Go that selects and sets a desktop wallpaper from user-defined categories. Each category points to a directory of images and controls how images are chosen and displayed.

**Features**
- Randomly selects a wallpaper from enabled categories
- Supports multiple wallpaper display modes (crop, tile, stretch, span, fit, center)
- Smooth native crossfade transition on Windows (`behavior.transition: fade`)
- Dynamic wallpapers: categories that switch source by time of day, calendar date, or live weather (via [Open-Meteo](https://open-meteo.com/))
- Configurable logging and output
- Generates a template `gopaper.yaml` configuration file

**Note:** This README focuses on usage and configuration. For developer notes and code-level docs, check the source files under `internal/`.

## Documentation

- [Getting Started](docs/GETTING-STARTED.md) — install, first config, first run
- [Configuration Reference](docs/CONFIGURATION.md) — full `gopaper.yaml` schema
- [Dynamic Wallpapers](docs/DYNAMIC-WALLPAPERS.md) — variants, time/date/weather conditions
- [Filters](docs/FILTERS.md) — narrowing which files a category picks from
- [Commands](docs/COMMANDS.md) — every command and flag
- [Interactive Config Editor](docs/EDIT.md) — `gopaper edit` walkthrough
- [Troubleshooting](docs/TROUBLESHOOTING.md) — common errors and fixes

Contributing? See [CONTRIBUTING.md](CONTRIBUTING.md).

## Installation

- Build from source (requires Go 1.20+):

```pwsh
go build -o gopaper ./...
```

- Or use the provided `Makefile` (if available):

```pwsh
make build
```

Place the produced binary in a folder that is in your `PATH`, or run it from the project directory.

## Configuration

Gopaper reads a YAML configuration file named `gopaper.yaml` (or another filename you specify with `-c/--config`). The config is composed of two main sections: `configuration` and `categories`. `configuration` is further split into `logging` and `history` sub-sections.

Example `gopaper.yaml`:

```yaml
configuration:
  logging:
    output: "both"        # one of: console, log, file, both, none
    file: "C:\\logs\\gopaper.log"
    level: "info"         # trace, debug, info, warn, error, fatal
    show-caller: false
  history:
    limit: 50             # max wallpapers kept for prev/next (default: 50)
    file: ""              # defaults to <executable_dir>/history/gopaper.json
    enabled: true          # set to false to disable prev/next history recording

categories:
  - name: "default"
    source: "C:\\wallpapers"
    enabled: true
  - name: "nature"
    source: "D:/Images/Nature"
    behavior:
      mode: "fit"        # overrides the crop default just for this category
    enabled: true
    filter:              # optional: narrow which files in source are eligible
      match:
        glob: "landscape_*"   # literal | regex | glob (mutually exclusive)
      age:
        max: 720h              # only files modified in the last 30 days
      size:
        min: "500KB"
```

### Filtering files within a category

Each category may set an optional `filter` to narrow eligible files beyond the fixed image-extension check (`.jpg`, `.jpeg`, `.png`, `.webp`):

- `filter.match` — `literal` (exact name), `regex` (RE2), or `glob` (wildcard), mutually exclusive; add `case-sensitive: true` to stop lowercasing before comparison.
- `filter.age` — `min`/`max` time since the file was last modified (e.g. `24h`, `720h`).
- `filter.size` — `min`/`max` file size (e.g. `"500KB"`, `"2MiB"`).

All three combine with AND semantics. `gopaper edit` and `gopaper validate` both check that regexes/globs compile and that size/age bounds are ordered.

You can generate a config file using the built-in generator (this creates `gopaper.yaml` at `<executable_dir>/conf/gopaper.yaml`):

```pwsh
gopaper init -t full   # or -i for interactive prompts
```

If you prefer to create the file manually, use the example above as a template.

### Dynamic wallpapers (time, date, and weather)

A category isn't limited to one fixed `source` — it can switch between multiple
directories automatically, based on the time of day, the calendar date, or live weather:

```yaml
configuration:
  behavior:
    transition: fade
  weather:
    provider: open-meteo
    latitude: -23.55
    longitude: -46.63
  conditions:
    day:   { hours: "06:00-17:59" }
    night: { hours: "18:00-05:59" }
    rainy: { weather: [rain, drizzle], priority: 10 }

categories:
  - name: "Saltern Study"
    source: "C:\\Walls\\DynamicWallpapers\\Saltern Study"
    enabled: true
    variants:
      - { source: "./day", condition: day }
      - { source: "./night", condition: night }
```

See [docs/DYNAMIC-WALLPAPERS.md](docs/DYNAMIC-WALLPAPERS.md) for the full guide: relative
variant paths, named conditions, calendar date ranges (including ones that span New
Year's Eve), weather thresholds (sky, wind, temperature), and how priority resolves ties
when more than one variant is active at once.

### Wallhaven source

Instead of a local directory, a category can pull its images from the
[Wallhaven](https://wallhaven.cc) API — downloads are cached locally, so the category keeps
working offline:

```yaml
configuration:
  wallhaven:
    api-key: "xxxxxxxx"   # optional; required only for sketchy/nsfw purity

categories:
  - name: "Wallhaven Landscapes"
    wallhaven:
      query: "landscape"
      purity: sfw          # sfw (default) | sketchy | nsfw
      cache-limit: 100
    enabled: true
```

### Behavior (transition, mode & monitor)

`configuration.behavior` sets how changes are applied; any category can override it with
its own `behavior` block, which wins when that category is drawn:

```yaml
configuration:
  behavior:
    transition: fade           # fade | none
    mode: crop                 # crop | tile | stretch | span | fit | center
    monitor: per-monitor       # all | per-monitor | monitor1, monitor2, ...

categories:
  - name: "Saltern Study"      # when drawn, mirrors on all monitors WITH fade
    behavior:
      monitor: all
    # ...
  - name: "Custom Selection"   # participates in per-monitor draws (instant)
    monitor: 1                 # optional: restricts it to monitor 1 within that draw
    # ...
  - name: "Desk Monitor Only"  # always goes straight to monitor 2, others untouched
    behavior:
      monitor: monitor2
    # ...
```

**The drawn category decides the run:** effective `all` mirrors one image everywhere (fade
works); effective `per-monitor` gives each monitor its own draw among per-monitor-eligible
categories; effective `monitorN` sends this category's own image straight to that monitor,
leaving the rest as they were. `per-monitor` and `monitorN` are always instant, since the
native crossfade can't target monitors individually. See
[docs/CONFIGURATION.md](docs/CONFIGURATION.md) for details and limitations.

Not sure which index is which physical screen? Run `gopaper monitors` — it lists each
connected monitor's 1-based index, EDID name, and desktop position/size, so you can pick the
right `monitorN` before writing it into the config. See
[docs/COMMANDS.md](docs/COMMANDS.md#gopaper-monitors).

## Editing the Configuration

`gopaper edit` opens an interactive two-panel TUI editor for your configuration file, with inline hints, presets, and validation on save:

```pwsh
gopaper edit
```

See [docs/EDIT.md](docs/EDIT.md) for the full walkthrough (keybindings, presets, themes, and flags).

Run `gopaper show-docs` to browse the same field reference (descriptions, defaults, allowed values) directly in the terminal, without opening the editor:

```pwsh
gopaper show-docs
gopaper show-docs --section history
```

## Validating the Configuration

`gopaper validate` runs the same checks as `gopaper edit` (required fields, allowed values, uniqueness, cross-field rules) without opening the TUI — useful in scripts or CI:

```pwsh
gopaper validate                      # pretty output, default lookup
gopaper validate -c ./gopaper.yaml -f json
gopaper validate --strict             # also verify source directories exist on disk
```

`--format` accepts `pretty` (default), `plain`, or `json`; `--summary` prints only the error count. The command exits non-zero when validation fails.

## Usage

Basic run (uses default config lookup paths):

```pwsh
gopaper
```

Specify a config file explicitly:

```pwsh
gopaper -c C:\path\to\gopaper.yaml
```

Restrict selection to specific categories (comma-separated), regardless of the random pick:

```pwsh
gopaper --category "Wallhaven,Nature"
gopaper --category "Wallhaven" --include-disabled   # allow a disabled category too
```

The tool will:
- Load configuration and categories
- Select a random category from the eligible set (enabled categories, or the ones named with `--category`)
- Read files from the category source directory, applying its `filter` if set
- Pick a random file and set it as the desktop wallpaper
- Log the selected category, new wallpaper path, previous wallpaper (if available), and the mode used

## Wallpaper Modes

Supported wallpaper modes (these map to what the system API expects):
- `crop`
- `tile`
- `stretch`
- `span`
- `fit`
- `center`

Choose the mode that best suits your screen and image aspect ratio.

## Logging & Troubleshooting

The CLI logs contextual errors and returns wrapped errors to the caller — check the log file (if `configuration.logging.output` is `log`, `file`, or `both`) for details. For specific error messages and fixes (missing config, empty categories, filters excluding everything, permission errors, self-update checksum mismatches), see [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md).

## Development

- Project layout highlights:
  - `internal/cmd` — CLI commands and command wiring
  - `internal/config` — configuration helpers and viper integration
  - `internal/helper` — wallpaper manipulation and utility functions
  - `internal/filters` — category file-filter compilation and matching
  - `internal/history` — wallpaper history persistence (prev/next)
  - `internal/models` — data structures and interactive config generation

Run linter/tests locally:

```pwsh
go vet ./...
go test ./...
```

## Examples

- Create a config interactively:

```pwsh
gopaper init -i
```

- Edit the config in the TUI editor:

```pwsh
gopaper edit
```

- Run using a custom config file:

```pwsh
gopaper -c C:\Users\lucas\configs\gopaper.yaml
```

- Browse the history and reapply any past wallpaper:

```pwsh
gopaper history
gopaper history --category "Saltern Study"
```

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for the dev setup, Makefile targets, and what to run before opening a pull request.

## License

This project does not include a license file in the repository. If you want to publish it, add an appropriate `LICENSE` file.