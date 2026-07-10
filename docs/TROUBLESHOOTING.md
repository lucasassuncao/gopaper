# Troubleshooting

## "configuration file not found"

`gopaper` looks for `gopaper.yaml` next to the executable, then in its `conf` subdirectory, unless you pass `-c`/`--config`. If neither exists, run:

```pwsh
gopaper init -i
```

or point explicitly at a file:

```pwsh
gopaper -c C:\path\to\gopaper.yaml
```

Check the exact search order and locations in [GETTING-STARTED.md](GETTING-STARTED.md#where-gopaper-looks-for-the-config-file).

## "enabled categories not found"

Every category in `gopaper.yaml` has `enabled: false`, or there are no categories at all. Enable at least one, or restrict to a specific one:

```pwsh
gopaper --category "MyCategory" --include-disabled
```

Run `gopaper validate` first — it catches an empty `categories` list before this ever comes up at runtime.

## "unknown category \"X\""

`--category` was given a name that doesn't match any `categories[].name` in the loaded config exactly (case-sensitive). Run without `--category` to confirm the real names, or check for a typo.

## "no supported image files found... matching the configured filter"

The category's `source` directory has no files with a supported extension (`.jpg`, `.jpeg`, `.png`, `.webp`), or its `filter` excludes everything present. Narrow down which:

```pwsh
gopaper validate --strict     # confirms source exists and is reachable
```

then check the filter against what's actually in the directory — see [FILTERS.md](FILTERS.md). A `filter.age`/`filter.size` bound that's too strict, or a `match.glob`/`match.regex` typo, are the usual causes.

## Log file isn't being written

`configuration.logging.output` must be `log`, `file`, or `both` — `console` (the default) never touches disk. When it is one of those, `configuration.logging.file` is required; `gopaper validate` reports this before you hit it at runtime. The parent directory is created automatically, but the process still needs write permission there.

## `~` in a path isn't resolving

A leading `~` or `~/` (also `~\` on Windows) in `configuration.logging.file`, `configuration.history.file`, or a category's `source` expands to the user's home directory automatically. Anything else — a bare `~username`, or `~` in the middle of a path — is left untouched. If a path still looks unexpanded, check for a typo in the prefix (it must be exactly `~` or start with `~/`/`~\`).

## A category with `variants` never gets picked ("no variant active for the current time")

None of that category's variants currently match — e.g. every `hours`/`date-range` window
excludes right now, or a weather-bucket condition's thresholds aren't met (or weather data
isn't available at all). This is logged at info level and is not an error by itself; it
only becomes the `"enabled categories not found"` error if it empties the entire candidate
pool. Run `gopaper validate` to confirm the condition definitions themselves are correct,
and see [DYNAMIC-WALLPAPERS.md](DYNAMIC-WALLPAPERS.md) for how conditions and priority are
resolved.

## "hours and condition are mutually exclusive" / "hours, date-range, and weather/... are mutually exclusive"

A variant (or a named condition) set more than one of the mutually-exclusive groups
described in [DYNAMIC-WALLPAPERS.md](DYNAMIC-WALLPAPERS.md#named-conditions) — pick exactly
one: `hours`, `date-range`, or the weather bucket (`weather`/`wind-speed-*`/`temperature-*`,
which combine with each other via AND, just not with the other two groups).

## "relative source requires the category to define source"

A variant's `source` is relative (e.g. `"./day"`) but its category has no `source` of its
own to resolve it against. Either add a `source` to the category (used as the base
directory) or make the variant's `source` absolute.

## "configuration.weather required because a condition uses weather/..."

Some entry in `configuration.conditions` uses `weather`, `wind-speed-min`/`max`, or
`temperature-min`/`max`, but `configuration.weather` (`provider`, `latitude`, `longitude`)
isn't set. Add it — see [DYNAMIC-WALLPAPERS.md](DYNAMIC-WALLPAPERS.md#weather-based-conditions).

## Weather-based variants never seem to activate

Check, in order: `configuration.weather.latitude`/`longitude` are correct for your actual
location; the condition's thresholds are realistic for current conditions (a
`temperature-min: 30` condition won't hold in winter); and whether the weather fetch is
failing outright — a failed fetch with no usable cache makes every weather-bucket
condition evaluate to "not holding" rather than erroring, so it can look identical to
"the weather just doesn't match" from the outside. There's no dedicated status command
yet to distinguish the two; check `configuration.logging` output for a fetch warning.

## `prev`/`next` say history is empty

Either no wallpaper has been set yet with the current history file, or `configuration.history.enabled: false` is set (which disables recording entirely — there's nothing to navigate). Check `configuration.history.file` if you've customized it; `prev`/`next` and the main run must agree on the same file to see the same history.

## `self-update` reports a checksum mismatch

The downloaded binary didn't match the checksum manifest published with the release — the update is aborted rather than installed. Retry (transient download corruption is the most common cause); if it persists, download the release manually from GitHub and compare checksums yourself before reporting it.

## Permission errors

Common on Windows binaries placed in `Program Files`, or on read-only mounts. Run from a directory you have write access to (for the log file, history file, and — during `self-update` — the binary's own directory), or elevate as needed.

## Still stuck?

- [`gopaper validate`](COMMANDS.md#gopaper-validate) catches most config mistakes with a specific field path and message.
- [`gopaper show-docs`](COMMANDS.md#gopaper-show-docs) prints the full schema reference, including allowed values and defaults, without leaving the terminal.
- Check the log file (if `configuration.logging.output` isn't `console`) for the exact error gopaper hit.
